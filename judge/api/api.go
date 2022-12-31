package api

import (
	bridgeApi "github.com/lcpu-club/hpcjudge/bridge/api"
	"github.com/lcpu-club/hpcjudge/common"
)

type ExecuteCommandReportRequest struct {
	bridgeApi.ExecuteCommandReport
}

type ExecuteCommandReportResponse struct {
	common.ResponseBase
}

type ExecuteCommandReportData struct {
	ProblemID  string `json:"problem-id"`
	SolutionID string `json:"solution-id"`
}

type GetSignedUploadURLRequest struct {
	Bucket string `json:"bucket"`
	Path   string `json:"path"`
}

type GetSignedUploadURLResponse struct {
	common.ResponseBase
	URL string `json:"url"`
}

type JudgeReportRequest struct {
	ProblemID       string                `json:"problem-id"`
	SolutionID      string                `json:"solution-id"`
	Done            bool                  `json:"done"`
	Score           int                   `json:"score"`
	Message         string                `json:"message"`
	DetailedMessage string                `json:"detailed-message"` // Not reported using nsq
	Subtasks        []*JudgeReportRequest `json:"subtasks"`         // Not reported using nsq
}

type JudgeReportResponse struct {
	common.ResponseBase
}
