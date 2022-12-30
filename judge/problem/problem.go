package problem

type Problem struct {
	ID          string              `json:"id" yaml:"id" toml:"id"`
	Name        string              `json:"name" yaml:"name" toml:"name"`
	Environment *ProblemEnvironment `json:"environment" yaml:"environment" toml:"environment"`
	JudgeScript string              `json:"judge-script" yaml:"judge-script" toml:"judge-script"`
}

type ProblemEnvironment struct {
	Tags        []string `json:"tags" yaml:"tags" toml:"tags"`
	ExcludeTags []string `json:"exclude-tags" yaml:"exclude-tags" toml:"exclude-tags"`
}
