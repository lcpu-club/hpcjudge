package problem

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/pelletier/go-toml/v2"
)

func GetProblemMeta(ctx context.Context, mc *minio.Client, bucket string, problemID string) (*Problem, error) {
	obj, err := mc.GetObject(ctx, bucket, problemID+"/problem.toml", minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	pToml, err := io.ReadAll(obj)
	if err != nil {
		return nil, err
	}
	rslt := new(Problem)
	err = toml.Unmarshal(pToml, rslt)
	if err != nil {
		return nil, err
	}
	return rslt, nil
}
