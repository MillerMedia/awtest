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
			return output, err
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "iam:ListUsers", err)
			}
			if iamOutput, ok := output.(*iam.ListUsersOutput); ok {
				for _, user := range iamOutput.Users {
					utils.PrintResult(debug, "", "iam:ListUsers", fmt.Sprintf("Found IAM user: %s", *user.UserName), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
