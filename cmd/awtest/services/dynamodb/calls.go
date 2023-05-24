package dynamodb

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var DynamoDBCalls = []types.AWSService{
	{
		Name: "dynamodb:ListTables",
		Call: func(sess *session.Session) (interface{}, error) {
			var allTableNames []*string
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := dynamodb.New(sess)
				output, err := svc.ListTables(&dynamodb.ListTablesInput{})
				if err != nil {
					return nil, err
				}
				allTableNames = append(allTableNames, output.TableNames...)
			}
			return allTableNames, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "dynamodb:ListTables", err)
			}
			if tableNames, ok := output.([]*string); ok {
				for _, tableName := range tableNames {
					utils.PrintResult(debug, "", "dynamodb:ListTables", fmt.Sprintf("DynamoDB table: %s", utils.ColorizeItem(*tableName)), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
