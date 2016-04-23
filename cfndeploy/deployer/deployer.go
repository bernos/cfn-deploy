package deployer

import (
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

// Deployer is and interface that can deploy a cloudformation
// stack
type Deployer interface {
	Deploy(*DeployOptions) error
}

// deployer implements the Deployer interface
type deployer struct {
	svc cloudformationiface.CloudFormationAPI
}

// New creates a new Deployer instance
func New(cw cloudformationiface.CloudFormationAPI) Deployer {
	return &deployer{
		svc: cw,
	}
}

// Deploys our stack
func (d *deployer) Deploy(options *DeployOptions) error {
	// upload templates to S3

	//

	return nil
}
