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
	Function      *lambda.FunctionConfiguration
	Region        string
	Code          *lambda.GetFunctionOutput
	Configuration *lambda.FunctionConfiguration
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

					configOutput, err := svc.GetFunctionConfiguration(&lambda.GetFunctionConfigurationInput{
						FunctionName: function.FunctionName,
					})
					if err != nil {
						return nil, err
					}

					details := LambdaDetails{
						Function:      function,
						Region:        region,
						Code:          getFuncOutput,
						Configuration: configOutput,
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
				if len(details) == 0 {
					utils.PrintResult(debug, "", "lambda:ListFunctions", "No Lambda functions found.", nil)
				} else {
					for _, detail := range details {
						functionName := *detail.Function.FunctionName
						codeLocation := *detail.Code.Code.Location
						memorySize := fmt.Sprintf("%d MB", *detail.Configuration.MemorySize)
						timeout := fmt.Sprintf("%d seconds", *detail.Configuration.Timeout)
						runtime := *detail.Configuration.Runtime
						handler := *detail.Configuration.Handler
						role := *detail.Configuration.Role
						lastModified := *detail.Configuration.LastModified

						utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("Lambda function: %s", utils.ColorizeItem(functionName)), nil)
						utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("Runtime: %s", runtime), nil)
						utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("Memory Size: %s", memorySize), nil)
						utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("Timeout: %s", timeout), nil)
						utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("Handler: %s", handler), nil)
						utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("Role: %s", role), nil)
						utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("Last Modified: %s", lastModified), nil)

						// Print Description if it is not empty
						if detail.Configuration.Description != nil && *detail.Configuration.Description != "" {
							utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("Description: %s", *detail.Configuration.Description), nil)
						}

						// Print environment variables if they exist
						if detail.Configuration.Environment != nil && len(detail.Configuration.Environment.Variables) > 0 {
							utils.PrintResult(debug, "", "lambda:ListFunctions", "Environment Variables:", nil)
							for key, value := range detail.Configuration.Environment.Variables {
								utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("  %s: %s", key, aws.StringValue(value)), nil)
							}
						}

						// Print KMS Key ARN if it exists
						if detail.Configuration.KMSKeyArn != nil && *detail.Configuration.KMSKeyArn != "" {
							utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("KMS Key ARN: %s", *detail.Configuration.KMSKeyArn), nil)
						}

						utils.PrintResult(debug, "", "lambda:ListFunctions", fmt.Sprintf("Code Location: %s", codeLocation), nil)
					}
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
