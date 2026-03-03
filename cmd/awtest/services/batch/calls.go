package batch

import (
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/batch"
)

var BatchCalls = []types.AWSService{
	{
		Name: "batch:ListJobs",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := batch.New(sess)
			output, err := svc.ListJobs(&batch.ListJobsInput{})
			if err != nil {
				return nil, err
			} else {
				return output, nil
			}
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "batch:ListJobs", err)
			} else {
				if jobs, ok := output.(*batch.ListJobsOutput); ok {
					for _, job := range jobs.JobSummaryList {
						utils.PrintResult(debug, "", "batch:ListJobs", *job.JobName, nil)
					}
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
