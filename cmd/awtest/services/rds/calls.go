package rds

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

var RDSCalls = []types.AWSService{
	{
		Name: "rds:DescribeDBInstances",
		Call: func(sess *session.Session) (interface{}, error) {
			var allDBInstances []*rds.DBInstance
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := rds.New(sess)
				output, err := svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{})
				if err != nil {
					return nil, err
				}
				allDBInstances = append(allDBInstances, output.DBInstances...)
			}
			return allDBInstances, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "rds:DescribeDBInstances", err)
			}
			if dbInstances, ok := output.([]*rds.DBInstance); ok {
				for _, db := range dbInstances {
					utils.PrintResult(debug, "", "rds:DescribeDBInstances", fmt.Sprintf("Found RDS instance: %s (%s)", *db.DBInstanceIdentifier, *db.AvailabilityZone), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
