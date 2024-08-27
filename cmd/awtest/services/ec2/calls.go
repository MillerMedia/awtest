package ec2

import (
	"encoding/base64"
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type EC2InstanceDetails struct {
	Instance  *ec2.Instance
	UserData  string
	ElasticIP string
	Tags      []*ec2.Tag
}

var EC2Calls = []types.AWSService{
	{
		Name: "ec2:DescribeInstances",
		Call: func(sess *session.Session) (interface{}, error) {
			var allDetails []EC2InstanceDetails
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := ec2.New(sess)
				input := &ec2.DescribeInstancesInput{}
				output, err := svc.DescribeInstances(input)
				if err != nil {
					return nil, err
				}
				for _, reservation := range output.Reservations {
					for _, instance := range reservation.Instances {
						// Get user data
						userData, err := getInstanceUserData(svc, *instance.InstanceId)
						if err != nil {
							userData = fmt.Sprintf("Error retrieving user data: %v", err)
						}

						// Get Elastic IP
						elasticIP := ""
						addressesOutput, err := svc.DescribeAddresses(&ec2.DescribeAddressesInput{
							Filters: []*ec2.Filter{
								{
									Name:   aws.String("instance-id"),
									Values: []*string{instance.InstanceId},
								},
							},
						})
						if err == nil && len(addressesOutput.Addresses) > 0 {
							elasticIP = *addressesOutput.Addresses[0].PublicIp
						}

						// Store instance details with user data, Elastic IP, and tags
						details := EC2InstanceDetails{
							Instance:  instance,
							UserData:  userData,
							ElasticIP: elasticIP,
							Tags:      instance.Tags,
						}
						allDetails = append(allDetails, details)
					}
				}
			}
			return allDetails, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "ec2:DescribeInstances", err)
			}
			if details, ok := output.([]EC2InstanceDetails); ok {
				for _, detail := range details {
					instance := detail.Instance
					ipAddress := ""
					if instance.PublicIpAddress != nil {
						ipAddress = *instance.PublicIpAddress
					}

					// Print instance details
					utils.PrintResult(debug, "", "ec2:DescribeInstances", fmt.Sprintf("EC2 instance: [ID: %s, State: %s, Type: %s, IP: %s]",
						utils.ColorizeItem(*instance.InstanceId),
						*instance.State.Name,
						*instance.InstanceType,
						ipAddress), nil)

					// Print Elastic IP if it exists
					if detail.ElasticIP != "" {
						utils.PrintResult(debug, "", "ec2:DescribeInstances", fmt.Sprintf("Elastic IP: %s", detail.ElasticIP), nil)
					}

					// Print user data
					if detail.UserData != "" {
						utils.PrintResult(debug, "", "ec2:DescribeInstances", fmt.Sprintf("User Data for instance %s:\n%s", *instance.InstanceId, detail.UserData), nil)
					}

					// Print tags if they exist
					if len(detail.Tags) > 0 {
						utils.PrintResult(debug, "", "ec2:DescribeInstances", "Tags:", nil)
						for _, tag := range detail.Tags {
							utils.PrintResult(debug, "", "ec2:DescribeInstances", fmt.Sprintf("  %s: %s", *tag.Key, *tag.Value), nil)
						}
					}
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}

// getInstanceUserData retrieves the user data for a specific EC2 instance.
func getInstanceUserData(svc *ec2.EC2, instanceId string) (string, error) {
	input := &ec2.DescribeInstanceAttributeInput{
		InstanceId: aws.String(instanceId),
		Attribute:  aws.String("userData"),
	}

	result, err := svc.DescribeInstanceAttribute(input)
	if err != nil {
		return "", err
	}

	// User data is returned as base64 encoded, so we need to decode it
	if result.UserData != nil && result.UserData.Value != nil {
		userDataBytes, err := base64.StdEncoding.DecodeString(*result.UserData.Value)
		if err != nil {
			return "", err
		}
		return string(userDataBytes), nil
	}

	return "", nil
}
