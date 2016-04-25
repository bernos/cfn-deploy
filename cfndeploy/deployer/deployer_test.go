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

	t.Errorf("Error: %s", err.Error())
}
