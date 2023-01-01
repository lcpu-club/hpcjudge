package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	bridgeApi "github.com/lcpu-club/hpcjudge/bridge/api"
	"github.com/lcpu-club/hpcjudge/common/consts"
	"github.com/lcpu-club/hpcjudge/common/models"
	"github.com/lcpu-club/hpcjudge/discovery"
	discoveryProtocol "github.com/lcpu-club/hpcjudge/discovery/protocol"
	"github.com/lcpu-club/hpcjudge/judge/configure"
	"github.com/lcpu-club/hpcjudge/judge/message"
	"github.com/lcpu-club/hpcjudge/judge/problem"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nsqio/go-nsq"
	"github.com/satori/uuid"
)

type Judger struct {
	id                      uuid.UUID
	nsqConsumer             *nsq.Consumer
	nsqReport               *nsq.Producer
	discoveryClient         *discovery.Client
	redisConn               redis.Conn
	configure               *configure.Configure
	minio                   *minio.Client
	nsqMessageTouchInterval time.Duration
}

func NewJudger(conf *configure.Configure) (*Judger, error) {
	j := new(Judger)
	j.configure = conf
	j.id = conf.ID
	if j.id == uuid.Nil {
		j.id = uuid.NewV4()
	}
	return j, nil
}

func (j *Judger) Run() error {
	err := j.connectDiscovery()
	if err != nil {
		log.Println("Connect to Discovery failed")
		return err
	}
	err = j.connectMinIO()
	if err != nil {
		log.Println("Connect to MinIO failed")
		return err
	}
	err = j.connectRedis()
	if err != nil {
		log.Println("Connect to Redis failed")
		return err
	}
	err = j.connectNSQ()
	if err != nil {
		log.Println("Connect to NSQ failed")
		return err
	}
	j.listenMinIOEvent()
	return nil
}

func (j *Judger) connectDiscovery() error {
	j.discoveryClient = discovery.NewClient(j.configure.Discovery.Address, j.configure.Discovery.AccessKey, j.configure.Discovery.Timeout)
	return nil
}

func (j *Judger) connectNSQ() error {
	config := nsq.NewConfig()
	config.AuthSecret = j.configure.Nsq.AuthSecret
	config.MaxAttempts = uint16(j.configure.Nsq.MaxAttempts) + 1
	config.MaxRequeueDelay = j.configure.Nsq.RequeueDelay
	config.MsgTimeout = j.configure.Nsq.MsgTimeout
	if j.configure.Nsq.MsgTimeout >= 3*time.Second {
		j.nsqMessageTouchInterval = j.configure.Nsq.MsgTimeout - (1 * time.Second)
	} else {
		j.nsqMessageTouchInterval = j.configure.Nsq.MsgTimeout * 2 / 3
	}
	var err error
	j.nsqConsumer, err = nsq.NewConsumer(j.configure.Nsq.Topics.Judge, j.configure.Nsq.Channel, config)
	if err != nil {
		return err
	}
	j.nsqConsumer.AddConcurrentHandlers(j, j.configure.Nsq.Concurrent)
	err = j.nsqConsumer.ConnectToNSQLookupds(j.configure.Nsq.NsqLookupd.Address)
	if err != nil {
		return err
	}
	j.nsqReport, err = nsq.NewProducer(j.configure.Nsq.Nsqd.Address, config)
	log.Println("Connected to NSQ Server")
	return err
}

func (j *Judger) connectMinIO() error {
	var err error
	j.minio, err = minio.New(j.configure.MinIO.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			j.configure.MinIO.Credentials.AccessKey, j.configure.MinIO.Credentials.SecretKey, "",
		),
		Secure: j.configure.MinIO.SSL,
	})
	if err != nil {
		return err
	}
	log.Println("Connected to MinIO Server")
	return nil
}

func (j *Judger) connectRedis() error {
	options := []redis.DialOption{}
	if j.configure.Redis.Password != "" {
		options = append(options, redis.DialPassword(j.configure.Redis.Password))
	}
	options = append(options, redis.DialKeepAlive(j.configure.Redis.KeepAlive))
	options = append(options, redis.DialDatabase(j.configure.Redis.Database))
	var err error
	j.redisConn, err = redis.Dial("tcp", j.configure.Redis.Address, options...)
	if err != nil {
		return err
	}
	log.Println("Connected to Redis Server")
	return nil
}

