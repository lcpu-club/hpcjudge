package server

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lcpu-club/hpcjudge/discovery/protocol"
	"github.com/satori/uuid"
	"nhooyr.io/websocket"
)

type Service struct {
	ID         uuid.UUID `json:"id"`
	Address    string    `json:"address"`
	Type       string    `json:"type"`
	Tags       []string  `json:"tags"`
	tagIndex   map[string]bool
	tagIndexed bool
}

func (svc *Service) indexTags() {
	if svc.tagIndexed {
		return
	}
	svc.tagIndexed = true
	svc.tagIndex = make(map[string]bool)
	for _, tag := range svc.Tags {
		svc.tagIndex[tag] = true
	}
}

func (svc *Service) HasTag(tag string) bool {
	if !svc.tagIndexed {
		svc.indexTags()
	}
	v, ok := svc.tagIndex[tag]
	if ok && v {
		return true
	}
	return false
}

type Server struct {
	id              uuid.UUID
	listen          string
	externalAddress string
	dataFilePath    string
	peers           []string
	services        []*Service
	indexByType     map[string][]*Service
	uniqueness      map[uuid.UUID]bool
	mux             *http.ServeMux
	accessKey       string
	lock            *sync.RWMutex
	peerLock        *sync.RWMutex
	peerTimeout     time.Duration
}

func NewServer(listen string, externalAddress string, dataFilePath string, peers []string, accessKey string, peerTimeout time.Duration) (*Server, error) {
	srv := &Server{
		id:              uuid.NewV4(),
		listen:          listen,
		externalAddress: externalAddress,
		dataFilePath:    dataFilePath,
		peers:           []string{},
		services:        []*Service{},
		indexByType:     make(map[string][]*Service),
		uniqueness:      make(map[uuid.UUID]bool),
		accessKey:       accessKey,
		lock:            &sync.RWMutex{},
		peerLock:        &sync.RWMutex{},
		peerTimeout:     peerTimeout,
	}
	log.Println("Server ID:", srv.id)
	log.Println("Access Key:", srv.accessKey)
	log.Println("Peer Timeout:", srv.peerTimeout)
	log.Println("Data File:", srv.dataFilePath)
	log.Println("External Address:", srv.externalAddress)
	for _, peer := range peers {
		srv.AddPeer(peer, true)
	}
	for _, peerItem := range peers {
		resp, err := srv.doPeerRequest(peerItem, &PeerRequest{
			Operation: OperationPeers,
		})
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}
		for _, peer := range resp.Peers {
			srv.AddPeer(peer, true)
		}
		break
	}
	for _, peerItem := range peers {
		resp, err := srv.doPeerRequest(peerItem, &PeerRequest{
			Operation: OperationList,
		})
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}
		srv.services = append(srv.services, resp.Services...)
		srv.buildIndex()
		log.Println("Initial service list got from", resp.ID)
		break
	}
	return srv, nil
}

type OperationType string

const (
	OperationAdd    = OperationType("add")
	OperationDelete = OperationType("delete")
	OperationNoop   = OperationType("noop")
	OperationList   = OperationType("list")
	OperationPeers  = OperationType("peers")
)

type PeerRequest struct {
	ID        uuid.UUID     `json:"id"`
	Address   string        `json:"address"`
	Operation OperationType `json:"operation"`
	Delta     *Service      `json:"delta"`
}

type PeerResponse struct {
	ID       uuid.UUID  `json:"id"`
	Success  bool       `json:"success"`
	Error    string     `json:"error"`
	Peers    []string   `json:"peers"`
	Services []*Service `json:"services"`
}

func (s *Server) selectFromServiceSlice(a []*Service) *Service {
	if len(a) <= 0 {
		return nil
	}
	return a[rand.Intn(len(a))]
}

func (s *Server) Query(params *protocol.QueryParameters) []*Service {
	if params == nil {
		params = &protocol.QueryParameters{}
	}
	var haystack []*Service
	var result []*Service
	s.lock.RLock()
	if params.Type == "" {
		haystack = s.services
	} else {
		haystack = s.indexByType[params.Type]
	}
	s.lock.RUnlock()
	for _, item := range haystack {
		flag := true
		for _, tag := range params.Tags {
			if !item.HasTag(tag) {
				flag = false
				break
			}
		}
		if flag {
			for _, exTag := range params.ExcludeTags {
				if item.HasTag(exTag) {
					flag = false
					break
				}
			}
		}
		if flag {
			if params.ID != uuid.Nil {
				if params.ID == item.ID {
					result = append(result, item)
					break
				}
				continue
			}
			if params.Address != "" {
				if params.Address == item.Address {
					result = append(result, item)
					break
				}
				continue
			}
			result = append(result, item)
		}
	}
	return result
}

