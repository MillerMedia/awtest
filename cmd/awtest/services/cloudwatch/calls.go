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

type LogGroupWithStreams struct {
	LogGroup   *cloudwatchlogs.LogGroup
	LogStreams []*cloudwatchlogs.LogStream
}

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
		Name: "cloudwatchlogs:DescribeLogGroupsAndStreams",
		Call: func(sess *session.Session) (interface{}, error) {
			var allLogGroupsWithStreams []*LogGroupWithStreams

			for _, region := range types.Regions {
				regionConfig := &aws.Config{
					Region:      aws.String(region),
					Credentials: sess.Config.Credentials,
				}
				regionSess, err := session.NewSession(regionConfig)
				if err != nil {
					return nil, err
				}
				svc := cloudwatchlogs.New(regionSess)

				// Describe Log Groups
				input := &cloudwatchlogs.DescribeLogGroupsInput{}
				err = svc.DescribeLogGroupsPages(input, func(output *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
					for _, logGroup := range output.LogGroups {
						// Describe Log Streams for each Log Group
						streamInput := &cloudwatchlogs.DescribeLogStreamsInput{
							LogGroupName: logGroup.LogGroupName,
						}
						var logStreams []*cloudwatchlogs.LogStream
						err := svc.DescribeLogStreamsPages(streamInput, func(streamOutput *cloudwatchlogs.DescribeLogStreamsOutput, lastPage bool) bool {
							logStreams = append(logStreams, streamOutput.LogStreams...)
							return true // continue paging
						})
						if err != nil {
							return false
						}

						allLogGroupsWithStreams = append(allLogGroupsWithStreams, &LogGroupWithStreams{
							LogGroup:   logGroup,
							LogStreams: logStreams,
						})
					}
					return true // continue paging
				})
				if err != nil {
					return nil, err
				}
			}

			return allLogGroupsWithStreams, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "cloudwatchlogs:DescribeLogGroupsAndStreams", err)
			}
			if logGroupsWithStreams, ok := output.([]*LogGroupWithStreams); ok {
				if len(logGroupsWithStreams) == 0 {
					utils.PrintResult(debug, "", "cloudwatchlogs:DescribeLogGroupsAndStreams", "No log groups found.", nil)
				} else {
					for _, lgws := range logGroupsWithStreams {
						// Print the Log Group
						utils.PrintResult(debug, "", "cloudwatchlogs:DescribeLogGroupsAndStreams", fmt.Sprintf("Found Log Group: %s", *lgws.LogGroup.LogGroupName), nil)

						// Check if there are any Log Streams
						if len(lgws.LogStreams) > 0 {
							utils.PrintResult(debug, "", "cloudwatchlogs:DescribeLogGroupsAndStreams", fmt.Sprintf("  Log Streams for %s:", *lgws.LogGroup.LogGroupName), nil)
							for _, logStream := range lgws.LogStreams {
								utils.PrintResult(debug, "", "cloudwatchlogs:DescribeLogGroupsAndStreams", fmt.Sprintf("    - Log Stream: %s", *logStream.LogStreamName), nil)
							}
						} else {
							utils.PrintResult(debug, "", "cloudwatchlogs:DescribeLogGroupsAndStreams", "  No log streams found or access denied.", nil)
						}
					}
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
