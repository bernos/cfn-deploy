package deployer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bernos/cfn-deploy/cfndeploy/uploader"
	"testing"
)

var (
	defaultBucket       = "seek-candidate-cfn-templates-test"
	defaultPrefix       = "deployer-test"
	defaultBucketFolder = "bucket-folder"
	defaultStackName    = "test-stack"
)

func TestDeploy(t *testing.T) {
	sess := session.New(&aws.Config{Region: aws.String("ap-southeast-2")})
	s3 := s3manager.NewUploader(sess)
	cw := cloudformation.New(sess)
	u := uploader.New(s3)
	d := New(cw, u)

	o := &DeployOptions{
		Bucket:         defaultBucket,
		BucketFolder:   defaultBucketFolder,
		StackName:      defaultStackName,
		TemplateFolder: "./test-fixtures/templates/valid",
		MainTemplate:   "Stack.json",
	}

	err := d.Deploy(o)

	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
}

func TestCalculateBucketPrefix(t *testing.T) {
	tests := []struct {
		bucketFolder string
		stackName    string
		version      string
		want         string
	}{
		{"foo", "stack", "123", "foo/stack/123"},
		{"", "stack", "123", "stack/123"},
	}

	for _, tt := range tests {
		got := calculateBucketPrefix(tt.stackName, tt.bucketFolder, tt.version)

		if got != tt.want {
			t.Errorf("Want %s, got %s", tt.want, got)
		}
	}
}

func TestBaseURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"http://example.com/foo/bar/baz.txt", "http://example.com/foo/bar/"},
		{"http://example.com/foo.txt", "http://example.com/"},
	}

	for _, tt := range tests {
		got := baseURL(tt.url)

		if got != tt.want {
			t.Errorf("Want %s, got %s", tt.want, got)
		}
	}
}

func TestChecksumTemplates(t *testing.T) {
	files, _ := findTemplates("./test-fixtures/templates/valid")
	// files = append(files, "bogman")
	sum, _ := checksumTemplates(files)
	t.Error(sum)
}
