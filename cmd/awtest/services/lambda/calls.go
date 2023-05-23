package lambda

import (
	"fmt"
	"github.com/MillerMedia/AWTest/cmd/awtest/types"
	"github.com/MillerMedia/AWTest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var LambdaCalls = []types.AWSService{
	{
		Name: "lambda:ListFunctions",
		Call: func(sess *session.Session) (interface{}, error) {
			var allFunctions []*lambda.FunctionConfiguration
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := lambda.New(sess)
				output, err := svc.ListFunctions(&lambda.ListFunctionsInput{})
				if err != nil {
					return nil, err
				}
				allFunctions = append(allFunctions, output.Functions...)
			}
			return allFunctions, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "lambda:ListFunctions", err)
			}
			if functions, ok := output.([]*lambda.FunctionConfiguration); ok {
				for _, function := range functions {
					utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("Found Lambda function: %s", *function.FunctionName), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
