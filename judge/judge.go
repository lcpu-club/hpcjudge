package judge

import (
	"github.com/lcpu-club/hpcjudge/judge/configure"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nsqio/go-nsq"
)

type Judger struct {
	nsqConsumer *nsq.Consumer
	nsqReport   *nsq.Producer
	configure   *configure.Configure
	minio       *minio.Client
}

func NewJudger(conf *configure.Configure) (*Judger, error) {
	j := new(Judger)
	j.configure = conf
	return j, nil
}

func (j *Judger) ConnectNSQ() error {
	config := nsq.NewConfig()
	config.AuthSecret = j.configure.Nsq.AuthSecret
	var err error
	j.nsqConsumer, err = nsq.NewConsumer(j.configure.Nsq.NsqLookupd.Topic, j.configure.Nsq.NsqLookupd.Channel, config)
	if err != nil {
		return err
	}
	err = j.nsqConsumer.ConnectToNSQLookupds(j.configure.Nsq.NsqLookupd.Address)
	if err != nil {
		return err
	}
	j.nsqConsumer.AddConcurrentHandlers(j, 4)
	j.nsqReport, err = nsq.NewProducer(j.configure.Nsq.Nsqd.Address, config)
	return err
}

func (j *Judger) ConnectMinIO() error {
	var err error
	j.minio, err = minio.New(j.configure.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(j.configure.MinIO.Credentials.AccessKey, j.configure.MinIO.Credentials.SecretKey, ""),
		Secure: j.configure.MinIO.SSL,
	})
	if err != nil {
		return err
	}
	return nil
}

func (j *Judger) HandleMessage(msg *nsq.Message) error {
	return nil
}

func (j *Judger) Init() error {
	return nil
}

func (j *Judger) Run() error {
	return nil
}
