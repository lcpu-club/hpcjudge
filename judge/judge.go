package judge

import (
	"context"
	"log"

	"github.com/lcpu-club/hpcjudge/discovery"
	discoveryProtocol "github.com/lcpu-club/hpcjudge/discovery/protocol"
	"github.com/lcpu-club/hpcjudge/judge/configure"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nsqio/go-nsq"
	"github.com/satori/uuid"
)

type Judger struct {
	id                 uuid.UUID
	nsqConsumerJudge   *nsq.Consumer
	nsqConsumerSandbox *nsq.Consumer
	nsqReport          *nsq.Producer
	discoveryClient    *discovery.Client
	discoveryService   *discovery.Service
	configure          *configure.Configure
	minio              *minio.Client
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

func (j *Judger) Start() {
	j.connectMinIO()
	j.connectDiscovery()
	j.connectNSQ()
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
		Type:    j.configure.Discovery.InformType,
		Tags:    j.configure.Discovery.InformTags,
	})
	if err != nil {
		return err
	}
	j.id = s.ID
	return nil
}

func (j *Judger) connectNSQ() error {
	config := nsq.NewConfig()
	config.AuthSecret = j.configure.Nsq.AuthSecret
	var err error
	j.nsqConsumerJudge, err = nsq.NewConsumer(j.configure.Nsq.NsqLookupd.Topics.Judge, j.configure.Nsq.NsqLookupd.Channel, config)
	if err != nil {
		return err
	}
	err = j.nsqConsumerJudge.ConnectToNSQLookupds(j.configure.Nsq.NsqLookupd.Address)
	if err != nil {
		return err
	}
	j.nsqConsumerJudge.AddConcurrentHandlers(nsq.HandlerFunc(j.HandleMessageJudge), j.configure.Nsq.Concurrent)
	j.nsqConsumerSandbox, err = nsq.NewConsumer(j.configure.Nsq.NsqLookupd.Topics.Sandbox, j.configure.Nsq.NsqLookupd.Channel, config)
	if err != nil {
		return err
	}
	err = j.nsqConsumerSandbox.ConnectToNSQLookupds(j.configure.Nsq.NsqLookupd.Address)
	if err != nil {
		return err
	}
	j.nsqConsumerSandbox.AddConcurrentHandlers(nsq.HandlerFunc(j.HandleMessageSandbox), j.configure.Nsq.Concurrent)
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

func (j *Judger) HandleMessageJudge(msg *nsq.Message) error {
	return nil
}

func (j *Judger) HandleMessageSandbox(msg *nsq.Message) error {
	return nil
}

func (j *Judger) Init() error {
	return nil
}

func (j *Judger) Run() error {
	return nil
}
