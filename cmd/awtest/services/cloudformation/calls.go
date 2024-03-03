package cloudformation

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var CloudFormationCalls = []types.AWSService{
	{
		Name: "cloudformation:ListStacks",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := cloudformation.New(sess)
			input := &cloudformation.ListStacksInput{}
			return svc.ListStacks(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "cloudformation:ListStacks", err)
			}
			if stacks, ok := output.(*cloudformation.ListStacksOutput); ok {
				for _, stack := range stacks.StackSummaries {
					utils.PrintResult(debug, "", "cloudformation:ListStacks", fmt.Sprintf("Stack: %s", *stack.StackName), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
