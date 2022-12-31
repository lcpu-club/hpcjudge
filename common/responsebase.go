package common

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding"
	"fmt"
)

type Request interface{}

type Response interface {
	IsSuccess() bool
	GetError() error
	SetError(e error)
}

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

func SignMessage(message []byte, secretKey []byte) ([]byte, error) {
	h := hmac.New(sha256.New, secretKey)
	h.Write(bytes.Trim(message, " \r\n\t"))
	return h.(encoding.TextMarshaler).MarshalText()
}

func CheckSignedMessage(message []byte, secretKey []byte, signature []byte) (bool, error) {
	b, err := SignMessage(message, secretKey)
	if err != nil {
		return false, err
	}
	return hmac.Equal(b, signature), nil
}