func (s *Server) QueryOne(params *protocol.QueryParameters) *Service {
	return s.selectFromServiceSlice(s.Query(params))
}

func (s *Server) doPeerRequest(peer string, req *PeerRequest) (*PeerResponse, error) {
	client := &http.Client{}
	req.ID = s.id
	req.Address = s.externalAddress
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", peer+"/peer", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if s.accessKey != "" {
		request.Header.Add("X-Access-Key", s.accessKey)
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respJSON, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	rslt := &PeerResponse{}
	err = json.Unmarshal(respJSON, rslt)
	if err != nil {
		return nil, err
	}
	if !rslt.Success {
		return nil, fmt.Errorf(rslt.Error)
	}
	return rslt, nil
}

func (s *Server) NotifyPeersServiceMutation(typ OperationType, svc *Service) {
	if typ != OperationAdd && typ != OperationDelete {
		return
	}
	s.peerLock.RLock()
	wg := &sync.WaitGroup{}
	for _, peer := range s.peers {
		wg.Add(1)
		go func(peer string) {
			_, err := s.doPeerRequest(peer, &PeerRequest{
				Operation: typ,
				Delta:     svc,
			})
			if err != nil {
				log.Println("ERROR:", err)
			}
			wg.Done()
		}(peer)
	}
	s.peerLock.RUnlock()
	wg.Wait()
}

func (s *Server) NormalizeService(service *Service) error {
	service.indexTags()
	if service.ID == uuid.Nil {
		qRslt := s.QueryOne(&protocol.QueryParameters{
			Address: service.Address,
			Type:    service.Type,
		})
		if qRslt != nil {
			service.ID = qRslt.ID
			return protocol.ErrServiceAlreadyExists
		}
		service.ID = uuid.NewV4()
	}
	return nil
}

func (s *Server) Add(service *Service, notifyPeers bool) error {
	err := s.NormalizeService(service)
	if err != nil {
		return err
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	uniqueness, ok := s.uniqueness[service.ID]
	if uniqueness && ok {
		return protocol.ErrServiceAlreadyExists
	}
	s.uniqueness[service.ID] = true
	s.services = append(s.services, service)
	_, ok = s.indexByType[service.Type]
	if !ok {
		s.indexByType[service.Type] = []*Service{}
	}
	s.indexByType[service.Type] = append(s.indexByType[service.Type], service)
	if notifyPeers {
		go s.NotifyPeersServiceMutation(OperationAdd, service)
	}
	log.Println("Service added:", service)
	return nil
}

func (s *Server) findServiceInSlice(haystack []*Service, needle *Service) (int, error) {
	pos := -1
	err := protocol.ErrServiceDoesNotExist
	for p, item := range haystack {
		if needle.ID == uuid.Nil {
			if item.Address == needle.Address && item.Type == needle.Type {
				pos = p
				err = nil
				break
			}
		} else {
			if item.ID == needle.ID {
				pos = p
				err = nil
				break
			}
		}
	}
	return pos, err
}

func (s *Server) Delete(service *Service, notifyPeers bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	pos, err := s.findServiceInSlice(s.services, service)
	if err != nil {
		return err
	}
	s.services = append(s.services[0:pos], s.services[pos+1:]...)
	pos, err = s.findServiceInSlice(s.indexByType[service.Type], service)
	s.indexByType[service.Type] = append(s.indexByType[service.Type][0:pos], s.indexByType[service.Type][pos+1:]...)
	delete(s.uniqueness, service.ID)
	if err != nil {
		return err
	}
	if notifyPeers {
		go s.NotifyPeersServiceMutation(OperationDelete, service)
	}
	log.Println("Service deleted:", service)
	return nil
}

func (s *Server) buildIndex() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.indexByType = make(map[string][]*Service)
	for _, svc := range s.services {
		svc.indexTags()
		_, ok := s.indexByType[svc.Type]
		if !ok {
			s.indexByType[svc.Type] = []*Service{}
		}
		s.indexByType[svc.Type] = append(s.indexByType[svc.Type], svc)
	}
}

