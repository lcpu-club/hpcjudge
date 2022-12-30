package common

import "fmt"

type ResponseBase struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
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

func (rb *ResponseBase) SetError(e error) {
	if e != nil {
		rb.Success = false
		rb.Error = e.Error()
	} else {
		rb.Success = true
	}
}
