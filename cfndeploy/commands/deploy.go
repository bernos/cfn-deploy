package commands

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bernos/cfn-deploy/cfndeploy/deployer"
	"github.com/bernos/cfn-deploy/cfndeploy/uploader"
	"github.com/codegangsta/cli"
	"os"
	"strings"
)

func validateRequiredStringParam(param string, c *cli.Context) error {
	if c.String(param) == "" {
		return fmt.Errorf("Missing required '%s' param", param)
	}
	return nil
}

func validateDeployContext(c *cli.Context) error {
	ps := []string{
		"stackname",
		"main",
		"region",
		"bucket",
	}

	for _, p := range ps {
		if err := validateRequiredStringParam(p, c); err != nil {
			return err
		}
	}

	if c.NArg() != 1 {
		return fmt.Errorf("Expected template folder as argument")
	}

	return nil
}

func Deploy(c *cli.Context) {
	if err := validateDeployContext(c); err != nil {
		fmt.Printf("Error! %s\n", err.Error())
		cli.ShowCommandHelp(c, "deploy")
		os.Exit(1)
	}

	params, err := parseMap(c.String("params"))

	if err != nil {
		fmt.Printf("Error! %s", err.Error())
		os.Exit(1)
	}

	tags, err := parseMap(c.String("tags"))

	if err != nil {
		fmt.Printf("Error! %s", err.Error())
		os.Exit(1)
	}

	options := &deployer.DeployOptions{
		StackName:      c.String("stackname"),
		TemplateFolder: c.Args().First(),
		MainTemplate:   c.String("main"),
		Region:         c.String("region"),
		Bucket:         c.String("bucket"),
		BucketFolder:   c.String("bucketfolder"),
		StackParams:    deployer.StackParams(params),
		StackTags:      deployer.StackTags(tags),
	}

	if err := options.Validate(); err != nil {
		fmt.Printf("Error! %s", err.Error())
		os.Exit(1)
	}

	sess := session.New(&aws.Config{Region: aws.String(options.Region)})
	s3 := s3manager.NewUploader(sess)
	cfn := cloudformation.New(sess)
	upl := uploader.New(s3)
	dep := deployer.New(cfn, upl)

	if err := dep.Deploy(options); err != nil {
		fmt.Printf("Error! %s", err.Error())
		os.Exit(1)
	}

	fmt.Printf("Deployment sucessful!\n")
}

func parseMap(s string) (map[string]string, error) {
	m := make(map[string]string)

	if len(s) > 0 {
		pairs := strings.Split(s, ",")

		for _, pair := range pairs {
			kv := strings.Split(pair, "=")

			if len(kv) != 2 {
				return m, fmt.Errorf("Badly formed command line param '%s'. Expected format 'key=value'", pair)
			}

			m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	return m, nil
}
