package protocol

import (
	"fmt"

	"github.com/satori/uuid"
)

type Service struct {
	ID      uuid.UUID `json:"id"`
	Address string    `json:"address"`
	Type    string    `json:"type"`
	Tags    []string  `json:"tags"`
}

type QueryParameters struct {
	ID          uuid.UUID `json:"id,omitempty"`
	Address     string    `json:"address,omitempty"`
	Type        string    `json:"type,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	ExcludeTags []string  `json:"exclude-tags,omitempty"`
}

type ResponseBase struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (rb *ResponseBase) IsSuccess() bool {
	return rb.Success
}

func (rb *ResponseBase) GetError() error {
	if rb.IsSuccess() {
		return nil
	}
	return fmt.Errorf(rb.Error)
}

type Response interface {
	IsSuccess() bool
	GetError() error
}

type ListRequest struct {
	Condition *QueryParameters `json:"condition"`
}

type ListResponse struct {
	ResponseBase
	Data []*Service `json:"data"`
}

type QueryRequest struct {
	Condition *QueryParameters `json:"condition"`
}

type QueryResponse struct {
	ResponseBase
	Data *Service `json:"data"`
}

type AddRequest struct {
	Service *Service `json:"service"`
}

type AddResponse struct {
	ResponseBase
	Service *Service `json:"service"`
}

type DeleteRequest struct {
	Service *Service `json:"service"`
}

type DeleteResponse struct {
	ResponseBase
}

type ListPeersResponse struct {
	ResponseBase
	Data []string `json:"data"`
}

type RemovePeerRequest struct {
	Peer string `json:"peer"`
}

type RemovePeerResponse struct {
	ResponseBase
}
