package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
)

const DefaultModuleName = "AWTest"

type AWSService struct {
	Name       string
	Call       func(*session.Session) (interface{}, error)
	Process    func(interface{}, bool)
	ModuleName string
}

var AWSListCalls = []AWSService{
	{
		Name: "s3:ListBuckets",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := s3.New(sess)
			output, err := svc.ListBuckets(&s3.ListBucketsInput{})
			return output, err
		},
		Process: func(output interface{}, debug bool) {
			if s3Output, ok := output.(*s3.ListBucketsOutput); ok {
				for _, bucket := range s3Output.Buckets {
					printResult(debug, "", "s3:ListBuckets", fmt.Sprintf("s3://%s", *bucket.Name), nil)
				}
			}
		},
	},
	{
		Name: "ec2:DescribeInstances",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := ec2.New(sess)
			input := &ec2.DescribeInstancesInput{}
			output, err := svc.DescribeInstances(input)
			return output, err
		},
		Process: func(output interface{}, debug bool) {
			if ec2Output, ok := output.(*ec2.DescribeInstancesOutput); ok {
				for _, reservation := range ec2Output.Reservations {
					for _, instance := range reservation.Instances {
						printResult(debug, "", "ec2:DescribeInstances", fmt.Sprintf("Found instance: %s", *instance.InstanceId), nil)
					}
				}
			}
		},
	},
	// ... Add other service calls as required ...
}
