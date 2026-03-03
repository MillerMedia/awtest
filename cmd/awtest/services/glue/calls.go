package glue

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glue"
)

var GlueCalls = []types.AWSService{
	{
		Name: "glue:ListWorkflows",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := glue.New(sess)
			input := &glue.ListWorkflowsInput{}
			return svc.ListWorkflows(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "glue:ListWorkflows", err)
			}
			if workflows, ok := output.(*glue.ListWorkflowsOutput); ok {
				for _, workflowName := range workflows.Workflows {
					utils.PrintResult(debug, "", "glue:ListWorkflows", fmt.Sprintf("Workflow: %s", workflowName), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "glue:ListJobs",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := glue.New(sess)
			input := &glue.ListJobsInput{}
			return svc.ListJobs(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "glue:ListJobs", err)
			}
			if jobs, ok := output.(*glue.ListJobsOutput); ok {
				for _, job := range jobs.JobNames {
					utils.PrintResult(debug, "", "glue:ListJobs", fmt.Sprintf("Job: %s", *job), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
