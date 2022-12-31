package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/lcpu-club/hpcjudge/bridge/api"
	"github.com/lcpu-club/hpcjudge/bridge/configure"
	"github.com/lcpu-club/hpcjudge/common"
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
	cs               *common.CommonServer
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
	err = s.discoveryService.Add()
	if err != nil {
		return err
	}
	s.cs = common.NewCommonServer(s.configure.Listen, s.configure.SecretKey)
	s.registerRoutes(s.cs.GetMux())
	return nil
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/fetch-object", s.HandleFetchObject)
	mux.HandleFunc("/calculate-path", s.HandleCalculatePath)
}

func (s *Server) getStoragePath(partition string, path string) (string, error) {
	partPath, ok := s.configure.StoragePath[partition]
	if !ok {
		return "", api.ErrPartitionNotFound
	}
	p := filepath.Join(partPath, path)
	if !strings.HasPrefix(p, partPath) {
		return "", api.ErrPathOverflowsPartitionPath
	}
	return p, nil
}

func (s *Server) HandleFetchObject(w http.ResponseWriter, r *http.Request) {
	req := new(api.FetchObjectRequest)
	if !s.cs.ParseRequest(w, r, req) {
		return
	}
	resp := &api.FetchObjectResponse{
		ResponseBase: common.ResponseBase{
			Success: true,
		},
	}
	targetPath, err := s.getStoragePath(req.Path.Partition, req.Path.Path)
	if err != nil {
		resp.SetError(err)
		s.cs.Respond(w, resp)
		return
	}
	target, err := os.Create(targetPath)
	if err != nil {
		log.Println("ERROR:", err)
		resp.SetError(api.ErrFileCreationError)
		s.cs.Respond(w, resp)
		return
	}
	remote, err := http.Get(req.ObjectURL)
	if err != nil || remote.StatusCode != 200 {
		log.Println("ERROR:", err)
		resp.SetError(api.ErrFailedToFetch)
		s.cs.Respond(w, resp)
		return
	}
	_, err = io.Copy(target, remote.Body)
	if err != nil {
		log.Println("ERROR:", err)
		resp.SetError(err)
	}
	s.cs.Respond(w, resp)
}

func (s *Server) HandleCalculatePath(w http.ResponseWriter, r *http.Request) {
	req := new(api.CalculatePathRequest)
	if !s.cs.ParseRequest(w, r, req) {
		return
	}
	resp := &api.CalculatePathResponse{
		ResponseBase: common.ResponseBase{
			Success: true,
		},
	}
	p, err := s.getStoragePath(req.Path.Partition, req.Path.Path)
	if err != nil {
		resp.SetError(err)
	}
	resp.Path = p
	s.cs.Respond(w, resp)
}

func (s *Server) Start() error {
	err := s.discoveryService.Add()
	if err != nil {
		return err
	}
	return s.cs.Start()
}

func (s *Server) Suspend() error {
	return s.discoveryService.Delete()
}

func (s *Server) Close() error {
	err := s.discoveryService.Close()
	return err
}
