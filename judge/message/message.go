package message

type JudgeMessage struct {
	ProblemID    string `json:"problem_id"`
	SubmissionID string `json:"submission_id"`
	RunnerArgs   string `json:"runner_args"`
}

type JudgeReportMessage struct {
	SubmissionID string `json:"submission_id"`
	Success      bool   `json:"success"`
	Error        string `json:"error"`
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
)
