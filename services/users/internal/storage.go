package internal

import (
	"github.com/jakottelaar/relay-microservices/services/users/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewStorageClient(cfg *config.Config) (*minio.Client, error) {
    client, err := minio.New(cfg.Storage.Endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(cfg.Storage.AccessKeyID, cfg.Storage.SecretAccessKey, ""),
        Secure: cfg.Storage.UseSSL,
    })
    if err != nil {
        return nil, err
    }
    return client, nil
}