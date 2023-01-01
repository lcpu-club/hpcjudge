package api

import "github.com/lcpu-club/hpcjudge/common"

type PartitionedPath struct {
	Partition string `json:"partition"`
	Path      string `json:"path"`
}

type FetchObjectRequest struct {
	ObjectURL string           `json:"object-url"`
	Path      *PartitionedPath `json:"path"`
}

type FetchObjectResponse struct {
	common.ResponseBase
}

type CalculatePathRequest struct {
	Path *PartitionedPath `json:"path"`
}

type CalculatePathResponse struct {
	common.ResponseBase
	Path string `json:"path"`
}

type ExecuteCommandRequest struct {
	Command       string           `json:"command"`
	Arguments     []string         `json:"arguments"`
	WorkDirectory *PartitionedPath `json:"work-directory"`
	User          string           `json:"user"`
	Async         bool             `json:"async"`
	ReportURL     string           `json:"report-url"` // Used with async
}

type ExecuteCommandResponse struct {
	common.ResponseBase
	ExitStatus int    `json:"exit-status"`
	StdOut     string `json:"stdout"`
	StdErr     string `json:"stderr"`
}

type RemoveFileRequest struct {
	Path *PartitionedPath `json:"path"`
}

type RemoveFileResponse struct {
	common.ResponseBase
}

type UploadFileRequest struct {
	Path   *PartitionedPath `json:"path"`
	Bucket string           `json:"bucket"`
	Object string           `json:"object"`
}

type UploadFileResponse struct {
	common.ResponseBase
}
