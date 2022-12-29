package message

type JobMessage struct {
	ProblemID    string `json:"problem_id"`
	SubmissionID string `json:"submission_id"`
	RunnerArgs   string `json:"runner_args"`
}

type ReportMessage struct {
	SubmissionID string `json:"submission_id"`
	Done         bool   `json:"done"`
	Score        int    `json:"score"` // Set to 0 if not done
	Timestamp    int64  `json:"timestamp"`
}
