package sqs

import (
	"fmt"
	"github.com/MillerMedia/AWTest/cmd/awtest/types"
	"github.com/MillerMedia/AWTest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var SQSCalls = []types.AWSService{
	{
		Name: "sqs:ListQueues",
		Call: func(sess *session.Session) (interface{}, error) {
			var allQueues []*string
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := sqs.New(sess)
				output, err := svc.ListQueues(&sqs.ListQueuesInput{})
				if err != nil {
					return nil, err
				}
				allQueues = append(allQueues, output.QueueUrls...)
			}
			return allQueues, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "sqs:ListQueues", err)
			}
			if queues, ok := output.([]*string); ok {
				for _, queue := range queues {
					utils.PrintResult(debug, "", "sqs:ListQueues", fmt.Sprintf("Found SQS queue: %s", *queue), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
