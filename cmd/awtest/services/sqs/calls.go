package sqs

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/sqs"
	"strings"
)

func arnToQueueUrl(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) != 6 {
		return "" // Return empty string or handle error appropriately
	}
	region := parts[3]
	accountId := parts[4]
	queueName := parts[5]

	return fmt.Sprintf("https://sqs.%s.amazonaws.com/%s/%s", region, accountId, queueName)
}

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
					utils.PrintResult(debug, "", "sqs:ListQueues", fmt.Sprintf("================================================================================"), nil)
					utils.PrintResult(debug, "", "sqs:ListQueues", fmt.Sprintf("Found SQS queue: %s", *queue), nil)

					sess := session.Must(session.NewSession())
					sess.Config.Region = aws.String("us-east-1")
					svc := sqs.New(sess)

					// Receive messages from the queue
					receiveOutput, receiveErr := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
						QueueUrl: queue,
					})

					if receiveErr != nil {
						utils.HandleAWSError(debug, "sqs:ReceiveMessage", receiveErr)
					} else {
						if len(receiveOutput.Messages) == 0 {
							utils.PrintResult(debug, "", "sqs:ReceiveMessage", "Access Granted: No messages in queue", nil)
						}

						for _, message := range receiveOutput.Messages {
							utils.PrintResult(debug, "", "sqs:ReceiveMessage", fmt.Sprintf("Received message: %s", *message.Body), nil)
						}
					}

					attrOutput, attrErr := svc.GetQueueAttributes(&sqs.GetQueueAttributesInput{
						QueueUrl: queue,
						AttributeNames: []*string{
							aws.String("All"),
						},
					})

					if attrErr != nil {
						utils.HandleAWSError(debug, "sqs:GetQueueAttributes", attrErr)
					} else {
						// Format the attributes
						attributes := make(map[string]string)
						for key, value := range attrOutput.Attributes {
							attributes[key] = *value
							utils.PrintResult(debug, "", "sqs:GetQueueAttributes", fmt.Sprintf("Queue attribute: %s = %s", key, *value), nil)

							if key == "Policy" {
								// Get event source ARNs from the policy
								policy := *value
								policyMap := make(map[string]interface{})
								err := utils.UnmarshalJSON([]byte(policy), &policyMap)
								if err != nil {
									utils.PrintResult(debug, "", "sqs:GetQueueAttributes", fmt.Sprintf("Error parsing Policy: %s", err), nil)
								} else {
									if policyMap["Statement"] != nil {
										statements := policyMap["Statement"].([]interface{})
										for _, statement := range statements {
											statementMap := statement.(map[string]interface{})
											if statementMap["Condition"] != nil {
												condition := statementMap["Condition"].(map[string]interface{})
												if condition["ArnEquals"] != nil {
													arns := condition["ArnEquals"].(map[string]interface{})
													for _, arn := range arns {
														utils.PrintResult(debug, "", "sqs:GetQueueAttributes", fmt.Sprintf("Found event source ARN: %s", arn), nil)

														// Parse the ARN and get the event source rule
														ruleName := strings.Split(arn.(string), "/")[1]

														// Attempt describe rule
														eventBridgeSess := session.Must(session.NewSession())
														eventBridgeSess.Config.Region = aws.String("us-west-2")
														eventBridgeSvc := eventbridge.New(eventBridgeSess)
														eventBridgeOutput, eventBridgeErr := eventBridgeSvc.DescribeRule(&eventbridge.DescribeRuleInput{
															Name: aws.String(ruleName),
														})

														if eventBridgeErr != nil {
															utils.HandleAWSError(debug, "eventbridge:DescribeRule", eventBridgeErr)
														} else {
															utils.PrintResult(debug, "", "eventbridge:DescribeRule", fmt.Sprintf("Found event source rule: %s", *eventBridgeOutput.Name), nil)
														}
													}
												}
											}
										}
									}
								}
							}

							if key == "RedrivePolicy" {
								// Get the dead letter queue
								redrivePolicy := *value
								redrivePolicyMap := make(map[string]interface{})
								err := utils.UnmarshalJSON([]byte(redrivePolicy), &redrivePolicyMap)
								if err != nil {
									utils.PrintResult(debug, "", "sqs:GetQueueAttributes", fmt.Sprintf("Error parsing RedrivePolicy: %s", err), nil)
								} else {
									deadLetterQueue := redrivePolicyMap["deadLetterTargetArn"]
									if deadLetterQueue != nil {
										utils.PrintResult(debug, "", "sqs:GetQueueAttributes", fmt.Sprintf("Found Dead letter queue: %s", deadLetterQueue), nil)

										deadLetterQueueArn := deadLetterQueue.(string)
										deadLetterQueueUrl := arnToQueueUrl(deadLetterQueueArn)

										// Receive messages from the dead letter queue
										deadLetterSess := session.Must(session.NewSession())
										deadLetterSess.Config.Region = aws.String("us-east-1")
										deadLetterSvc := sqs.New(deadLetterSess)
										deadLetterAttrOutput, deadLetterAttrErr := deadLetterSvc.ListDeadLetterSourceQueues(&sqs.ListDeadLetterSourceQueuesInput{
											QueueUrl: aws.String(deadLetterQueueUrl),
										})

										if deadLetterAttrErr != nil {
											utils.HandleAWSError(debug, "sqs:ListDeadLetterSourceQueues", deadLetterAttrErr)
										} else {
											if len(deadLetterAttrOutput.QueueUrls) == 0 {
												utils.PrintResult(debug, "", "sqs:ListDeadLetterSourceQueues", "Access Granted: No dead letter source queues", nil)
											}
											for _, deadLetterQueue := range deadLetterAttrOutput.QueueUrls {
												utils.PrintResult(debug, "", "sqs:ListDeadLetterSourceQueues", fmt.Sprintf("Found dead letter source queue: %s", *deadLetterQueue), nil)
											}
										}

										// Receive messages from the dead letter queue
										deadLetterReceiveOutput, deadLetterReceiveErr := deadLetterSvc.ReceiveMessage(&sqs.ReceiveMessageInput{
											QueueUrl: aws.String(deadLetterQueueUrl),
										})

										if deadLetterReceiveErr != nil {
											utils.HandleAWSError(debug, "sqs:ReceiveMessage", deadLetterReceiveErr)
										} else {
											if len(deadLetterReceiveOutput.Messages) == 0 {
												utils.PrintResult(debug, "", "sqs:ReceiveMessage", "Access Granted: No messages in dead letter queue", nil)
											}
											for _, message := range deadLetterReceiveOutput.Messages {
												utils.PrintResult(debug, "", "sqs:ReceiveMessage", fmt.Sprintf("Received message: %s", *message.Body), nil)
											}
										}
									}
								}
							}
						}
					}
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
