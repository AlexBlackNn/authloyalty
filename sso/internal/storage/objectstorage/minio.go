package objectstorage

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"strings"

	"github.com/AlexBlackNn/authloyalty/sso/internal/config"
	"github.com/AlexBlackNn/authloyalty/sso/internal/dto"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const contentType = "application/octet-stream"

// Client interacts with minio
type Client struct {
	client *minio.Client
	cfg    *config.Config
}

// New creates minio client
func New(cfg *config.Config) (*Client, error) {
	minioClient, err := minio.New(cfg.Minio.URL, &minio.Options{
		Creds: credentials.NewStaticV4(
			cfg.Minio.AccessKeyID,
			cfg.Minio.SecretAccessKey,
			"",
		),
		//Secure: cfg.Minio.Secure,
		Secure:     false,
		MaxRetries: 15,
	})
	if err != nil {
		return nil, err
	}
	return &Client{client: minioClient, cfg: cfg}, nil
}

// UploadData uploads data to minio
func (c *Client) UploadData(ctx context.Context, register *dto.Register) (string, error) {
	fileName := uuid.New().String()
	// might be better to use decoder separately (add it in constructor), but for now it seems overengineering
	data, err := base64.StdEncoding.DecodeString(strings.Split(register.Avatar, "|")[1])
	if err != nil {
		return "", err
	}
	_, err = c.client.PutObject(
		ctx,
		c.cfg.Minio.BucketName,
		fileName,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

// RemoveObject removes data from minio
func (c *Client) RemoveObject(ctx context.Context, fileName string) error {
	err := c.client.RemoveObject(
		ctx,
		c.cfg.Minio.BucketName,
		fileName,
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		return err
	}
	return nil
}

// DownloadData downloads data to minio
func (c *Client) DownloadData(ctx context.Context, userInfo *dto.UserInfo) ([]byte, error) {
	object, err := c.client.GetObject(
		ctx,
		c.cfg.Minio.BucketName,
		userInfo.FileName,
		minio.GetObjectOptions{},
	)
	defer object.Close()

	if err != nil {
		return nil, err
	}

	return io.ReadAll(object)
}
