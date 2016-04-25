package deployer

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/bernos/cfn-deploy/cfndeploy/uploader"
	"log"
	"os"
	"path"
	"path/filepath"
)

// Deployer is and interface that can deploy a cloudformation
// stack
type Deployer interface {
	Deploy(*DeployOptions) error
}

// deployer implements the Deployer interface
type deployer struct {
	svc cloudformationiface.CloudFormationAPI
	u   uploader.Uploader
}

// New creates a new Deployer instance
func New(cw cloudformationiface.CloudFormationAPI, u uploader.Uploader) Deployer {
	return &deployer{
		svc: cw,
		u:   u,
	}
}

// Deploys our stack
func (d *deployer) Deploy(options *DeployOptions) error {
	folder := options.TemplateFolder
	bucket := options.Bucket
	helper := &cloudFormationHelper{d.svc}
	mainTemplate := path.Join(options.TemplateFolder, options.MainTemplate)

	if options.StackParams == nil {
		options.StackParams = StackParams(make(map[string]string))
	}

	log.Printf("Checking if stack already exists")
	exists, err := helper.StackExists(options.StackName)

	if err != nil {
		return err
	}

	templates, err := findTemplates(folder)

	if err != nil {
		return err
	}

	log.Printf("Validating templates")
	validationError := helper.ValidateTemplates(templates)

	if validationError != nil {
		return validationError
	}

	version := checksumTemplates(templates)
	prefix := path.Join(
		calculateBucketPrefix(options.BucketFolder, options.StackName, version),
		"templates")

	log.Printf("Uploading templates")
	templateURL, err := d.uploadTemplates(templates, folder, mainTemplate, bucket, prefix)

	if err != nil {
		return err
	}

	// TODO: this is wrong. Needs to be folder of templateURL
	options.StackParams["TemplateBaseUrl"] = templateURL
	options.StackParams["Version"] = version

	if exists {
		log.Printf("Updating stack")
		return d.update(templateURL, options.StackParams, options.StackTags)
	}

	log.Printf("Creating stack")
	return d.create(templateURL, options.StackParams, options.StackTags)
}

func (d *deployer) create(templateURL string, params StackParams, tags StackTags) error {
	log.Printf("create(%s, %v, %v)", templateURL, params, tags)
	return fmt.Errorf("Not implemented")
}

func (d *deployer) update(templateURL string, params StackParams, tags StackTags) error {
	return fmt.Errorf("Not implemented")
}

// uploadTemplates uploads all templates, and returns the URL
// of the main template
func (d *deployer) uploadTemplates(templates []string, basePath, mainTemplate, bucket, prefix string) (string, error) {
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

func checksumTemplates(files []string) string {
	// TODO: implement this
	return ""
}

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

func calculateBucketPrefix(stackName, bucketFolder, version string) string {
	return path.Join(bucketFolder, stackName, version)
}
