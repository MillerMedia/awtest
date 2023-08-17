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
	{
		Name: "dynamodb:ListExports",
		Call: func(sess *session.Session) (interface{}, error) {
			var allExports []*dynamodb.ExportSummary
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := dynamodb.New(sess)
				output, err := svc.ListExports(&dynamodb.ListExportsInput{})
				if err != nil {
					return nil, err
				}
				allExports = append(allExports, output.ExportSummaries...)
			}
			return allExports, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "dynamodb:ListExports", err)
			}
			if exports, ok := output.([]*dynamodb.ExportSummary); ok {
				if len(exports) == 0 {
					utils.PrintAccessGranted(debug, "dynamodb:ListExports", "DynamoDB exports")
				} else {
					for _, export := range exports {
						utils.PrintResult(debug, "", "dynamodb:ListExports", fmt.Sprintf("DynamoDB export: %s", utils.ColorizeItem(*export.ExportArn)), nil)
					}
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "dynamodb:ListBackups",
		Call: func(sess *session.Session) (interface{}, error) {
			var allBackups []*dynamodb.BackupSummary
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := dynamodb.New(sess)
				output, err := svc.ListBackups(&dynamodb.ListBackupsInput{})
				if err != nil {
					return nil, err
				}
				allBackups = append(allBackups, output.BackupSummaries...)
			}
			return allBackups, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "dynamodb:ListBackups", err)
			}
			if backups, ok := output.([]*dynamodb.BackupSummary); ok {
				if len(backups) == 0 {
					utils.PrintAccessGranted(debug, "dynamodb:ListBackups", "DynamoDB backups")
				} else {
					for _, backup := range backups {
						utils.PrintResult(debug, "", "dynamodb:ListBackups", fmt.Sprintf("DynamoDB backup: %s", utils.ColorizeItem(*backup.BackupArn)), nil)
					}
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
