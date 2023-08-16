package apigateway

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
)

type ApiWithStages struct {
	Api       *apigateway.RestApi
	Stages    []*apigateway.Stage
	Models    []*apigateway.Model
	Resources []*apigateway.Resource // Add this line
	Region    string
}

var APIGatewayCalls = []types.AWSService{
	{
		Name: "apigateway:RestApis",
		Call: func(sess *session.Session) (interface{}, error) {
			var allApisWithStages []ApiWithStages

			for _, region := range types.Regions {
				regionSess, err := session.NewSession(&aws.Config{
					Region: aws.String(region),
				})
				if err != nil {
					return nil, err
				}
				svc := apigateway.New(regionSess)
				apisOutput, err := svc.GetRestApis(&apigateway.GetRestApisInput{})
				if err != nil {
					return nil, err
				}
				for _, api := range apisOutput.Items {
					stagesOutput, err := svc.GetStages(&apigateway.GetStagesInput{RestApiId: api.Id})
					if err != nil {
						return nil, err
					}
					modelsOutput, err := svc.GetModels(&apigateway.GetModelsInput{RestApiId: api.Id})
					if err != nil {
						return nil, err
					}
					resourcesOutput, err := svc.GetResources(&apigateway.GetResourcesInput{RestApiId: api.Id}) // Add this line
					if err != nil {
						return nil, err
					}
					apiWithStages := ApiWithStages{
						Api:       api,
						Stages:    stagesOutput.Item,
						Models:    modelsOutput.Items,
						Resources: resourcesOutput.Items, // Add this line
						Region:    region,
					}
					allApisWithStages = append(allApisWithStages, apiWithStages)
				}
			}
			return allApisWithStages, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "apigateway:RestApis", err)
			}
			if apisWithStages, ok := output.([]ApiWithStages); ok {
				for _, apiWithStages := range apisWithStages {
					apiName := *apiWithStages.Api.Name
					restApiId := *apiWithStages.Api.Id
					region := apiWithStages.Region
					apiUrl := fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com", restApiId, region)
					utils.PrintResult(debug, "", "apigateway:RestApis", fmt.Sprintf("Found API Gateway: %s, Base URL: %s", apiName, apiUrl), nil)

					// Add this loop to print resources
					if len(apiWithStages.Resources) > 0 {
						for _, resource := range apiWithStages.Resources {
							resourceID := *resource.Id
							resourcePath := *resource.Path

							utils.PrintResult(debug, "", "apigateway:GetResources", fmt.Sprintf("Resource ID: %s", resourceID), nil)
							utils.PrintResult(debug, "", "apigateway:GetResources", fmt.Sprintf("Resource Path: %s", resourcePath), nil)

							// Check if the ResourceMethods map is not nil
							if resource.ResourceMethods != nil {
								for method, _ := range resource.ResourceMethods {
									utils.PrintResult(debug, "", "apigateway:GetResources", fmt.Sprintf("Resource Method: %s", method), nil)
								}
							}
						}
					}

					if len(apiWithStages.Stages) == 0 {
						utils.PrintResult(debug, "", "apigateway:GetStages", fmt.Sprintf("No stages found for API: %s, but access is granted.", apiName), nil)
					} else {
						for _, stage := range apiWithStages.Stages {
							utils.PrintResult(debug, "", "apigateway:GetStages", fmt.Sprintf("Found Stage: %s (%s)", *stage.StageName, apiName), nil)
						}
					}
					if len(apiWithStages.Models) > 0 {
						for _, model := range apiWithStages.Models {
							utils.PrintResult(debug, "", "apigateway:GetModels", fmt.Sprintf("Found Model: %s (%s)", *model.Name, apiName), nil)
						}
					}
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
	{
		Name: "apigateway:GetDomainNames",
		Call: func(sess *session.Session) (interface{}, error) {
			var allDomainNames []*apigateway.DomainName
			for _, region := range types.Regions {
				regionSess, err := session.NewSession(&aws.Config{
					Region: aws.String(region),
				})
				if err != nil {
					return nil, err
				}
				svc := apigateway.New(regionSess)
				output, err := svc.GetDomainNames(&apigateway.GetDomainNamesInput{})
				if err != nil {
					return nil, err
				}
				allDomainNames = append(allDomainNames, output.Items...)
			}
			return allDomainNames, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "apigateway:GetDomainNames", err)
			}
			if domainNames, ok := output.([]*apigateway.DomainName); ok {
				if len(domainNames) == 0 {
					utils.PrintResult(debug, "", "apigateway:GetDomainNames", "No domain names found, but access is granted.", nil)
				} else {
					for _, domainName := range domainNames {
						utils.PrintResult(debug, "", "apigateway:GetDomainNames", fmt.Sprintf("Found Domain Name: %s", *domainName.DomainName), nil)
					}
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
