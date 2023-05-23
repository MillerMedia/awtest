package sns

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

var SNSCalls = []types.AWSService{
	{
		Name: "sns:ListTopics",
		Call: func(sess *session.Session) (interface{}, error) {
			var allTopics []*sns.Topic
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := sns.New(sess)
				output, err := svc.ListTopics(&sns.ListTopicsInput{})
				if err != nil {
					return nil, err
				}
				allTopics = append(allTopics, output.Topics...)
			}
			return allTopics, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "sns:ListTopics", err)
			}
			if topics, ok := output.([]*sns.Topic); ok {
				for _, topic := range topics {
					utils.PrintResult(debug, "", "sns:ListTopics", fmt.Sprintf("Found SNS topic: %s", *topic.TopicArn), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
