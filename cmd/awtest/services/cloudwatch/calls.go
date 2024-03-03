package cloudwatch

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

var CloudwatchCalls = []types.AWSService{
	{
		Name: "cloudwatch:DescribeAlarms",
		Call: func(sess *session.Session) (interface{}, error) {
			var allAlarms []*cloudwatch.MetricAlarm
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := cloudwatch.New(sess)
				output, err := svc.DescribeAlarms(&cloudwatch.DescribeAlarmsInput{})
				if err != nil {
					return nil, err
				}
				allAlarms = append(allAlarms, output.MetricAlarms...)
			}
			return allAlarms, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "cloudwatch:DescribeAlarms", err)
			}
			if alarms, ok := output.([]*cloudwatch.MetricAlarm); ok {
				for _, alarm := range alarms {
					utils.PrintResult(debug, "", "cloudwatch:DescribeAlarms", fmt.Sprintf("CloudWatch alarm: %s", utils.ColorizeItem(*alarm.AlarmName)), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "cloudwatchlogs:DescribeLogGroups",
		Call: func(sess *session.Session) (interface{}, error) {
			originalConfig := sess.Config
			var allLogGroups []*cloudwatchlogs.LogGroup
			for _, region := range types.Regions {
				regionConfig := &aws.Config{
					Region:      aws.String(region),
					Credentials: originalConfig.Credentials,
				}
				regionSess, err := session.NewSession(regionConfig)
				if err != nil {
					return nil, err
				}
				svc := cloudwatchlogs.New(regionSess)
				input := &cloudwatchlogs.DescribeLogGroupsInput{}
				err = svc.DescribeLogGroupsPages(input, func(output *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
					allLogGroups = append(allLogGroups, output.LogGroups...)
					return true // continue paging
				})
				if err != nil {
					return nil, err
				}
			}
			return allLogGroups, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "cloudwatchlogs:DescribeLogGroups", err)
			}
			if logGroups, ok := output.([]*cloudwatchlogs.LogGroup); ok {
				for _, logGroup := range logGroups {
					utils.PrintResult(debug, "", "cloudwatchlogs:DescribeLogGroups", fmt.Sprintf("Found Log Group: %s", *logGroup.LogGroupName), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "cloudwatchlogs:ListMetrics",
		Call: func(sess *session.Session) (interface{}, error) {
			var allMetrics []*cloudwatch.Metric
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := cloudwatch.New(sess)
				input := &cloudwatch.ListMetricsInput{}
				output, err := svc.ListMetrics(input)
				if err != nil {
					return nil, err
				}
				allMetrics = append(allMetrics, output.Metrics...)
			}
			return allMetrics, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "cloudwatchlogs:ListMetrics", err)
			}
			if metrics, ok := output.([]*cloudwatch.Metric); ok {
				for _, metric := range metrics {
					utils.PrintResult(debug, "", "cloudwatchlogs:ListMetrics", fmt.Sprintf("Found Metric: %s", *metric.MetricName), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
