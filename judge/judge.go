package judge

import "github.com/nsqio/go-nsq"

type Judger struct {
}

func (j *Judger) ConnectNSQ() error {
	config := nsq.NewConfig()
	_ = config
	return nil
}
