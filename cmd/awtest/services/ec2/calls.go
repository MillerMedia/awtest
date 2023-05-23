package ec2

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var EC2Calls = []types.AWSService{
	{
		Name: "ec2:DescribeInstances",
		Call: func(sess *session.Session) (interface{}, error) {
			var allInstances []*ec2.Instance
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := ec2.New(sess)
				input := &ec2.DescribeInstancesInput{}
				output, err := svc.DescribeInstances(input)
				if err != nil {
					return nil, err
				}
				for _, reservation := range output.Reservations {
					allInstances = append(allInstances, reservation.Instances...)
				}
			}
			return allInstances, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "ec2:DescribeInstances", err)
			}
			if instances, ok := output.([]*ec2.Instance); ok {
				for _, instance := range instances {
					utils.PrintResult(debug, "", "ec2:DescribeInstances", fmt.Sprintf("Found EC2 instance: %s", *instance.InstanceId), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
