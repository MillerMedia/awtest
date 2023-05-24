package iam

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

var IAMCalls = []types.AWSService{
	{
		Name: "iam:ListUsers",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := iam.New(sess)
			output, err := svc.ListUsers(&iam.ListUsersInput{})
			return map[string]interface{}{
				"output": output,
				"sess":   sess,
			}, err
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "iam:ListUsers", err)
			}
			if outputMap, ok := output.(map[string]interface{}); ok {
				iamOutput, _ := outputMap["output"].(*iam.ListUsersOutput)
				sess, _ := outputMap["sess"].(*session.Session)
				svc := iam.New(sess)
				for _, user := range iamOutput.Users {
					utils.PrintResult(debug, "", "iam:ListUsers", fmt.Sprintf("IAM user: %s", utils.ColorizeItem(*user.UserName)), nil)

					// list groups for user
					groupOutput, err := svc.ListGroupsForUser(&iam.ListGroupsForUserInput{
						UserName: user.UserName,
					})
					if err != nil {
						return utils.HandleAWSError(debug, "iam:ListGroupsForUser", err)
					}
					for _, group := range groupOutput.Groups {
						utils.PrintResult(debug, "", "iam:ListGroupsForUser", fmt.Sprintf("IAM User: %s | group: %s", utils.ColorizeItem(*user.UserName), *group.GroupName), nil)
					}

					// list attached user policies
					attachedPolicyOutput, err := svc.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
						UserName: user.UserName,
					})
					if err != nil {
						return utils.HandleAWSError(debug, "iam:ListAttachedUserPolicies", err)
					}
					for _, policy := range attachedPolicyOutput.AttachedPolicies {
						utils.PrintResult(debug, "", "iam:ListAttachedUserPolicies", fmt.Sprintf("IAM user: %s | attached policy: %s", utils.ColorizeItem(*user.UserName), *policy.PolicyName), nil)
					}

					// list user policies
					policyOutput, err := svc.ListUserPolicies(&iam.ListUserPoliciesInput{
						UserName: user.UserName,
					})
					if err != nil {
						return utils.HandleAWSError(debug, "iam:ListUserPolicies", err)
					}
					for _, policyName := range policyOutput.PolicyNames {
						utils.PrintResult(debug, "", "iam:ListUserPolicies", fmt.Sprintf("IAM user: %s | inline policy: %s", utils.ColorizeItem(*user.UserName), *policyName), nil)
					}

					// list access keys
					accessKeyOutput, err := svc.ListAccessKeys(&iam.ListAccessKeysInput{
						UserName: user.UserName,
					})
					if err != nil {
						return utils.HandleAWSError(debug, "iam:ListAccessKeys", err)
					}
					for _, accessKey := range accessKeyOutput.AccessKeyMetadata {
						utils.PrintResult(debug, "", "iam:ListAccessKeys", fmt.Sprintf("IAM user: %s | access key: %s", utils.ColorizeItem(*user.UserName), *accessKey.AccessKeyId), nil)
					}
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
