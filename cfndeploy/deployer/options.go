package deployer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// DeployOptions holds options for template deployment
type DeployOptions struct {
	StackName      string
	TemplateFolder string
	MainTemplate   string
	Region         string
	Bucket         string
	BucketFolder   string
	StackParams    StackParams
	StackTags      StackTags
}

// Validate returns an error if the options are not valid
func (o *DeployOptions) Validate() error {
	// TODO: implement me
	return nil
}

// StackParams holds parameters for a cloudforamtion stack
type StackParams map[string]string

// AWSParams converts StackParams to a slice of *cloudformation.Parameter
func (s StackParams) AWSParams() []*cloudformation.Parameter {
	var params []*cloudformation.Parameter

	for k, v := range s {
		params = append(params, &cloudformation.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: aws.String(v),
		})
	}

	return params
}

// StackTags holds tags for a cloudformation stack
type StackTags map[string]string

// AWSTags converts StackTags to slice of *cloudformation.Tag
func (t StackTags) AWSTags() []*cloudformation.Tag {
	var tags []*cloudformation.Tag

	for k, v := range t {
		tags = append(tags, &cloudformation.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	return tags
}
