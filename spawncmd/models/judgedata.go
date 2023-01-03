package models

type RunJudgeScriptData struct {
	ProblemID          string           `json:"problem-id"`
	SolutionID         string           `json:"solution-id"`
	Username           string           `json:"username"`
	ResourceControl    *ResourceControl `json:"resource-control"`
	Command            string           `json:"command"`
	Script             string           `json:"script"`
	AutoRemoveSolution bool             `json:"auto-remove-solution"`
}
