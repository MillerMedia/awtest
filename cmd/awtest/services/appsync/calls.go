package appsync

import (
	"fmt"
	"github.com/MillerMedia/AWTest/cmd/awtest/types"
	"github.com/MillerMedia/AWTest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appsync"
)

var AppSyncCalls = []types.AWSService{
	{
		Name: "appsync:ListGraphqlApis",
		Call: func(sess *session.Session) (interface{}, error) {
			var allApis []*appsync.GraphqlApi
			for _, region := range types.Regions {
				regionSess, err := session.NewSession(&aws.Config{
					Region: aws.String(region),
				})
				if err != nil {
					return nil, err
				}
				svc := appsync.New(regionSess)
				output, err := svc.ListGraphqlApis(&appsync.ListGraphqlApisInput{})
				if err != nil {
					return nil, err
				}
				allApis = append(allApis, output.GraphqlApis...)
			}
			return allApis, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "appsync:ListGraphqlApis", err)
			}
			if apis, ok := output.([]*appsync.GraphqlApi); ok {
				for _, api := range apis {
					utils.PrintResult(debug, "", "appsync:ListGraphqlApis", fmt.Sprintf("Found AppSync API: %s", *api.Name), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
