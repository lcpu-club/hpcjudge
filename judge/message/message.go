package message

import "github.com/satori/uuid"

type JudgeMessage struct {
	ProblemID    string `json:"problem_id"`
	SubmissionID string `json:"submission_id"`
	RunnerArgs   string `json:"runner_args"`
}

type JudgeReportMessage struct {
	SubmissionID string `json:"submission_id"`
	Done         bool   `json:"done"`
	Score        int    `json:"score"` // Set to 0 if not done
	Message      string `json:"message"`
	Timestamp    int64  `json:"timestamp"` // time.Now().UnixMicro()
	// We don't use time.Now().UnixNano() as it exceeds js's Number.MAX_SAFE_INTEGER
}

type SandboxOperation string

const (
	SandboxOperationCreate  SandboxOperation = "create"
	SandboxOperationDestroy SandboxOperation = "destroy"
	SandboxOperationStart   SandboxOperation = "start"
	SandboxOperationStop    SandboxOperation = "stop"
	SandboxOperationStatus  SandboxOperation = "status"
	SandboxOperationConnect SandboxOperation = "connect"
)

type SandboxMessage struct {
	ID        uuid.UUID        `json:"id"`         // Unique message ID, should be uuid.V4
	SandboxID string           `json:"sandbox_id"` // Should be a valid unix user name and should ensure uniqueness
	Operation SandboxOperation `json:"operation"`
	Type      string           `json:"type"`
}

type SandboxReportMessage struct {
	ID        uuid.UUID        `json:"id"` // Being uuid.Nil for broadcast / notify
	SandboxID string           `json:"sandbox_id"`
	Operation SandboxOperation `json:"operation"`
	Type      string           `json:"type"`
	Address   string           `json:"address"`   // Expose code-server address, when Operation == SandboxOperationStart
	StreamID  uuid.UUID        `json:"stream_id"` // Shell stream ID in hpc-bridge, when Operation == SandboxOperationConnect
}
