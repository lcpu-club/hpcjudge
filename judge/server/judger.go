package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/lcpu-club/hpcjudge/bridge"
	bridgeApi "github.com/lcpu-club/hpcjudge/bridge/api"
	"github.com/lcpu-club/hpcjudge/common"
	"github.com/lcpu-club/hpcjudge/common/consts"
	"github.com/lcpu-club/hpcjudge/common/models"
	"github.com/lcpu-club/hpcjudge/common/version"
	"github.com/lcpu-club/hpcjudge/discovery"
	discoveryProtocol "github.com/lcpu-club/hpcjudge/discovery/protocol"
	"github.com/lcpu-club/hpcjudge/judge/configure"
	"github.com/lcpu-club/hpcjudge/judge/message"
	"github.com/lcpu-club/hpcjudge/judge/problem"
	spawnConsts "github.com/lcpu-club/hpcjudge/spawncmd/consts"
	spawnModels "github.com/lcpu-club/hpcjudge/spawncmd/models"
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

func (j *Judger) Start() error {
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

func (j *Judger) Wait() error {
	select {}
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
	if j.configure.EnableStatistics {
		j.redisConn.Do("SET", j.configure.Redis.Prefix+"stats-version", version.Version)
	}
	log.Println("Connected to Redis Server")
	return nil
}

func (j *Judger) listenMinIOEvent() {
	chResults := j.minio.ListenBucketNotification(
		context.Background(),
		j.configure.MinIO.Buckets.Solution, "", consts.JudgeReportFile, []string{
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
				id, err := j.resultObjectKeyToSolutionID(k, consts.JudgeReportFile)
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
					err := j.setRequestNotExist(id + v)
					if err != nil {
						log.Println("ERROR:", err)
					}
					continue
				}
				r := new(models.JudgeResult)
				err = json.Unmarshal(rsltJSON, r)
				if err != nil {
					log.Println("ERROR:", err)
					err = j.publishToReport(&message.JudgeReportMessage{
						Success:   false,
						Error:     "Invalid report from judge script: " + err.Error(),
						Done:      true,
						Timestamp: time.Now().UnixMicro(),
					})
					if err != nil {
						log.Println("ERROR:", err)
					}
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
		j.configure.MinIO.Buckets.Solution, "", consts.RunCommandReportFile, []string{
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
				id, err := j.resultObjectKeyToSolutionID(k, consts.RunCommandReportFile)
				if err != nil {
					continue
				}
				exists, err := j.checkIfRequestExists(id+v, j.configure.Redis.Expire.Report)
				if err != nil {
					log.Println("ERROR:", err)
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
				r := new(bridgeApi.ExecuteCommandResponse)
				err = json.Unmarshal(rsltJSON, r)
				if err != nil {
					log.Println("ERROR:", err)
					continue
				}
				resp := &message.JudgeReportMessage{
					SolutionID: id,
					Success:    true,
					Done:       true,
					Timestamp:  time.Now().UnixMicro(),
				}
				if (!r.Success) || r.ExitStatus != 0 {
					resp.Success = false
					if !r.Success {
						resp.Error = r.Error
					} else if r.ExitStatus != 0 {
						resp.Error = r.StdErr
						if r.StdErr == "" {
							resp.Error = r.StdOut
						}
					}
					err = j.publishToReport(resp)
					if err != nil {
						log.Println("ERROR:", err)
					}
				} else {
					go func() {
						time.Sleep(2500 * time.Millisecond)
						ex, err := j.checkIfRequestExists(id, j.configure.Redis.Expire.Judge)
						if err != nil {
							log.Println("ERROR:", err)
						}
						if ex {
							err = j.setRequestNotExist(id)
							if err != nil {
								log.Println("ERROR:", err)
							}
							err = j.publishToReport(&message.JudgeReportMessage{
								Success:   false,
								Error:     "Judge script exited before reporting done",
								Done:      true,
								Timestamp: time.Now().UnixMicro() - 100000, // avoid competence
							})
							if err != nil {
								log.Println("ERROR:", err)
							}
						}
					}()
				}
			}
		}
	}()
}

func (j *Judger) resultObjectKeyToSolutionID(key string, suffix string) (string, error) {
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
	rInteger, ok := rslt.(int64)
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
	_, err := j.redisConn.Do("DEL", j.configure.Redis.Prefix+key)
	return err
}

func (j *Judger) publishToReport(msg *message.JudgeReportMessage) error {
	mText, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	err = j.nsqReport.Publish(j.configure.Nsq.Topics.Report, mText)
	if err != nil {
		return err
	}
	if !msg.Success {
		if j.configure.EnableStatistics {
			go j.redisConn.Do("INCR", j.configure.Redis.Prefix+"stats-judge-failed")
		}
	}
	if msg.Done {
		err := j.setRequestNotExist(msg.SolutionID)
		if err != nil {
			return err
		}
	}
	return nil
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
	if j.configure.EnableStatistics {
		go j.redisConn.Do("INCR", j.configure.Redis.Prefix+"stats-judge-all")
	}
	probMeta, err := problem.GetProblemMeta(
		context.Background(), j.minio, j.configure.MinIO.Buckets.Problem, msg.ProblemID,
	)
	if err != nil {
		return fmt.Errorf("get-problem-meta: %v", err)
	}
	bridgeSvc, err := j.discoverBridge(probMeta.Environment.Tags, probMeta.Environment.ExcludeTags)
	if err != nil {
		return err
	}
	cc := common.NewCommonSignedClient(
		bridgeSvc.Address, []byte(j.configure.Bridge.SecretKey), j.configure.Bridge.Timeout,
	)
	bc := bridge.NewClient(cc)
	url, err := j.minio.PresignedGetObject(
		context.Background(), j.configure.MinIO.Buckets.Solution,
		filepath.Join(msg.SolutionID, consts.OSSSolutionFileName),
		j.configure.MinIO.PresignedExpiry,
		nil,
	)
	if err != nil {
		return fmt.Errorf("pre-sign-report-url: %v", err)
	}
	err = bc.FetchObject(
		url.String(), "solution", filepath.Join(msg.SolutionID, consts.SolutionFileName), msg.Username, os.FileMode(0600),
	)
	// NOTICE: Due to turning to async process, this is not usable
	// defer bc.RemoveFile("solution", filepath.Join(msg.SolutionID, consts.SolutionFileName))
	if err != nil {
		return fmt.Errorf("bridge-fetch-object: %v", err)
	}
	runData := &spawnModels.RunJudgeScriptData{
		ProblemID:  msg.ProblemID,
		SolutionID: msg.SolutionID,
		Username:   msg.Username,
		ResourceControl: &spawnModels.ResourceControl{
			Memory: probMeta.Environment.ScriptLimits.Memory,
			CPU:    probMeta.Environment.ScriptLimits.CPU,
		},
		Command:            probMeta.Entrance.Command,
		Script:             probMeta.Entrance.Script,
		AutoRemoveSolution: true,
	}
	runArgs, err := json.Marshal(runData)
	if err != nil {
		return err
	}
	reportURL, err := j.minio.PresignedPutObject(
		context.Background(),
		j.configure.MinIO.Buckets.Solution,
		consts.RunCommandReportFile,
		j.configure.MinIO.PresignedExpiry,
	)
	if err != nil {
		return err
	}
	err = bc.ExecuteCommandAsync(
		j.configure.SpawnCmd,
		[]string{
			"--data",
			string(runArgs),
		},
		"home",
		msg.Username,
		msg.Username,
		[]string{
			spawnConsts.SpawnEnvVar + "=" + spawnConsts.SpawnEnvVarValue,
		},
		reportURL.String(),
	)
	if err != nil {
		return fmt.Errorf("bridge-execute-command: %v", err)
	}
	return nil
}

func (j *Judger) HandleMessage(msg *nsq.Message) error {
	msg.Touch()
	jMsg := &message.JudgeMessage{}
	log.Println("judge message:", string(msg.Body))
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
			Timestamp:  time.Now().UnixMicro(),
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
		errs := j.setRequestNotExist(jMsg.SolutionID)
		if errs != nil {
			log.Println("ERROR:", errs)
		}
		log.Println("ERROR:", err)
		if msg.Attempts == uint16(j.configure.Nsq.MaxAttempts) {
			err := j.publishToReport(&message.JudgeReportMessage{
				SolutionID: jMsg.SolutionID,
				Success:    false,
				Done:       true,
				Error:      err.Error(),
				Score:      0,
				Message:    "Internal Error: " + err.Error(),
				Timestamp:  time.Now().UnixMicro(),
			})
			if err != nil {
				log.Println("ERROR:", err)
			}
			msg.Finish()
			return err
		}
		msg.Requeue(-1)
		return err
	}
	msg.Finish()
	return nil
}
