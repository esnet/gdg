package test

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
	"net/url"
)

//Minio URL Implementation of blob.Bucket intended purely for testing.
//Ensure that AWS_ACCESS_KEY and AWS_SECRET_KEY are both set correctly.

const (
	Scheme     = "minio"
	MINIO_HOST = "MINIO_HOST"
	MINIO_SSL  = "MINIO_SSL_ENABLED"
)

type URLOpener struct{}

func init() {
	blob.DefaultURLMux().RegisterBucket(Scheme, &URLOpener{})
}

func (o *URLOpener) OpenBucketURL(ctx context.Context, u *url.URL) (*blob.Bucket, error) {
	minioHost := "http://localhost:9000"
	useSSL := false
	if val := ctx.Value(MINIO_HOST); val != nil {
		minioHost = val.(string)
	}
	if val := ctx.Value(MINIO_SSL); val != nil {
		useSSL = val.(bool)
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewEnvCredentials(),
		Endpoint:         aws.String(minioHost),
		DisableSSL:       aws.Bool(!useSSL),
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String("us-east-1"),
	})
	if err != nil {
		return nil, fmt.Errorf("open bucket %v: %v", u, err)
	}
	bucketName := u.Host

	//Create testing bucket
	func() {
		client := s3.New(sess)
		m := s3.CreateBucketInput{
			Bucket: &bucketName,
		}
		//attempt to create bucket
		_, err := client.CreateBucket(&m)
		if err != nil {
			log.Warn("testing bucket already exists or cannot be created")
		}
	}()

	return s3blob.OpenBucket(context.Background(), sess, bucketName, nil)
}
