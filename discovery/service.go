package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lcpu-club/hpcjudge/discovery/protocol"
	"github.com/satori/uuid"
	"nhooyr.io/websocket"
)

var ErrNotConnected = fmt.Errorf("not connected")

type Service struct {
	address   []string
	accessKey string
	conn      *websocket.Conn
	ctx       context.Context
	service   *protocol.Service
}

func NewService(ctx context.Context, address []string, accessKey string) *Service {
	svc := &Service{
		address:   address,
		accessKey: accessKey,
		ctx:       ctx,
	}
	return svc
}

func (s *Service) connectN(n int) error {
	header := make(http.Header)
	header.Add("X-Access-Key", s.accessKey)
	var err error
	s.conn, _, err = websocket.Dial(s.ctx, s.address[n]+"/register", &websocket.DialOptions{
		HTTPHeader: header,
	})
	if err != nil {
		s.conn = nil
		return err
	}
	return nil
}

func (s *Service) Connect() error {
	var err error
	for k := range s.address {
		err = s.connectN(k)
		if err == nil {
			return nil
		}
	}
	return err
}

func (s *Service) Close() error {
	if s.conn == nil {
		return ErrNotConnected
	}
	return s.conn.Close(websocket.StatusNormalClosure, "disconnect")
}

func (s *Service) sendMessage(cMsg *protocol.ClientRegisterMessage, sMsg *protocol.ServerRegisterMessage) error {
	cMsgText, err := json.Marshal(cMsg)
	if err != nil {
		return err
	}
	err = s.conn.Write(s.ctx, websocket.MessageText, cMsgText)
	if err != nil {
		return err
	}
	_, sMsgText, err := s.conn.Read(s.ctx)
	if err != nil {
		return err
	}
	err = json.Unmarshal(sMsgText, sMsg)
	if err != nil {
		return err
	}
	return sMsg.GetError()
}

func (s *Service) Inform(service *protocol.Service) (*protocol.Service, error) {
	resp := &protocol.ServerRegisterMessage{
		Service: &protocol.Service{},
	}
	err := s.sendMessage(&protocol.ClientRegisterMessage{
		Operation: protocol.RegisterOperationInform,
		Data:      service,
	}, resp)
	if resp.Service.ID != uuid.Nil {
		s.service = resp.Service
	}
	return s.service, err
}

func (s *Service) Add() error {
	resp := &protocol.ServerRegisterMessage{
		Service: &protocol.Service{},
	}
	return s.sendMessage(&protocol.ClientRegisterMessage{
		Operation: protocol.RegisterOperationAdd,
	}, resp)
}

func (s *Service) Delete() error {
	resp := &protocol.ServerRegisterMessage{
		Service: &protocol.Service{},
	}
	return s.sendMessage(&protocol.ClientRegisterMessage{
		Operation: protocol.RegisterOperationDelete,
	}, resp)
}

func (s *Service) Has() (bool, error) {
	resp := &protocol.ServerRegisterMessage{
		Service: &protocol.Service{},
	}
	err := s.sendMessage(&protocol.ClientRegisterMessage{
		Operation: protocol.RegisterOperationHas,
	}, resp)
	return resp.Has, err
}

func (s *Service) Noop() error {
	resp := &protocol.ServerRegisterMessage{
		Service: &protocol.Service{},
	}
	return s.sendMessage(&protocol.ClientRegisterMessage{
		Operation: protocol.RegisterOperationNoop,
	}, resp)
}

func (s *Service) Ping() error {
	return s.conn.Ping(s.ctx)
}