var ErrPeerAlreadyExists = fmt.Errorf("peer already exists")
var ErrPeerNotExist = fmt.Errorf("peer not exist")
var ErrPeerIsSelf = fmt.Errorf("peer is self")

func (s *Server) HasPeer(peerAddress string) bool {
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()
	for _, peer := range s.peers {
		if peerAddress == peer {
			return true
		}
	}
	return false
}

func (s *Server) AddPeer(peerAddress string, inform bool) error {
	if s.HasPeer(peerAddress) {
		return ErrPeerAlreadyExists
	}
	s.peerLock.Lock()
	s.peers = append(s.peers, peerAddress)
	s.peerLock.Unlock()
	if inform {
		_, err := s.doPeerRequest(peerAddress, &PeerRequest{
			Operation: OperationNoop,
		})
		if err != nil {
			return err
		}
	}
	log.Println("Peer added:", peerAddress)
	return nil
}

func (s *Server) RemovePeer(peerAddress string) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	pos := -1
	for p, peer := range s.peers {
		if peer == peerAddress {
			pos = p
		}
	}
	if pos == -1 {
		return ErrPeerAlreadyExists
	}
	s.peers = append(s.peers[0:pos], s.peers[pos+1:]...)
	log.Println("Peer removed:", peerAddress)
	return nil
}

func ServiceToProtocolService(svc *Service) *protocol.Service {
	if svc == nil {
		return nil
	}
	return &protocol.Service{
		ID:      svc.ID,
		Address: svc.Address,
		Type:    svc.Type,
		Tags:    svc.Tags,
	}
}

func ProtocolServiceToService(svc *protocol.Service) *Service {
	if svc == nil {
		return nil
	}
	return &Service{
		ID:         svc.ID,
		Address:    svc.Address,
		Type:       svc.Type,
		Tags:       svc.Tags,
		tagIndexed: false,
	}
}

func ServicesToProtocolServices(services []*Service) []*protocol.Service {
	var rslt []*protocol.Service
	for _, svc := range services {
		rslt = append(rslt, ServiceToProtocolService(svc))
	}
	return rslt
}

func (s *Server) error500(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(500)
	w.Write([]byte("500 Internal Server Error"))
}

func (s *Server) error400(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(400)
	w.Write([]byte("400 Bad Request"))
}

func (s *Server) error403(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(403)
	w.Write([]byte("403 Access Denied"))
}

func (s *Server) responseJSON(w http.ResponseWriter, resp interface{}) {
	j, err := json.Marshal(resp)
	if err != nil {
		log.Println("ERROR:", err)
		s.error500(w)
		return
	}
	w.Write(j)
}

func (s *Server) parseRequest(w http.ResponseWriter, r *http.Request, to interface{}) (ok bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("ERROR:", err)
		s.error500(w)
		return false
	}
	err = json.Unmarshal(body, to)
	if err != nil {
		log.Println("ERROR:", err)
		s.error400(w)
		return false
	}
	return true
}

var ErrUnknownOperation = fmt.Errorf("unknown operation")
var ErrPeerFail = fmt.Errorf("peer fail")

func (s *Server) HandlePeer(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(r.Method) != "POST" {
		s.error400(w)
		return
	}
	req := &PeerRequest{
		Delta: &Service{},
	}
	if !s.parseRequest(w, r, req) {
		return
	}
	if req.ID == uuid.Nil || req.Address == "" {
		s.error500(w)
		return
	}
	if req.ID == s.id {
		s.error400(w)
		return
	}
	err := s.AddPeer(req.Address, false)
	if err != nil && err != ErrPeerAlreadyExists {
		s.responseJSON(w, &PeerResponse{
			ID:      s.id,
			Success: false,
			Error:   ErrPeerFail.Error(),
		})
		s.RemovePeer(req.Address)
		return
	}
	resp := &PeerResponse{
		ID:      s.id,
		Success: true,
	}
	switch req.Operation {
	case OperationAdd:
		err = s.Add(req.Delta, false)
		if err != nil && err != protocol.ErrServiceAlreadyExists {
			resp.Success = false
			resp.Error = err.Error()
		}
		s.responseJSON(w, resp)
	case OperationDelete:
		err = s.Delete(req.Delta, false)
		if err != nil && err != protocol.ErrServiceDoesNotExist {
			resp.Success = false
			resp.Error = err.Error()
		}
		s.responseJSON(w, resp)
	case OperationNoop:
		s.responseJSON(w, resp)
	case OperationList:
		resp.Services = s.services
		s.responseJSON(w, resp)
	case OperationPeers:
		resp.Peers = s.peers
		s.responseJSON(w, resp)
	default:
		s.responseJSON(w, &PeerResponse{
			ID:      s.id,
			Success: false,
			Error:   ErrUnknownOperation.Error(),
		})
		return
	}
}

