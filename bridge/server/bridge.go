package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/lcpu-club/hpcjudge/bridge/api"
	"github.com/lcpu-club/hpcjudge/bridge/configure"
	"github.com/lcpu-club/hpcjudge/common"
	"github.com/lcpu-club/hpcjudge/common/consts"
	"github.com/lcpu-club/hpcjudge/common/runner"
	"github.com/lcpu-club/hpcjudge/discovery"
	discoveryProtocol "github.com/lcpu-club/hpcjudge/discovery/protocol"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/satori/uuid"
)

type Server struct {
	id               uuid.UUID
	discoveryService *discovery.Service
	discovery        *discovery.Client
	configure        *configure.Configure
	cs               *common.CommonServer
	minio            *minio.Client
}

func NewServer(conf *configure.Configure) (*Server, error) {
	srv := new(Server)
	return srv, srv.Init(conf)
}

var ErrNilConfigure = fmt.Errorf("nil configure")

func (s *Server) connectMinIO() error {
	var err error
	s.minio, err = minio.New(s.configure.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.configure.MinIO.Credentials.AccessKey, s.configure.MinIO.Credentials.SecretKey, ""),
		Secure: s.configure.MinIO.SSL,
	})
	if err != nil {
		return err
	}
	log.Println("Connected to MinIO Server")
	return nil
}

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
	err := s.connectMinIO()
	if err != nil {
		return err
	}
	s.discoveryService = discovery.NewService(context.Background(), s.configure.Discovery.Address, s.configure.Discovery.AccessKey)
	s.discovery = discovery.NewClient(s.configure.Discovery.Address, s.configure.Discovery.AccessKey, s.configure.Discovery.Timeout)
	err = s.discoveryService.Connect()
	if err != nil {
		return err
	}
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
	log.Println("Connected to Discovery Server")
	s.cs = common.NewCommonServer(s.configure.Listen, []byte(s.configure.SecretKey))
	s.registerRoutes(s.cs.GetMux())
	return nil
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/fetch-object", s.HandleFetchObject)
	mux.HandleFunc("/calculate-path", s.HandleCalculatePath)
	mux.HandleFunc("/remove-file", s.HandleRemoveFile)
	mux.HandleFunc("/upload-file", s.HandleUploadFile)
	mux.HandleFunc("/upload-file-presigned", s.HandleUploadFilePresigned)
	mux.HandleFunc("/execute-command", s.HandleExecuteCommand)
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

