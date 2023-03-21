package integration_tests

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
	"testing"
)

func TestFoobar(t *testing.T) {
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials("test", "secretsss", ""),
		Endpoint:         aws.String("http://localhost:9000"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String("us-east-1"),
	})

	cloudUrl := "s3://testing"
	bucket, err := s3blob.OpenBucket(context.Background(), sess, "testing", nil)
	assert.Nil(t, err)
	i := bucket.List(nil)
	for {
		next, err := i.Next(context.Background())
		if err != nil {
			break
		}
		log.Info(next.Key)
	}
	_ = bucket
	bucketObj, err := blob.OpenBucket(context.Background(), cloudUrl)
	_, _ = bucketObj, err
}
