package common

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type CommonSignedClient struct {
	hc        *http.Client
	address   string
	secretKey []byte
	timeout   time.Duration
}

func NewCommonSignedClient(address string, secretKey []byte, timeout time.Duration) *CommonSignedClient {
	return &CommonSignedClient{
		hc:        &http.Client{},
		secretKey: secretKey,
		address:   address,
		timeout:   timeout,
	}
}

func (cc *CommonSignedClient) CreateRequest(method string, endpoint string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, cc.address+"/"+endpoint, body)
	if err != nil {
		return req, err
	}
	return req, nil
}

func (cc *CommonSignedClient) CreatePostRequestWithJSON(endpoint string, data interface{}) (*http.Request, error) {
	j, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := cc.CreateRequest("POST", endpoint, bytes.NewReader(j))
	if err != nil {
		return nil, err
	}
	signature, err := SignMessage(j, cc.secretKey)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Signature", string(signature))
	return req, nil
}

func (cc *CommonSignedClient) DoPostRequest(endpoint string, data interface{}, resp Response) error {
	req, err := cc.CreatePostRequestWithJSON(endpoint, data)
	if err != nil {
		return err
	}
	r, err := cc.hc.Do(req)
	if err != nil {
		return err
	}
	j, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(j, resp)
	if err != nil {
		return err
	}
	return nil
}
