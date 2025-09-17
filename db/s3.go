package db

import (
	"bytes"
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pagefaultgames/rogueserver/defs"
)

func (s *store) GetSystemSaveFromS3(uuid []byte) (defs.SystemSaveData, error) {
	var system defs.SystemSaveData

	username, err := Store.FetchUsernameFromUUID(uuid)
	if err != nil {
		return system, err
	}

	resp, err := s3client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_SYSTEM_BUCKET_NAME")),
		Key:    aws.String(username),
	})
	if err != nil {
		return system, err
	}

	err = json.NewDecoder(resp.Body).Decode(&system)
	if err != nil {
		return system, err
	}

	return system, nil
}

func (s *store) StoreSystemSaveDataS3(uuid []byte, data defs.SystemSaveData) error {
	username, err := s.FetchUsernameFromUUID(uuid)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	err = json.NewEncoder(buf).Encode(data)
	if err != nil {
		return err
	}

	_, err = s3client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("S3_SYSTEM_BUCKET_NAME")),
		Key:    aws.String(username),
		Body:   buf,
	})
	if err != nil {
		return err
	}

	return nil
}
