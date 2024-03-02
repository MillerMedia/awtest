package sts

import (
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"strings"
)

var STSCalls = []types.AWSService{
	{
		Name: "sts:GetCallerIdentity",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := sts.New(sess)
			output, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
			return output, err
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "sts:GetCallerIdentity", err)
			}
			if stsOutput, ok := output.(*sts.GetCallerIdentityOutput); ok {
				utils.PrintResult(debug, "", "user-id", *stsOutput.UserId, nil)
				utils.PrintResult(debug, "", "account-number", *stsOutput.Account, nil)
				utils.PrintResult(debug, "", "iam-arn", *stsOutput.Arn, nil)

				// Parse the ARN inline to get user name
				arnParts := strings.Split(*stsOutput.Arn, "/")
				userName := arnParts[len(arnParts)-1]
				utils.PrintResult(debug, "", "iam-user", userName, nil)

				// List attached user policies by calling the IAM service using the Policy Simulator
				sess := session.Must(session.NewSession())
				svc := iam.New(sess)
				attachedPolicyOutput, err := svc.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
					UserName: &userName,
				})

				if err != nil {
					utils.HandleAWSError(debug, "iam:ListAttachedUserPolicies", err)
				} else {
					for _, policy := range attachedPolicyOutput.AttachedPolicies {
						utils.PrintResult(debug, "", "iam:ListAttachedUserPolicies", *policy.PolicyArn, nil)
					}
				}

				// List user policies by calling the IAM service
				policyOutput, err := svc.ListUserPolicies(&iam.ListUserPoliciesInput{
					UserName: &userName,
				})

				if err != nil {
					utils.HandleAWSError(debug, "iam:ListUserPolicies", err)
				} else {
					for _, policy := range policyOutput.PolicyNames {
						utils.PrintResult(debug, "", "iam:ListUserPolicies", *policy, nil)
					}
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