func (s *Server) HandleAdd(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(r.Method) != "POST" {
		s.error400(w)
		return
	}
	req := &protocol.AddRequest{
		Service: &protocol.Service{},
	}
	if !s.parseRequest(w, r, req) {
		return
	}
	svc := ProtocolServiceToService(req.Service)
	err := s.Add(svc, true)
	req.Service.ID = svc.ID
	resp := &protocol.AddResponse{
		ResponseBase: protocol.ResponseBase{
			Success: true,
		},
		Service: req.Service,
	}
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
	}
	s.responseJSON(w, resp)
}

func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println("ERROR:", err)
		s.error400(w)
		return
	}
	defer conn.Close(websocket.StatusAbnormalClosure, "handler exited unexpectedly")
	ctx := r.Context()
	var svc *Service
	for {
		_, msg, err := conn.Read(ctx)
		eReal := new(websocket.CloseError)
		if errors.As(err, eReal) {
			if eReal.Code != websocket.StatusNormalClosure && eReal.Code != websocket.StatusNoStatusRcvd {
				log.Println("ERROR:", eReal.Error())
			}
			if svc != nil {
				err := s.Delete(svc, true)
				if err != nil {
					log.Println("ERROR:", err)
				}
			}
			return
		}
		if err != nil {
			log.Println("ERROR:", err)
			conn.Close(websocket.StatusAbnormalClosure, err.Error())
			if svc != nil {
				err := s.Delete(svc, true)
				if err != nil {
					log.Println("ERROR:", err)
				}
			}
			return
		}
		m := &protocol.ClientRegisterMessage{
			Data: &protocol.Service{},
		}
		resp := &protocol.ServerRegisterMessage{}
		resp.Success = true
		err = json.Unmarshal(msg, m)
		if err != nil {
			resp.Success = false
			resp.Error = err.Error()
			j, err := json.Marshal(resp)
			if err != nil {
				log.Println("ERROR:", err)
			}
			err = conn.Write(ctx, websocket.MessageText, j)
			if err != nil {
				log.Println("ERROR:", err)
			}
			continue
		}
		switch m.Operation {
		case protocol.RegisterOperationNoop:
			// Just do nothing...
		case protocol.RegisterOperationAdd:
			if svc == nil {
				resp.Success = false
				resp.Error = protocol.ErrNoServiceInformed.Error()
				break
			}
			err = s.Add(svc, true)
			if err != nil {
				resp.Success = false
				resp.Error = err.Error()
			}
		case protocol.RegisterOperationDelete:
			if svc == nil {
				resp.Success = false
				resp.Error = protocol.ErrNoServiceInformed.Error()
				break
			}
			err = s.Delete(svc, true)
			if err != nil {
				resp.Success = false
				resp.Error = err.Error()
			}
		case protocol.RegisterOperationInform:
			svc = ProtocolServiceToService(m.Data)
			err = s.NormalizeService(svc)
			resp.Service = ServiceToProtocolService(svc)
			if err != nil && err != protocol.ErrServiceAlreadyExists {
				resp.Success = false
				resp.Error = err.Error()
			}
		case protocol.RegisterOperationHas:
			if svc == nil {
				resp.Success = false
				resp.Error = protocol.ErrNoServiceInformed.Error()
				break
			}
			s.lock.RLock()
			resp.Has = s.uniqueness[svc.ID]
			s.lock.RUnlock()
		default:
			resp.Success = false
			resp.Error = protocol.ErrUnknownOperation.Error()
		}
		j, err := json.Marshal(resp)
		if err != nil {
			log.Println("ERROR:", err)
		}
		err = conn.Write(ctx, websocket.MessageText, j)
		if err != nil {
			log.Println("ERROR:", err)
		}
	}
}

