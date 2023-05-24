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
					colorizedTopic := utils.ColorizeItem(*topic.TopicArn)
					utils.PrintResult(debug, "", "sns:ListTopics", fmt.Sprintf("SNS Topic: %s", colorizedTopic), nil)

					arnParts := strings.Split(*topic.TopicArn, ":")
					if len(arnParts) < 4 {
						return fmt.Errorf("invalid ARN: %s", *topic.TopicArn)
					}
					region := arnParts[3]
					sessWithRegion := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
					svc := sns.New(sessWithRegion)

					// get topic attributes
					attrOutput, attrErr := svc.GetTopicAttributes(&sns.GetTopicAttributesInput{
						TopicArn: topic.TopicArn,
					})
					if attrErr != nil {
						return utils.HandleAWSError(debug, "sns:GetTopicAttributes", attrErr)
					}
					for name, value := range attrOutput.Attributes {
						// Only display DisplayName and number of subscriptions
						if (name == "DisplayName" || name == "SubscriptionsConfirmed" || name == "SubscriptionsPending" || name == "SubscriptionsDeleted") && *value != "" {
							utils.PrintResult(debug, "", "sns:GetTopicAttributes", fmt.Sprintf("SNS Topic: %s | %s = %s", colorizedTopic, name, *value), nil)
						}
					}

					// get subscriptions
					//subOutput, subErr := svc.ListSubscriptionsByTopic(&sns.ListSubscriptionsByTopicInput{
					//	TopicArn: topic.TopicArn,
					//})
					//if subErr != nil {
					//	return utils.HandleAWSError(debug, "sns:ListSubscriptionsByTopic", subErr)
					//}
					//for _, subscription := range subOutput.Subscriptions {
					//	utils.PrintResult(debug, "", "sns:ListSubscriptionsByTopic", fmt.Sprintf("SNS Topic: %s | Subscription: %s = %s", colorizedTopic, *subscription.Endpoint, *subscription.Protocol), nil)
					//}

					// print blank line
					fmt.Println()
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
