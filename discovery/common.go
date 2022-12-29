package discovery

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/lcpu-club/hpcjudge/discovery/protocol"
)

type commonClient struct {
	hc        *http.Client
	address   []string
	accessKey string
	timeout   time.Duration
}

func newCommonClient(address []string, accessKey string, timeout time.Duration) *commonClient {
	cc := &commonClient{
		hc: &http.Client{
			Timeout: timeout,
		},
		address:   address,
		accessKey: accessKey,
		timeout:   timeout,
	}
	return cc
}

func (cc *commonClient) createRequest(n int, method string, endpoint string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, cc.address[n]+"/"+endpoint, body)
	if err != nil {
		return req, err
	}
	req.Header.Add("X-Access-Key", cc.accessKey)
	return req, nil
}

func (cc *commonClient) createPostRequestWithJSON(n int, endpoint string, data interface{}) (*http.Request, error) {
	j, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return cc.createRequest(n, "POST", endpoint, bytes.NewReader(j))
}

func (cc *commonClient) doPostRequestN(n int, endpoint string, data interface{}, resp protocol.Response) error {
	req, err := cc.createPostRequestWithJSON(n, endpoint, data)
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

func (cc *commonClient) doPostRequest(endpoint string, data interface{}, resp protocol.Response) error {
	var err error
	for k := range cc.address {
		err = cc.doPostRequestN(k, endpoint, data, resp)
		if err == nil {
			return resp.GetError()
		}
	}
	return err
}
