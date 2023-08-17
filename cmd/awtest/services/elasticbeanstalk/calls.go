package elasticbeanstalk

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
)

var ElasticBeanstalkCalls = []types.AWSService{
	{
		Name: "elasticbeanstalk:DescribeApplications",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := elasticbeanstalk.New(sess)
			input := &elasticbeanstalk.DescribeApplicationsInput{}
			output, err := svc.DescribeApplications(input)
			if err != nil {
				return nil, err
			}
			return output.Applications, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "elasticbeanstalk:DescribeApplications", err)
			}
			if applications, ok := output.([]*elasticbeanstalk.ApplicationDescription); ok {
				for _, app := range applications {
					utils.PrintResult(debug, "", "elasticbeanstalk:DescribeApplications", fmt.Sprintf("Found Application: %s", *app.ApplicationName), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "elasticbeanstalk:DescribeEvents",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := elasticbeanstalk.New(sess)
			input := &elasticbeanstalk.DescribeEventsInput{}
			output, err := svc.DescribeEvents(input)
			if err != nil {
				return nil, err
			}
			return output.Events, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "elasticbeanstalk:DescribeEvents", err)
			}
			if events, ok := output.([]*elasticbeanstalk.EventDescription); ok {
				for _, event := range events {
					utils.PrintResult(debug, "", "elasticbeanstalk:DescribeEvents", fmt.Sprintf("Found Event: %s", *event.Message), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
