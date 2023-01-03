package bridge

import (
	"os"

	"github.com/lcpu-club/hpcjudge/bridge/api"
	"github.com/lcpu-club/hpcjudge/common"
)

type Client struct {
	cc common.CommonClient
}

func NewClient(cc common.CommonClient) *Client {
	c := &Client{
		cc: cc,
	}
	return c
}

func (c *Client) FetchObject(
	objectURL string, targetPartition string, targetPath string, owner string, fileMode os.FileMode,
) error {
	req := &api.FetchObjectRequest{
		ObjectURL: objectURL,
		Path: &api.PartitionedPath{
			Partition: targetPartition,
			Path:      targetPath,
		},
		Owner:    owner,
		FileMode: fileMode,
	}
	resp := new(api.FetchObjectResponse)
	err := c.cc.DoPostRequest("fetch-object", req, resp)
	if err != nil {
		return err
	}
	return resp.GetError()
}

func (c *Client) CalculatePath(partition string, path string) (string, error) {
	req := api.CalculatePathRequest{
		Path: &api.PartitionedPath{
			Partition: partition,
			Path:      path,
		},
	}
	resp := new(api.CalculatePathResponse)
	err := c.cc.DoPostRequest("calculate-path", req, resp)
	if err != nil {
		return "", err
	}
	return resp.Path, resp.GetError()
}

func (c *Client) RemoveFile(partition string, path string) error {
	req := api.RemoveFileRequest{
		Path: &api.PartitionedPath{
			Partition: partition,
			Path:      path,
		},
	}
	resp := new(api.RemoveFileResponse)
	err := c.cc.DoPostRequest("remove-file", req, resp)
	if err != nil {
		return err
	}
	return resp.GetError()
}

func (c *Client) UploadFile(partition string, path string, bucket api.BucketType, object string) error {
	req := &api.UploadFileRequest{
		Path: &api.PartitionedPath{
			Partition: partition,
			Path:      path,
		},
		Bucket: bucket,
		Object: object,
	}
	resp := new(api.RemoveFileResponse)
	err := c.cc.DoPostRequest("upload-file", req, resp)
	if err != nil {
		return err
	}
	return resp.GetError()
}

func (c *Client) ExecuteCommand(
	command string, arguments []string,
	workPartition string, workPath string,
	user string, environment []string,
) (*api.ExecuteCommandResponse, error) {
	req := &api.ExecuteCommandRequest{
		Command:   command,
		Arguments: arguments,
		WorkDirectory: &api.PartitionedPath{
			Partition: workPartition,
			Path:      workPath,
		},
		User:  user,
		Async: false,
	}
	resp := new(api.ExecuteCommandResponse)
	err := c.cc.DoPostRequest("execute-command", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, resp.GetError()
}

func (c *Client) ExecuteCommandAsync(
	command string, arguments []string,
	workPartition string, workPath string,
	user string, environment []string,
	reportURL string,
) error {
	req := &api.ExecuteCommandRequest{
		Command:   command,
		Arguments: arguments,
		WorkDirectory: &api.PartitionedPath{
			Partition: workPartition,
			Path:      workPath,
		},
		User:      user,
		Async:     true,
		ReportURL: reportURL,
	}
	resp := new(api.ExecuteCommandResponse)
	err := c.cc.DoPostRequest("execute-command", req, resp)
	if err != nil {
		return err
	}
	return resp.GetError()
}