func (s *Server) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(r.Method) != "POST" {
		s.error400(w)
		return
	}
	req := &protocol.DeleteRequest{
		Service: &protocol.Service{},
	}
	if !s.parseRequest(w, r, req) {
		return
	}
	err := s.Delete(ProtocolServiceToService(req.Service), true)
	resp := &protocol.DeleteResponse{
		ResponseBase: protocol.ResponseBase{
			Success: true,
		},
	}
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
	}
	s.responseJSON(w, resp)
}

func (s *Server) HandleQuery(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(r.Method) != "POST" {
		s.error400(w)
		return
	}
	req := &protocol.QueryRequest{
		Condition: &protocol.QueryParameters{},
	}
	if !s.parseRequest(w, r, req) {
		return
	}
	qRslt := s.QueryOne(req.Condition)
	resp := &protocol.QueryResponse{
		ResponseBase: protocol.ResponseBase{
			Success: true,
		},
		Data: ServiceToProtocolService(qRslt),
	}
	if qRslt == nil {
		resp.Success = false
		resp.Error = protocol.ErrNoServiceAvailable.Error()
	}
	s.responseJSON(w, resp)
}

func (s *Server) HandleList(w http.ResponseWriter, r *http.Request) {
	req := &protocol.ListRequest{
		Condition: &protocol.QueryParameters{},
	}
	if strings.ToUpper(r.Method) == "POST" {
		if !s.parseRequest(w, r, req) {
			return
		}
	}
	qRslt := s.Query(req.Condition)
	resp := &protocol.ListResponse{
		ResponseBase: protocol.ResponseBase{
			Success: true,
		},
		Data: ServicesToProtocolServices(qRslt),
	}
	if len(qRslt) == 0 {
		resp.Success = false
		resp.Error = protocol.ErrNoServiceAvailable.Error()
	}
	s.responseJSON(w, resp)
}

func (s *Server) HandleListPeers(w http.ResponseWriter, r *http.Request) {
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()
	resp := &protocol.ListPeersResponse{
		ResponseBase: protocol.ResponseBase{
			Success: true,
		},
	}
	resp.Data = s.peers
	if len(s.peers) == 0 {
		resp.Success = false
		resp.Error = protocol.ErrNoPeers.Error()
	}
	s.responseJSON(w, resp)
}

func (s *Server) HandleRemovePeer(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(r.Method) != "POST" {
		s.error400(w)
		return
	}
	req := &protocol.RemovePeerRequest{}
	if !s.parseRequest(w, r, req) {
		return
	}
	err := s.RemovePeer(req.Peer)
	resp := &protocol.RemovePeerResponse{
		ResponseBase: protocol.ResponseBase{
			Success: true,
		},
	}
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
	}
	s.responseJSON(w, resp)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "hpc-discovery/0.1.0")
	w.Header().Add("X-Discovery-ID", s.id.String())
	w.Header().Add("Content-Type", "application/json")
	accessKey := r.Header.Get("X-Access-Key")
	if accessKey == "" {
		accessKey = r.URL.Query().Get("access-key")
	}
	if s.accessKey != "" && subtle.ConstantTimeCompare([]byte(accessKey), []byte(s.accessKey)) == 0 {
		s.error403(w)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) Start() error {
	if s.mux == nil {
		s.mux = http.NewServeMux()
		s.mux.HandleFunc("/peer", s.HandlePeer)
		s.mux.HandleFunc("/add", s.HandleAdd)
		s.mux.HandleFunc("/register", s.HandleRegister)
		s.mux.HandleFunc("/delete", s.HandleDelete)
		s.mux.HandleFunc("/query", s.HandleQuery)
		s.mux.HandleFunc("/list", s.HandleList)
		s.mux.HandleFunc("/peers/list", s.HandleListPeers)
		s.mux.HandleFunc("/peers/remove", s.HandleRemovePeer)
	}
	log.Println("Listening on", s.listen)
	return http.ListenAndServe(s.listen, s)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