func (j *Judger) listenMinIOEvent() {
	chResults := j.minio.ListenBucketNotification(
		context.Background(),
		j.configure.MinIO.Buckets.Solution, "", "result.json", []string{
			"s3:ObjectCreated:*",
		},
	)
	go func() {
		for n := range chResults {
			if n.Err != nil {
				log.Println("ERROR:", n.Err)
				continue
			}
			for _, record := range n.Records {
				k := record.S3.Object.Key
				v := record.S3.Object.ETag
				id, err := j.resultObjectKeyToSolutionID(k, "result.json")
				if err != nil {
					continue
				}
				exists, err := j.checkIfRequestExists(id+v, j.configure.Redis.Expire.Report)
				if err != nil {
					log.Println("ERROR:", err)
					continue
				}
				if exists {
					continue
				}
				obj, err := j.minio.GetObject(
					context.Background(),
					j.configure.MinIO.Buckets.Solution,
					k,
					minio.GetObjectOptions{
						VersionID: record.S3.Object.VersionID,
					},
				)
				if err != nil {
					log.Println("ERROR:", err)
					err := j.setRequestNotExist(id + v)
					if err != nil {
						log.Println("ERROR:", err)
					}
					continue
				}
				rsltJSON, err := io.ReadAll(obj)
				if err != nil {
					log.Println("ERROR:", err)
					continue
				}
				r := new(models.JudgeResult)
				err = json.Unmarshal(rsltJSON, r)
				if err != nil {
					log.Println("ERROR:", err)
					continue
				}
				resp := &message.JudgeReportMessage{
					SolutionID: id,
					Success:    true,
					Done:       r.Done,
					Score:      r.Score,
					Message:    r.Message,
					Timestamp:  time.Now().UnixMicro(),
				}
				err = j.publishToReport(resp)
				if err != nil {
					log.Println("ERROR:", err)
				}
			}
		}
	}()
	chRunCommandReports := j.minio.ListenBucketNotification(
		context.Background(),
		j.configure.MinIO.Buckets.Solution, "", "run-command-report.json", []string{
			"s3:ObjectCreated:*",
		},
	)
	go func() {
		for n := range chRunCommandReports {
			if n.Err != nil {
				log.Println("ERROR:", n.Err)
				continue
			}
			for _, record := range n.Records {
				k := record.S3.Object.Key
				v := record.S3.Object.ETag
				id, err := j.resultObjectKeyToSolutionID(k, "run-command-report.json")
				if err != nil {
					continue
				}
				exists, err := j.checkIfRequestExists(id+v, j.configure.Redis.Expire.Report)
				if exists {
					continue
				}
				obj, err := j.minio.GetObject(
					context.Background(),
					j.configure.MinIO.Buckets.Solution,
					k,
					minio.GetObjectOptions{
						VersionID: record.S3.Object.VersionID,
					},
				)
				if err != nil {
					log.Println("ERROR:", err)
					err := j.setRequestNotExist(id + v)
					if err != nil {
						log.Println("ERROR:", err)
					}
					continue
				}
				rsltJSON, err := io.ReadAll(obj)
				if err != nil {
					log.Println("ERROR:", err)
					continue
				}
				r := new(bridgeApi.ExecuteCommandResponse)
				err = json.Unmarshal(rsltJSON, r)
				if err != nil {
					log.Println("ERROR:", err)
					continue
				}
				// TODO: implement report
			}
		}
	}()
}

func (j *Judger) resultObjectKeyToSolutionID(k string, suffix string) (string, error) {
	key := j.configure.Redis.Prefix + k
	id, res, found := strings.Cut(key, "/")
	if !found {
		return "", fmt.Errorf("/ not found")
	}
	if res != suffix {
		return "", fmt.Errorf("not result.json")
	}
	return id, nil
}

func (j *Judger) checkIfRequestExists(k string, expire time.Duration) (bool, error) {
	key := j.configure.Redis.Prefix + k
	rslt, err := j.redisConn.Do("INCR", key)
	if err != nil {
		return true, err
	}
	rInteger, ok := rslt.(int)
	if !ok {
		return true, fmt.Errorf("unexpected return type from redis")
	}
	if rInteger == 1 {
		_, err = j.redisConn.Do("EXPIRE", key, int(expire/time.Second))
		return false, err
	}
	return true, err
}

func (j *Judger) setRequestNotExist(key string) error {
	_, err := j.redisConn.Do("DEL", key)
	return err
}

func (j *Judger) publishToReport(msg *message.JudgeReportMessage) error {
	mText, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return j.nsqReport.Publish(j.configure.Nsq.Topics.Report, mText)
}

func (j *Judger) discoverBridge(tags []string, excludeTags []string) (*discoveryProtocol.Service, error) {
	return j.discoveryClient.Query(&discoveryProtocol.QueryParameters{
		Type:        consts.HpcBridgeDiscoveryType,
		Tags:        tags,
		ExcludeTags: excludeTags,
	})
}

func (j *Judger) ProcessJudge(msg *message.JudgeMessage) error {
	exists, err := j.checkIfRequestExists(msg.SolutionID, j.configure.Redis.Expire.Judge)
	if exists {
		return err
	}
	probMeta, err := problem.GetProblemMeta(context.Background(), j.minio, j.configure.MinIO.Buckets.Problem, msg.ProblemID)
	// TODO: implement judge
	if err != nil {
		return err
	}
	bridgeSvc, err := j.discoverBridge(probMeta.Environment.Tags, probMeta.Environment.ExcludeTags)
	if err != nil {
		return err
	}
	_ = bridgeSvc
	return nil
}

func (j *Judger) HandleMessage(msg *nsq.Message) error {
	msg.Touch()
	jMsg := &message.JudgeMessage{}
	err := json.Unmarshal(msg.Body, jMsg)
	if err != nil {
		if msg.Attempts > uint16(j.configure.Nsq.MaxAttempts) {
			msg.Finish()
			return nil
		}
		return err
	}
	if msg.Attempts > uint16(j.configure.Nsq.MaxAttempts) {
		err := j.publishToReport(&message.JudgeReportMessage{
			SolutionID: jMsg.SolutionID,
			Success:    false,
			Done:       true,
			Error:      message.ErrMaxAttemptsExceeded.Error(),
			Score:      0,
			Message:    "Internal Error: " + message.ErrMaxAttemptsExceeded.Error(),
		})
		if err != nil {
			log.Println("ERROR:", err)
		}
		msg.Finish()
		return message.ErrMaxAttemptsExceeded
	}
	finCh := make(chan bool)
	defer func() { finCh <- true }()
	go func() {
		select {
		case <-finCh:
			return
		default:
		}
		msg.Touch()
		time.Sleep(j.nsqMessageTouchInterval)
	}()
	err = j.ProcessJudge(jMsg)
	if err != nil {
		log.Println("ERROR:", err)
		return err
	}
	msg.Finish()
	return nil
}
