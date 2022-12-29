package discovery

import (
	"time"

	"github.com/lcpu-club/hpcjudge/discovery/protocol"
)

type Client struct {
	cc *commonClient
}

func NewClient(address []string, accessKey string, timeout time.Duration) *Client {
	return &Client{
		cc: newCommonClient(address, accessKey, timeout),
	}
}

func (c *Client) List(cond *protocol.QueryParameters) ([]*protocol.Service, error) {
	resp := &protocol.ListResponse{
		Data: []*protocol.Service{},
	}
	err := c.cc.doPostRequest("list", &protocol.ListRequest{
		Condition: cond,
	}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) Query(cond *protocol.QueryParameters) (*protocol.Service, error) {
	resp := &protocol.QueryResponse{
		Data: &protocol.Service{},
	}
	err := c.cc.doPostRequest("query", &protocol.QueryRequest{
		Condition: cond,
	}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) Add(s *protocol.Service) (*protocol.Service, error) {
	resp := &protocol.AddResponse{
		Service: &protocol.Service{},
	}
	err := c.cc.doPostRequest("add", &protocol.AddRequest{
		Service: s,
	}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Service, nil
}

func (c *Client) Delete(s *protocol.Service) error {
	resp := &protocol.AddResponse{
		Service: &protocol.Service{},
	}
	return c.cc.doPostRequest("delete", &protocol.DeleteRequest{
		Service: s,
	}, resp)
}
