package models

type RunJudgeScriptData struct {
	ProblemID       string           `json:"problem-id"`
	SolutionID      string           `json:"solution-id"`
	Username        string           `json:"username"`
	ResourceControl *ResourceControl `json:"resource-control"`
}