// mkdirAll with owner copied from os package
func mkdirAll(path string, perm os.FileMode, ownerUid int, ownerGid int) error {
	// Fast path: if we can tell whether path is a directory or file, stop with success or error.
	dir, err := os.Stat(path)
	if err == nil {
		if dir.IsDir() {
			return nil
		}
		return &os.PathError{Op: "mkdir", Path: path, Err: syscall.ENOTDIR}
	}

	// Slow path: make sure parent exists and then call Mkdir for path.
	i := len(path)
	for i > 0 && os.IsPathSeparator(path[i-1]) { // Skip trailing path separator.
		i--
	}

	j := i
	for j > 0 && !os.IsPathSeparator(path[j-1]) { // Scan backward over element.
		j--
	}

	if j > 1 {
		// Create parent.
		err = mkdirAll(path[:j-1], perm, ownerUid, ownerGid)
		if err != nil {
			return err
		}
	}

	// Parent now exists; invoke Mkdir and use its result.
	err = os.Mkdir(path, perm)
	if err != nil {
		// Handle arguments like "foo/." by
		// double-checking that directory doesn't exist.
		dir, err1 := os.Lstat(path)
		if err1 == nil && dir.IsDir() {
			return nil
		}
		return err
	}
	err = os.Chown(path, ownerUid, ownerGid)
	if err != nil {
		return err
	}
	return nil
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
	var uid, gid int
	if req.Owner != "" {
		u, err := user.Lookup(req.Owner)
		if err != nil {
			log.Println("ERROR:", err)
			resp.SetError(api.ErrFailedToLookupUser)
			s.cs.Respond(w, resp)
			return
		}
		uid, err = strconv.Atoi(u.Uid)
		if err != nil {
			log.Println("ERROR:", err)
			resp.SetError(api.ErrFailedToLookupUser)
			s.cs.Respond(w, resp)
			return
		}
		gid, err = strconv.Atoi(u.Gid)
		if err != nil {
			log.Println("ERROR:", err)
			resp.SetError(api.ErrFailedToLookupUser)
			s.cs.Respond(w, resp)
			return
		}
	}
	if req.Owner != "" {
		err = mkdirAll(filepath.Dir(targetPath), req.FileMode.Perm(), uid, gid)
	} else {
		err = os.MkdirAll(filepath.Dir(targetPath), req.FileMode.Perm())
	}
	if err != nil {
		log.Println("ERROR:", err)
		resp.SetError(api.ErrMakeDirectoryError)
		s.cs.Respond(w, resp)
		return
	}
	target, err := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, req.FileMode)
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
		resp.SetError(api.ErrFailedToWriteFile)
		s.cs.Respond(w, resp)
		return
	}
	if req.Owner != "" {
		err = os.Chown(targetPath, uid, gid)
		if err != nil {
			log.Println("ERROR:", err)
			resp.SetError(api.ErrFailedToChangeFilePermission)
		}
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

func (s *Server) HandleRemoveFile(w http.ResponseWriter, r *http.Request) {
	req := new(api.RemoveFileRequest)
	if !s.cs.ParseRequest(w, r, req) {
		return
	}
	resp := &api.RemoveFileResponse{
		ResponseBase: common.ResponseBase{
			Success: true,
		},
	}
	path, err := s.getStoragePath(req.Path.Partition, req.Path.Path)
	if err != nil {
		resp.SetError(err)
		s.cs.Respond(w, resp)
		return
	}
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			resp.SetError(api.ErrFileNotFound)
			s.cs.Respond(w, resp)
			return
		}
		log.Println("ERROR:", err)
		resp.SetError(api.ErrFailedToStatFile)
		s.cs.Respond(w, resp)
		return
	}
	if fi.IsDir() {
		err := os.RemoveAll(path)
		if err != nil {
			log.Println("ERROR:", err)
			resp.SetError(api.ErrFailedToRemove)
		}
	} else {
		err := os.Remove(path)
		if err != nil {
			log.Println("ERROR:", err)
			resp.SetError(api.ErrFailedToRemove)
		}
	}
	s.cs.Respond(w, resp)
}

