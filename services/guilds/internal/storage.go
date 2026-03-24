package internal

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/jakottelaar/relay-microservices/services/guilds/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type GuildStorage struct {
    client     *minio.Client
    bucketName string
    baseURL    string
}

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

func NewGuildStorage(client *minio.Client, bucketName string, baseURL string) *GuildStorage {
    return &GuildStorage{
        client:     client,
        bucketName: bucketName,
        baseURL:    baseURL,
    }
}

func (s *GuildStorage) UploadGuildIcon(ctx context.Context, guildID int64, file *multipart.FileHeader) (string, error) {
    src, err := file.Open()
    if err != nil {
        return "", err
    }
    defer src.Close()

    ext := filepath.Ext(file.Filename)
    objectName := fmt.Sprintf("%d/icon%s", guildID, ext)

    _, err = s.client.PutObject(ctx, s.bucketName, objectName, src, file.Size, minio.PutObjectOptions{
        ContentType: file.Header.Get("Content-Type"),
    })
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("%s/%s", s.bucketName, objectName), nil
}

func (s *GuildStorage) EnsureBucket(ctx context.Context) error {
    exists, err := s.client.BucketExists(ctx, s.bucketName)
    if err != nil {
        return err
    }
    if !exists {
        return s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
    }
    return nil
}