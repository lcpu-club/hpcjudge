package judge

import (
	"context"
	"encoding/json"
	"log"

	"github.com/lcpu-club/hpcjudge/common/consts"
	"github.com/lcpu-club/hpcjudge/discovery"
	discoveryProtocol "github.com/lcpu-club/hpcjudge/discovery/protocol"
	"github.com/lcpu-club/hpcjudge/judge/configure"
	"github.com/lcpu-club/hpcjudge/judge/message"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nsqio/go-nsq"
	"github.com/satori/uuid"
)

type Judger struct {
	id               uuid.UUID
	nsqConsumer      *nsq.Consumer
	nsqReport        *nsq.Producer
	discoveryClient  *discovery.Client
	discoveryService *discovery.Service
	configure        *configure.Configure
	minio            *minio.Client
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
	err := j.connectMinIO()
	if err != nil {
		log.Println("Connect to MinIO failed")
		return err
	}
	err = j.connectDiscovery()
	if err != nil {
		log.Println("Connect to Discovery failed")
		return err
	}
	err = j.connectNSQ()
	if err != nil {
		log.Println("Connect to NSQ failed")
	}
	return err
}

func (j *Judger) connectDiscovery() error {
	j.discoveryClient = discovery.NewClient(j.configure.Discovery.Address, j.configure.Discovery.AccessKey, j.configure.Discovery.Timeout)
	j.discoveryService = discovery.NewService(context.Background(), j.configure.Discovery.Address, j.configure.Discovery.AccessKey)
	err := j.discoveryService.Connect()
	if err != nil {
		return err
	}
	s, err := j.discoveryService.Inform(&discoveryProtocol.Service{
		ID:      j.id,
		Address: j.configure.ExternalAddress,
		Type:    consts.HpcJudgeDiscoveryType,
		Tags:    j.configure.Tags,
	})
	if err != nil {
		return err
	}
	j.id = s.ID
	return j.discoveryService.Add()
}

func (j *Judger) connectNSQ() error {
	config := nsq.NewConfig()
	config.AuthSecret = j.configure.Nsq.AuthSecret
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
		Creds:  credentials.NewStaticV4(j.configure.MinIO.Credentials.AccessKey, j.configure.MinIO.Credentials.SecretKey, ""),
		Secure: j.configure.MinIO.SSL,
	})
	if err != nil {
		return err
	}
	log.Println("Connected to MinIO Server")
	return nil
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

func (j *Judger) HandleMessage(msg *nsq.Message) error {
	msg.Touch()
	jMsg := &message.JudgeMessage{}
	err := json.Unmarshal(msg.Body, jMsg)
	if err != nil {
		return err
	}
	return nil
}
