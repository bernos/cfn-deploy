package deployer

import (
	"crypto/sha1"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/bernos/cfn-deploy/cfndeploy/uploader"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Deployer is and interface that can deploy a cloudformation stack
type Deployer interface {
	Deploy(*DeployOptions) error
}

// deployer implements the Deployer interface
type deployer struct {
	svc    cloudformationiface.CloudFormationAPI
	helper *cloudFormationHelper
	u      uploader.Uploader
}

// New creates a new Deployer instance
func New(c cloudformationiface.CloudFormationAPI, u uploader.Uploader) Deployer {
	return &deployer{
		svc:    c,
		u:      u,
		helper: &cloudFormationHelper{c},
	}
}

// Deploy detploys a cloudformation stack. If the stack does not exist it will
// be created, otherwise the existing stack will be updated.
func (d *deployer) Deploy(options *DeployOptions) error {
	mainTemplate := path.Join(options.TemplateFolder, options.MainTemplate)

	if options.StackParams == nil {
		options.StackParams = StackParams(make(map[string]string))
	}

	if options.StackTags == nil {
		options.StackTags = StackTags(make(map[string]string))
	}

	log.Printf("Checking if stack already exists")
	exists, err := d.helper.StackExists(options.StackName)

	if err != nil {
		return err
	}

	templates, err := findTemplates(options.TemplateFolder)

	if err != nil {
		return err
	}

	version, err := checksumTemplates(templates)

	if err != nil {
		return err
	}

	prefix := path.Join(
		calculateBucketPrefix(options.StackName, options.BucketFolder, version),
		"templates")

	log.Printf("Uploading templates")
	templateURL, err := d.uploadTemplates(templates, mainTemplate, options.Bucket, prefix)

	if err != nil {
		return err
	}

	params := d.buildStackParams(version, templateURL, options.StackParams)

	var (
		stackID       string
		desiredStatus string
	)

	if exists {
		stackID, err = d.update(options.StackName, templateURL, params, options.StackTags)
		desiredStatus = cloudformation.StackStatusUpdateComplete
	} else {
		stackID, err = d.create(options.StackName, templateURL, params, options.StackTags)
		desiredStatus = cloudformation.StackStatusCreateComplete
	}

	if err == nil {
		err = d.helper.WaitForStack(stackID, desiredStatus)
	}

	return err
}

// create creates a cloudforamtion stack
func (d *deployer) create(stackName, templateURL string, params StackParams, tags StackTags) (string, error) {
	if resp, err := d.svc.CreateStack(d.buildCreateStackInput(stackName, templateURL, params, tags)); err == nil {
		return *resp.StackId, nil
	} else {
		return "", err
	}
}

// buildCreateStackInput builds up the CreateStackInput struct
func (d *deployer) buildCreateStackInput(stackName, templateURL string, params StackParams, tags StackTags) *cloudformation.CreateStackInput {
	createStackInput := &cloudformation.CreateStackInput{
		StackName:   aws.String(stackName),
		Parameters:  params.AWSParams(),
		Tags:        tags.AWSTags(),
		TemplateURL: aws.String(templateURL),
	}

	return createStackInput
}

// buildStackParams builds up StackParams, including version number and template base URL
func (d *deployer) buildStackParams(version, mainTemplateURL string, params StackParams) StackParams {
	params["Version"] = version
	params["TemplateBaseUrl"] = baseURL(mainTemplateURL)

	return params
}

// update updates a cloudformation stack
func (d *deployer) update(stackName, templateURL string, params StackParams, tags StackTags) (string, error) {
	if resp, err := d.svc.UpdateStack(d.buildUpdateStackInput(stackName, templateURL, params, tags)); err == nil {
		return *resp.StackId, nil
	} else {
		return "", err
	}
}

// buildUpdateStackInput builds up the CreateStackInput struct
func (d *deployer) buildUpdateStackInput(stackName, templateURL string, params StackParams, tags StackTags) *cloudformation.UpdateStackInput {
	updateStackInput := &cloudformation.UpdateStackInput{
		StackName:   aws.String(stackName),
		Parameters:  params.AWSParams(),
		Tags:        tags.AWSTags(),
		TemplateURL: aws.String(templateURL),
	}

	return updateStackInput
}

// uploadTemplates uploads all templates to S3, and returns the base URL path of
// the main template
func (d *deployer) uploadTemplates(templates []string, mainTemplate, bucket, prefix string) (string, error) {
	log.Printf("Validating templates")

	if err := d.helper.ValidateTemplates(templates); err != nil {
		return "", err
	}

	basePath := filepath.Dir(mainTemplate)

	if results, err := d.u.UploadFiles(templates, basePath, bucket, prefix); err == nil {
		for _, result := range results {
			if result.File == mainTemplate {
				return result.URL, nil
			}
		}
		return "", fmt.Errorf("Unable to find url of main template")
	} else {
		return "", err
	}
}

func checksumTemplates(files []string) (string, error) {
	hash := sha1.New()

	for _, file := range files {
		data, err := ioutil.ReadFile(file)

		if err != nil {
			return "", err
		}

		_, werr := hash.Write(data)

		if werr != nil {
			return "", werr
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil))[:8], nil
}

// findTemplates recursively walks the provided dir and returns a slice of
// filenames of all files
func findTemplates(dir string) ([]string, error) {
	var files []string

	return files, filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
}

// calculateBucketPrefix returns the appropriate bucket key prefix for the given
// stackName, bucketFolder and version
func calculateBucketPrefix(stackName, bucketFolder, version string) string {
	return path.Join(bucketFolder, stackName, version)
}

// baseURL returns all but the file part of the given url
func baseURL(url string) string {
	return url[:strings.LastIndex(url, "/")+1]
}
