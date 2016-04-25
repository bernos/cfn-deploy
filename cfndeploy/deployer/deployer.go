package deployer

import (
	"crypto/sha1"
	"fmt"
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
	folder := options.TemplateFolder
	bucket := options.Bucket
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

	templates, err := findTemplates(folder)

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
	templateURL, err := d.uploadTemplates(templates, mainTemplate, bucket, prefix)

	if err != nil {
		log.Printf("Error uploading templates: %s", err.Error())
		return err
	}

	options.StackParams["TemplateBaseUrl"] = templateURL
	options.StackParams["Version"] = version

	if exists {
		log.Printf("Updating stack")
		return d.update(templateURL, options.StackParams, options.StackTags)
	}

	log.Printf("Creating stack")
	return d.create(templateURL, options.StackParams, options.StackTags)
}

// create creates a cloudforamtion stack
func (d *deployer) create(templateURL string, params StackParams, tags StackTags) error {
	log.Printf("create(%s, %v, %v)", templateURL, params, tags)
	return fmt.Errorf("Not implemented")
}

// update updates a cloudformation stack
func (d *deployer) update(templateURL string, params StackParams, tags StackTags) error {
	return fmt.Errorf("Not implemented")
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
				return baseURL(result.URL), nil
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
