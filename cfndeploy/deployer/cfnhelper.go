package deployer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"io/ioutil"
	"sync"
)

type cloudFormationHelper struct {
	svc cloudformationiface.CloudFormationAPI
}

// StackExists returns true if there is a single stack with the given
// name. It will return an error if more than one stack with the given
// name was found, or if there were any errors returned by the cloudformation
// API
func (c cloudFormationHelper) StackExists(name string) (bool, error) {
	params := &cloudformation.ListStacksInput{
		StackStatusFilter: []*string{
			aws.String(cloudformation.StackStatusCreateComplete),
			aws.String(cloudformation.StackStatusCreateFailed),
			aws.String(cloudformation.StackStatusCreateInProgress),
			aws.String(cloudformation.StackStatusRollbackComplete),
			aws.String(cloudformation.StackStatusRollbackFailed),
			aws.String(cloudformation.StackStatusRollbackInProgress),
			aws.String(cloudformation.StackStatusUpdateComplete),
			aws.String(cloudformation.StackStatusUpdateCompleteCleanupInProgress),
			aws.String(cloudformation.StackStatusUpdateInProgress),
			aws.String(cloudformation.StackStatusUpdateRollbackComplete),
			aws.String(cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress),
			aws.String(cloudformation.StackStatusUpdateRollbackFailed),
			aws.String(cloudformation.StackStatusUpdateRollbackInProgress),
		},
	}

	found := false

	err := c.svc.ListStacksPages(params, func(page *cloudformation.ListStacksOutput, lastPage bool) bool {
		for _, stack := range page.StackSummaries {
			if *stack.StackName == name {
				found = true
				return false
			}
		}
		return true
	})

	return found, err
}

func (c cloudFormationHelper) ValidateTemplates(files []string) error {
	var (
		errors []error
		wg     sync.WaitGroup
		out    = make(chan error)
	)

	defer close(out)

	for _, file := range files {
		wg.Add(1)

		go func(file string) {
			out <- c.ValidateTemplate(file)
		}(file)
	}

	go func() {
		for err := range out {
			if err != nil {
				errors = append(errors, err)
			}
			wg.Done()
		}
	}()

	wg.Wait()

	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

func (c cloudFormationHelper) ValidateTemplate(file string) error {
	buf, err := ioutil.ReadFile(file)

	if err != nil {
		return err
	}

	body := string(buf)

	params := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(body),
	}

	_, err = c.svc.ValidateTemplate(params)

	return err
}
