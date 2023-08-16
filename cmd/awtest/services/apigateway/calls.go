package apigateway

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
)

var APIGatewayCalls = []types.AWSService{
	{
		Name: "apigateway:RestApis",
		Call: func(sess *session.Session) (interface{}, error) {
			var allApis []*apigateway.RestApi
			for _, region := range types.Regions {
				regionSess, err := session.NewSession(&aws.Config{
					Region: aws.String(region),
				})
				if err != nil {
					return nil, err
				}
				svc := apigateway.New(regionSess)
				output, err := svc.GetRestApis(&apigateway.GetRestApisInput{})
				if err != nil {
					return nil, err
				}
				allApis = append(allApis, output.Items...)
			}
			return allApis, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "apigateway:RestApis", err)
			}
			if apis, ok := output.([]*apigateway.RestApi); ok {
				for _, api := range apis {
					utils.PrintResult(debug, "", "apigateway:RestApis", fmt.Sprintf("Found API Gateway: %s", *api.Name), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "apigateway:GetApiKeys",
		Call: func(sess *session.Session) (interface{}, error) {
			var allApiKeys []*apigateway.ApiKey
			for _, region := range types.Regions {
				regionSess, err := session.NewSession(&aws.Config{
					Region: aws.String(region),
				})
				if err != nil {
					return nil, err
				}
				svc := apigateway.New(regionSess)
				output, err := svc.GetApiKeys(&apigateway.GetApiKeysInput{})
				if err != nil {
					return nil, err
				}
				allApiKeys = append(allApiKeys, output.Items...)
			}
			return allApiKeys, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "apigateway:GetApiKeys", err)
			}
			if apiKeys, ok := output.([]*apigateway.ApiKey); ok {
				if len(apiKeys) == 0 {
					utils.PrintResult(debug, "", "apigateway:GetApiKeys", "No API keys found, but access is granted.", nil)
				} else {
					for _, apiKey := range apiKeys {
						utils.PrintResult(debug, "", "apigateway:GetApiKeys", fmt.Sprintf("Found API Key: %s", *apiKey.Id), nil)
					}
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
