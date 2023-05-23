package sts

import (
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
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
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
