package uploader

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"testing"
)

var (
	defaultBucket = "seek-candidate-cfn-templates-test"
)

func testBuildUploader() Uploader {
	sess := session.New(&aws.Config{Region: aws.String("ap-southeast-2")})
	s3 := s3manager.NewUploader(sess)
	return New(s3)
}

func TestBuildUploadList(t *testing.T) {
	u := &uploader{}

	list, _ := u.buildUploadList("../")

	t.Errorf("%s", list)
}

func TestUploadFile(t *testing.T) {
	u := testBuildUploader()
	r := u.UploadFile("./test-fixtures/one.txt", defaultBucket, "/tests/test-fixtures/one.txt")

	t.Errorf("%#v", r)
}

func TestUploadFiles(t *testing.T) {
	u := testBuildUploader()
	files := []string{
		"./test-fixtures/one.txt",
	}

	r, _ := u.UploadFiles(files, defaultBucket, "key-prefix")

	for _, rs := range r {
		t.Errorf("%#v\n", rs)
	}
}
