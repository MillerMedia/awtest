package cloudwatch

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
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
					utils.PrintResult(debug, "", "cloudwatch:DescribeAlarms", fmt.Sprintf("Found CloudWatch alarm: %s", *alarm.AlarmName), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
