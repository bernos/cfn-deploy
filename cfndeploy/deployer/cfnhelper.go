package deployer

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"io/ioutil"
	"regexp"
	"sync"
	"time"
)

var (
	inProgressRegexp = regexp.MustCompile(".+_IN_PROGRESS$")
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

func (c cloudFormationHelper) WaitForStack(stackID, desiredState string) error {
	params := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackID),
	}

	start := time.Now().UTC()
	timeout := time.Second * 60 * 20

	for {
		resp, err := c.svc.DescribeStacks(params)

		if err != nil {
			return err
		}

		if len(resp.Stacks) == 0 {
			return fmt.Errorf("Stack with ID %s not found", stackID)
		}

		if len(resp.Stacks) != 1 {
			return fmt.Errorf("Ambiguous stack ID %s", stackID)
		}

		status := *resp.Stacks[0].StackStatus

		if status == desiredState {
			return nil
		}

		if !inProgressRegexp.MatchString(status) {
			return fmt.Errorf("Unexpected stack status. Wanted %s, but got %s", desiredState, status)
		}

		if time.Since(start) > timeout {
			return fmt.Errorf("Stack %s failed to reach state %s within %s", stackID, desiredState, timeout)
		}

		time.Sleep(time.Second * 5)
	}
}

func (c cloudFormationHelper) LogStackEvents(stackID string, logger func(*cloudformation.StackEvent, error)) (cancel func()) {
	done := make(chan struct{})
	ticker := time.NewTicker(time.Second * 5)

	params := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackID),
	}

	go func() {
		defer ticker.Stop()
		var lastEventID string

		for {
			resp, err := c.svc.DescribeStackEvents(params)

			if err != nil {
				logger(nil, err)
			} else {
				if len(resp.StackEvents) > 0 {
					newEvents := resp.StackEvents[:1]

					if lastEventID != "" {
						newEvents = resp.StackEvents

						for i, event := range resp.StackEvents {
							if *event.EventId == lastEventID {
								newEvents = resp.StackEvents[:i]
								break
							}
						}
					}

					for i := len(newEvents) - 1; i >= 0; i-- {
						logger(newEvents[i], nil)
						lastEventID = *newEvents[i].EventId
					}
				}
			}

			select {
			case <-done:
				return
			case <-ticker.C:
			}
		}
	}()

	return func() {
		close(done)
	}
}
