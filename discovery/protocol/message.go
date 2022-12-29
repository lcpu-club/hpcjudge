package protocol

type RegisterOperationType string

const (
	RegisterOperationInform RegisterOperationType = "inform"
	RegisterOperationAdd    RegisterOperationType = "add"
	RegisterOperationDelete RegisterOperationType = "delete"
	RegisterOperationHas    RegisterOperationType = "has"
	RegisterOperationNoop   RegisterOperationType = "noop"
)

type ClientRegisterMessage struct {
	Operation RegisterOperationType `json:"operation"`
	Data      *Service              `json:"data"`
}

type ServerRegisterMessage struct {
	ResponseBase
	Has     bool     `json:"has"`
	Service *Service `json:"service"`
}
