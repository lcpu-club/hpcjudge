package api

// endpoint: /sandbox/{:sandbox_id}?access-key={:access_key}
// each sandbox has an access key

type Command struct {
	Executable string   `json:"executable"`
	Arguments  []string `json:"arguments"`
}

type SandboxOperation string

const (
	SandboxOperationConnect SandboxOperation = `json:"connect"` // Interactive execute
	SandboxOperationExecute SandboxOperation = `json:"execute"` // Non-interactive execute
	SandboxOperationStart   SandboxOperation = `json:"start"`   // Start the sandbox with ports mapped
	SandboxOperationStop    SandboxOperation = `json:"stop"`    // Stop the sandbox
)

type SandboxRequest struct {
}

type SandboxResponse struct {
}
