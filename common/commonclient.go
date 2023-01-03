package common

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/lcpu-club/hpcjudge/discovery"
	discoveryProtocol "github.com/lcpu-club/hpcjudge/discovery/protocol"
)

type CommonClient interface {
	DoPostRequest(endpoint string, data interface{}, resp Response) error
}

type CommonSignedClient struct {
	hc        *http.Client
	address   string
	secretKey []byte
	timeout   time.Duration
}

func NewCommonSignedClient(
	address string, secretKey []byte, timeout time.Duration,
) *CommonSignedClient {
	return &CommonSignedClient{
		hc: &http.Client{
			Timeout: timeout,
		},
		secretKey: secretKey,
		address:   address,
		timeout:   timeout,
	}
}

func (cc *CommonSignedClient) createRequest(
	method string, endpoint string, body io.Reader,
) (*http.Request, error) {
	req, err := http.NewRequest(method, cc.address+"/"+endpoint, body)
	if err != nil {
		return req, err
	}
	return req, nil
}

func (cc *CommonSignedClient) createPostRequestWithJSON(
	endpoint string, data interface{},
) (*http.Request, error) {
	j, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := cc.createRequest("POST", endpoint, bytes.NewReader(j))
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

func (cc *CommonSignedClient) DoPostRequest(
	endpoint string, data interface{}, resp Response,
) error {
	req, err := cc.createPostRequestWithJSON(endpoint, data)
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

type CommonDiscoveredSignedClient struct {
	hc              *http.Client
	discoveryClient *discovery.Client
	condition       *discoveryProtocol.QueryParameters
	secretKey       []byte
	timeout         time.Duration
}

func NewCommonDiscoveredSignedClient(
	dc *discovery.Client,
	cond *discoveryProtocol.QueryParameters,
	secretKey []byte,
	timeout time.Duration,
) *CommonDiscoveredSignedClient {
	hc := &http.Client{
		Timeout: timeout,
	}
	cc := &CommonDiscoveredSignedClient{
		hc:              hc,
		discoveryClient: dc,
		condition:       cond,
		secretKey:       secretKey,
		timeout:         timeout,
	}
	return cc
}

func (cc *CommonDiscoveredSignedClient) DoPostRequest(
	endpoint string, data interface{}, resp Response,
) error {
	svc, err := cc.discoveryClient.Query(cc.condition)
	if err != nil {
		return err
	}
	csc := &CommonSignedClient{
		hc:        cc.hc,
		address:   svc.Address,
		secretKey: cc.secretKey,
		timeout:   cc.timeout,
	}
	return csc.DoPostRequest(endpoint, data, resp)
}

type CommonSignedMultiAddressClient struct {
	hc        *http.Client
	address   []string
	secretKey []byte
	timeout   time.Duration
}

func NewCommonSignedMultiAddressClient(
	address []string, secretKey []byte, timeout time.Duration, shuffleAddress bool,
) *CommonSignedMultiAddressClient {
	addressShuffled := address
	if shuffleAddress {
		rand.Shuffle(len(address), func(i, j int) {
			addressShuffled[i], addressShuffled[j] = addressShuffled[j], addressShuffled[i]
		})
	}
	return &CommonSignedMultiAddressClient{
		hc: &http.Client{
			Timeout: timeout,
		},
		secretKey: secretKey,
		address:   addressShuffled,
		timeout:   timeout,
	}
}

func (cc *CommonSignedMultiAddressClient) DoPostRequest(
	endpoint string, data interface{}, resp Response,
) error {
	var err error
	for _, address := range cc.address {
		csc := &CommonSignedClient{
			hc:        cc.hc,
			address:   address,
			secretKey: cc.secretKey,
			timeout:   cc.timeout,
		}
		err = csc.DoPostRequest(endpoint, data, resp)
		if err == nil {
			return nil
		}
	}
	return err
}
