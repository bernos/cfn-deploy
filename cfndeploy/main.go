package main

import (
	"fmt"
	"github.com/bernos/cfn-deploy/cfndeploy/commands"
	"github.com/codegangsta/cli"
	"os"
)

var (
	version string
)

func main() {
	app := cli.NewApp()
	app.Name = "cfndeploy"
	app.Version = version
	app.Usage = "Deploy cfn templates"

	app.Commands = []cli.Command{
		{
			Name:        "deploy",
			ArgsUsage:   "path/to/template/folder",
			Usage:       "Deploy templates",
			Description: "Foobar",
			Action:      commands.Deploy,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "stackname,n",
					Usage:  "Name of the stack to create",
					EnvVar: "CFNDEPLOY_STACKNAME",
				},
				cli.StringFlag{
					Name:   "region,r",
					Usage:  "Region to deploy to",
					EnvVar: "CFNDEPLOY_REGION",
					Value:  "ap-southeast-2",
				},
				cli.StringFlag{
					Name:   "main,m",
					Usage:  "Name of the main cloudforamtion template",
					EnvVar: "CFNDEPLOY_MAIN",
					Value:  "Stack.json",
				},
				cli.StringFlag{
					Name:   "bucket,b",
					Usage:  "Name of the S3 bucket to upload templates to",
					EnvVar: "CFNDEPLOY_BUCKET",
				},
				cli.StringFlag{
					Name:   "bucketfolder,k",
					Usage:  "Optional bucket folder to upload templates to",
					EnvVar: "CFNDEPLOY_BUCKET_FOLDER",
				},
				cli.StringFlag{
					Name:  "params,p",
					Usage: "Stack parameters, in the format ParamOne=ValueOne,Param2=Value2",
				},
				cli.StringFlag{
					Name:  "tags,t",
					Usage: "Stack tag, in the format TagNameOne=TagValueOne,TagNameTwo=TagValueTwo",
				},
			},
		},
	}

	err := app.Run(os.Args)

	if err != nil {
		fmt.Printf("Error! %s", err.Error())
		os.Exit(1)
	}
}
