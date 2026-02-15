package internal

import (
	"github.com/jakottelaar/relay-microservices/services/users/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioClient(cfg *config.Config) (minioClient *minio.Client, err error) {
	
	minioClient, err = minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(cfg.Minio.AccessKeyID, cfg.Minio.SecretAccessKey, ""),
		Secure: cfg.Minio.UseSSL,
	})

	if err != nil {
		return nil, err
	}

	return minioClient, nil
}
