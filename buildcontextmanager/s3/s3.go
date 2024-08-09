package s3buildmanager

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3 struct {
	config map[string]string
}

func New(config map[string]string) *s3 {
	return &s3{
		config: config,
	}
}

const aws_access_key_id = "AWS_ACCESS_KEY_ID"
const aws_secret_access_key = "AWS_SECERET_ACCESS_KEY"
const aws_region = "AWS_REGION"
const bucket = "bucket"

func (s *s3) Upload(buf *bytes.Buffer, key string) (string, error) {
	fmt.Println(s.config)
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(s.config[aws_region]),
		Credentials: credentials.NewStaticCredentials(s.config[aws_access_key_id], s.config[aws_secret_access_key], ""),
	})
	if err != nil {
		return "", fmt.Errorf("error creating aws session: %s", err.Error())
	}

	uploader := s3manager.NewUploader(session)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.config[bucket]),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("s3://%s/%s", s.config[bucket], key), nil
}
