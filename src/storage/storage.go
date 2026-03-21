package storage

import (
	"context"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/kisbogdan-kolos/gallery/helper"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var client *minio.Client
var bucket string

func Init() error {
	endpoint := helper.EnvGet("STORAGE_URI", "localhost:8333")
	bucket = helper.EnvGet("STORAGE_BUCKET", "gallery")
	accessKey := helper.EnvGet("STORAGE_ACCESS_KEY", "some_access_key")
	secretKey := helper.EnvGet("STORAGE_SECRET_KEY", "some_secret_key")

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Printf("Failed to connect: %v\n", err)
		return err
	}

	client = minioClient

	bucketExists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		log.Printf("Failed to check bucket exists: %v\n", err)
		return err
	}

	if !bucketExists {
		err = client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Printf("Failed to create bucket: %v\n", err)
			return err
		}
	}

	return nil
}

func Set(id uuid.UUID, contentType string, data io.Reader, size uint) error {
	_, err := client.PutObject(context.Background(), bucket, id.String(), data, int64(size), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Printf("Failed to write data: %v\n", err)
	}
	return err
}

func Get(id uuid.UUID) (data io.Reader, length uint, contentType string, err error) {
	objinfo, err := client.GetObjectACL(context.Background(), bucket, id.String())
	if err != nil {
		log.Printf("Failed to load data: %v\n", err)
		return nil, 0, "", err
	}

	obj, err := client.GetObject(context.Background(), bucket, id.String(), minio.GetObjectOptions{})

	return obj, uint(objinfo.Size), objinfo.ContentType, err
}

func Delete(id uuid.UUID) error {
	err := client.RemoveObject(context.Background(), bucket, id.String(), minio.RemoveObjectOptions{})

	return err
}
