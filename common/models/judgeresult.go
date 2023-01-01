package models

type JudgeResult struct {
	Done            bool                  `json:"done"`
	Score           int                   `json:"score"`
	Message         string                `json:"message"`
	DetailedMessage string                `json:"detailed-message"`
	Subtasks        []*JudgeSubtaskResult `json:"subtasks"`
}

type JudgeSubtaskResult struct {
	Score           int    `json:"score"`
	Message         string `json:"message"`
	DetailedMessage string `json:"detailed-message"`
}