func (s *Server) HandleExecuteCommand(w http.ResponseWriter, r *http.Request) {
	req := new(api.ExecuteCommandRequest)
	if !s.cs.ParseRequest(w, r, req) {
		return
	}
	resp := &api.ExecuteCommandResponse{
		ResponseBase: common.ResponseBase{
			Success: true,
		},
	}
	wd, err := s.getStoragePath(req.WorkDirectory.Partition, req.WorkDirectory.Path)
	if err != nil {
		resp.SetError(err)
		s.cs.Respond(w, resp)
		return
	}
	// Fix home directory not created
	if req.WorkDirectory.Partition == "home" {
		func() {
			s := strings.Split(wd, "/")
			if len(s) < 2 {
				return
			}
			uName := s[1]
			u, err := user.Lookup(uName)
			if err != nil {
				return
			}
			uid, err := strconv.Atoi(u.Uid)
			if err != nil {
				return
			}
			gid, err := strconv.Atoi(u.Gid)
			if err != nil {
				return
			}
			_, err = os.Stat(wd)
			if os.IsNotExist(err) {
				err = os.Mkdir(wd, os.FileMode(0700))
				if err != nil {
					return
				}
				os.Chown(wd, uid, gid)
			}
		}()
	}
	cmd := exec.Command(req.Command, req.Arguments...)
	currEnv := os.Environ()
	cmd.Env = append(currEnv, req.Environment...)
	cmd, err = runner.CommandUseUser(cmd, req.User)
	if err != nil {
		log.Println("ERROR:", err)
		resp.SetError(api.ErrFailedToLookupUser)
		s.cs.Respond(w, resp)
		return
	}
	cmd.Dir = wd
	pipeStdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("ERROR:", err)
		resp.SetError(api.ErrFailedToCreateCommandPipe)
		s.cs.Respond(w, resp)
		return
	}
	pipeStderr, err := cmd.StderrPipe()
	if err != nil {
		log.Println("ERROR:", err)
		resp.SetError(api.ErrFailedToCreateCommandPipe)
		s.cs.Respond(w, resp)
		return
	}
	err = cmd.Start()
	if err != nil {
		log.Println("ERROR:", err)
		resp.SetError(api.ErrFailedToStartCommand)
		s.cs.Respond(w, resp)
		return
	}
	wait := func() {
		stdout, err := io.ReadAll(pipeStdout)
		if err != nil {
			log.Println("ERROR:", err)
			resp.SetError(api.ErrFailedToReadFromPipe)
			return
		}
		stderr, err := io.ReadAll(pipeStderr)
		if err != nil {
			log.Println("ERROR:", err)
			resp.SetError(api.ErrFailedToReadFromPipe)
			return
		}
		cmd.Wait()
		resp.StdOut = string(stdout)
		resp.StdErr = string(stderr)
		resp.ExitStatus = cmd.ProcessState.ExitCode()
	}
	if !req.Async {
		wait()
	} else {
		go func() {
			wait()
			if req.ReportURL != "" {
				b, err := json.Marshal(resp)
				if err != nil {
					log.Println("ERROR:", err)
					return
				}
				req, err := http.NewRequest("PUT", req.ReportURL, bytes.NewReader(b))
				if err != nil {
					log.Println("ERROR:", err)
					return
				}
				client := &http.Client{}
				_, err = client.Do(req)
				if err != nil {
					log.Println("ERROR:", err)
					return
				}
			}
		}()
	}
	s.cs.Respond(w, resp)
}

func (s *Server) HandleUploadFile(w http.ResponseWriter, r *http.Request) {
	req := new(api.UploadFileRequest)
	if !s.cs.ParseRequest(w, r, req) {
		return
	}
	resp := &api.UploadFileResponse{
		ResponseBase: common.ResponseBase{
			Success: true,
		},
	}
	path, err := s.getStoragePath(req.Path.Partition, req.Path.Path)
	if err != nil {
		resp.SetError(err)
		s.cs.Respond(w, resp)
		return
	}
	bucket := ""
	switch req.Bucket {
	case api.BucketProblem:
		bucket = s.configure.MinIO.Buckets.Problem
	case api.BucketSolution:
		bucket = s.configure.MinIO.Buckets.Solution
	default:
		resp.SetError(api.ErrInvalidBucketType)
		s.cs.Respond(w, resp)
		return
	}
	_, err = s.minio.FPutObject(context.Background(), bucket, req.Object, path, minio.PutObjectOptions{})
	if err != nil {
		log.Println("ERROR:", err)
		resp.SetError(api.ErrUploadFileError)
		s.cs.Respond(w, resp)
		return
	}
	s.cs.Respond(w, resp)
}

func (s *Server) HandleUploadFilePresigned(w http.ResponseWriter, r *http.Request) {
	req := new(api.UploadFilePresignedRequest)
	if !s.cs.ParseRequest(w, r, req) {
		return
	}
	resp := &api.UploadFilePresignedResponse{
		ResponseBase: common.ResponseBase{
			Success: true,
		},
	}
	path, err := s.getStoragePath(req.Path.Partition, req.Path.Path)
	if err != nil {
		resp.SetError(err)
		s.cs.Respond(w, resp)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		log.Println("ERROR:", err)
		resp.SetError(api.ErrFileNotFound)
		s.cs.Respond(w, resp)
		return
	}
	defer f.Close()
	hr, err := http.NewRequest("PUT", req.PresignedURL, f)
	if err != nil {
		log.Println("ERROR:", err)
		resp.SetError(api.ErrUploadFileError)
		s.cs.Respond(w, resp)
		return
	}
	hc := &http.Client{}
	_, err = hc.Do(hr)
	if err != nil {
		log.Println("ERROR:", err)
		resp.SetError(api.ErrUploadFileError)
		s.cs.Respond(w, resp)
		return
	}
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
