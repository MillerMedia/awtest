package lambda

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type LambdaDetails struct {
	Function *lambda.FunctionConfiguration
	Region   string
	Code     *lambda.GetFunctionOutput
}

var LambdaCalls = []types.AWSService{
	{
		Name: "lambda:ListFunctions",
		Call: func(sess *session.Session) (interface{}, error) {
			var allDetails []LambdaDetails

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
				svc := lambda.New(regionSess)
				output, err := svc.ListFunctions(&lambda.ListFunctionsInput{})
				if err != nil {
					return nil, err
				}

				for _, function := range output.Functions {
					getFuncOutput, err := svc.GetFunction(&lambda.GetFunctionInput{
						FunctionName: function.FunctionName,
					})
					if err != nil {
						return nil, err
					}

					details := LambdaDetails{
						Function: function,
						Region:   region,
						Code:     getFuncOutput,
					}
					allDetails = append(allDetails, details)
				}
			}
			return allDetails, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "lambda:ListFunctions", err)
			}
			if details, ok := output.([]LambdaDetails); ok {
				for _, detail := range details {
					functionName := *detail.Function.FunctionName
					codeLocation := *detail.Code.Code.Location
					utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("Lambda function: %s", utils.ColorizeItem(functionName)), nil)
					utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("Code Location: %s", codeLocation), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
