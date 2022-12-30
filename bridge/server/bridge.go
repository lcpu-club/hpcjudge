package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding"
	"fmt"
	"hash"
	"net/http"

	"github.com/lcpu-club/hpcjudge/bridge/configure"
	"github.com/lcpu-club/hpcjudge/common/consts"
	"github.com/lcpu-club/hpcjudge/discovery"
	discoveryProtocol "github.com/lcpu-club/hpcjudge/discovery/protocol"
	"github.com/satori/uuid"
)

type Server struct {
	id               uuid.UUID
	discoveryService *discovery.Service
	discovery        *discovery.Client
	configure        *configure.Configure
	mux              *http.ServeMux
	signatureHasher  func() hash.Hash
}

func NewServer(conf *configure.Configure) (*Server, error) {
	srv := new(Server)
	return srv, srv.Init(conf)
}

var ErrNilConfigure = fmt.Errorf("nil configure")

func (s *Server) Init(conf *configure.Configure) error {
	if conf != nil {
		s.configure = conf
	}
	if conf == nil && s.configure == nil {
		return ErrNilConfigure
	}
	s.id = uuid.UUID(s.configure.ID)
	if s.configure.ID == uuid.Nil {
		s.id = uuid.NewV4()
	}
	s.discoveryService = discovery.NewService(context.Background(), s.configure.Discovery.Address, s.configure.Discovery.AccessKey)
	s.discovery = discovery.NewClient(s.configure.Discovery.Address, s.configure.Discovery.AccessKey, s.configure.Discovery.Timeout)
	rSvc, err := s.discoveryService.Inform(&discoveryProtocol.Service{
		ID:      s.id,
		Address: s.configure.ExternalAddress,
		Type:    consts.HpcBridgeDiscoveryType,
		Tags:    s.configure.Tags,
	})
	if err != nil {
		return err
	}
	s.id = rSvc.ID
	s.signatureHasher = func() hash.Hash {
		return hmac.New(sha256.New, s.configure.SecretKey)
	}
	return nil
}

func (s *Server) validateMessageSignature(msg []byte, signature string) bool {
	hmacHasher := s.signatureHasher()
	_, err := hmacHasher.Write(msg)
	if err != nil {
		panic(err)
	}
	txt, err := hmacHasher.(encoding.TextMarshaler).MarshalText()
	if err != nil {
		panic(err)
	}
	return string(txt) == signature
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) Start() error {
	err := s.discoveryService.Add()
	if err != nil {
		return err
	}
	err = http.ListenAndServe(s.configure.Listen, s)
	return err
}

func (s *Server) Suspend() error {
	return s.discoveryService.Delete()
}

func (s *Server) Close() error {
	err := s.discoveryService.Close()
	return err
}
