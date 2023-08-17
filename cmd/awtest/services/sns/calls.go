package sns

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"strings"
)

type TopicWithAttributes struct {
	Topic      *sns.Topic
	Attributes map[string]*string
}

var SNSCalls = []types.AWSService{
	{
		Name: "sns:ListTopics",
		Call: func(sess *session.Session) (interface{}, error) {
			var allTopicsWithAttributes []TopicWithAttributes
			originalConfig := sess.Config
			for _, region := range types.Regions {
				regionConfig := &aws.Config{
					Region:      aws.String(region),
					Credentials: originalConfig.Credentials,
				}
				regionSess, err := session.NewSession(regionConfig)
				if err != nil {
					return nil, err
				}
				svc := sns.New(regionSess)
				output, err := svc.ListTopics(&sns.ListTopicsInput{})
				if err != nil {
					return nil, err
				}

				for _, topic := range output.Topics {
					arnParts := strings.Split(*topic.TopicArn, ":")
					if len(arnParts) < 4 {
						return nil, fmt.Errorf("invalid ARN: %s", *topic.TopicArn)
					}
					region := arnParts[3]
					regionConfigForAttr := &aws.Config{
						Region:      aws.String(region),
						Credentials: originalConfig.Credentials,
					}
					attrSess, err := session.NewSession(regionConfigForAttr)
					if err != nil {
						return nil, err
					}
					svc = sns.New(attrSess)

					attrOutput, attrErr := svc.GetTopicAttributes(&sns.GetTopicAttributesInput{
						TopicArn: topic.TopicArn,
					})
					if attrErr != nil {
						return nil, attrErr
					}

					allTopicsWithAttributes = append(allTopicsWithAttributes, TopicWithAttributes{Topic: topic, Attributes: attrOutput.Attributes})
				}
			}
			return allTopicsWithAttributes, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "sns:ListTopics", err)
			}
			if topics, ok := output.([]TopicWithAttributes); ok {
				for _, topicWithAttr := range topics {
					colorizedTopic := utils.ColorizeItem(*topicWithAttr.Topic.TopicArn)
					utils.PrintResult(debug, "", "sns:ListTopics", fmt.Sprintf("SNS Topic: %s", colorizedTopic), nil)

					// Print attributes
					for name, value := range topicWithAttr.Attributes {
						// Only display DisplayName and number of subscriptions
						if (name == "DisplayName" || name == "SubscriptionsConfirmed" || name == "SubscriptionsPending" || name == "SubscriptionsDeleted") && *value != "" {
							utils.PrintResult(debug, "", "sns:GetTopicAttributes", fmt.Sprintf("SNS Topic: %s | %s = %s", colorizedTopic, name, *value), nil)
						}
					}
					// print blank line
					fmt.Println()
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
