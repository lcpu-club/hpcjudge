package server

import (
	"context"
	"net/http"
	"time"

	"github.com/lcpu-club/hpcjudge/discovery"
	discoveryProtocol "github.com/lcpu-club/hpcjudge/discovery/protocol"
	"github.com/satori/uuid"
)

type Server struct {
	discoveryService   *discovery.Service
	discovery          *discovery.Client
	discoveryAddress   []string
	discoveryAccessKey string
	discoveryTimeout   time.Duration
	discoveryType      string
	id                 uuid.UUID
	externalAddress    string
	accessKey          string
	listen             string
	mux                *http.ServeMux
}

func (s *Server) Init() error {
	s.discoveryService = discovery.NewService(context.Background(), s.discoveryAddress, s.discoveryAccessKey)
	s.discovery = discovery.NewClient(s.discoveryAddress, s.discoveryAccessKey, s.discoveryTimeout)
	rSvc, err := s.discoveryService.Inform(&discoveryProtocol.Service{
		ID:      s.id,
		Address: s.externalAddress,
		Type:    s.discoveryType,
	})
	if err != nil {
		return err
	}
	s.id = rSvc.ID

	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) Start() error {
	err := s.discoveryService.Add()
	if err != nil {
		return err
	}
	err = http.ListenAndServe(s.listen, s)
	return err
}

func (s *Server) Suspend() error {
	return s.discoveryService.Delete()
}

func (s *Server) Close() error {
	err := s.discoveryService.Close()
	return err
}
