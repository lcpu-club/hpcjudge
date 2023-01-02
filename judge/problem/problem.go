package problem

type Problem struct {
	ID          string              `json:"id" yaml:"id" toml:"id"`
	Name        string              `json:"name" yaml:"name" toml:"name"`
	Environment *ProblemEnvironment `json:"environment" yaml:"environment" toml:"environment"`
	Entrance    *ProblemEntrance    `json:"entrance" yaml:"entrance" toml:"entrance"`
}

type ProblemEntrance struct {
	Command string `json:"command" yaml:"command" toml:"command"` // Prior to script
	Script  string `json:"script" yaml:"script" toml:"script"`
}

type ProblemEnvironment struct {
	Tags              []string                             `json:"tags" yaml:"tags" toml:"tags"`
	ExcludeTags       []string                             `json:"exclude-tags" yaml:"exclude-tags" toml:"exclude-tags"`
	ScriptLimits      *ProblemEnvironmentScriptLimits      `json:"script-limits" yaml:"script-limits" toml:"script-limits"`
	EstimatedResource *ProblemEnvironmentEstimatedResource `json:"estimated-resource" yaml:"estimated-resource" toml:"estimated-resource"`
}

type ProblemEnvironmentScriptLimits struct {
	CPU    int `json:"cpu" yaml:"cpu" toml:"cpu"`          // in percentage, 200 for 2 CPUs, ...
	Memory int `json:"memory" yaml:"memory" toml:"memory"` // in MB
}

type ProblemEnvironmentEstimatedResource struct {
	Nodes   int `json:"nodes" yaml:"nodes" toml:"nodes"`
	Cores   int `json:"cores" yaml:"cores" toml:"cores"`
	Memory  int `json:"memory" yaml:"memory" toml:"memory"`    // in MB
	Storage int `json:"storage" yaml:"storage" toml:"storage"` // in MB
	GPU     int `json:"gpu" yaml:"gpu" toml:"gpu"`
}
