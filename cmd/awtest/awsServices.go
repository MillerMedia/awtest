package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/sts"
)

const DefaultModuleName = "AWTest"
const InvalidAccessKeyId = "InvalidAccessKeyId"

var Regions = []string{
	"us-east-1",
	"us-east-2",
	"us-west-1",
	"us-west-2",
	// Add more regions as needed...
}

type AWSService struct {
	Name       string
	Call       func(*session.Session) (interface{}, error)
	Process    func(interface{}, error, bool) error
	ModuleName string
}

var AWSListCalls = []AWSService{
	{
		Name: "sts:GetCallerIdentity",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := sts.New(sess)
			output, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
			return output, err
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return handleAWSError(debug, "sts:GetCallerIdentity", err)
			}
			if stsOutput, ok := output.(*sts.GetCallerIdentityOutput); ok {
				printResult(debug, "", "user-id", *stsOutput.UserId, nil)
				printResult(debug, "", "account-number", *stsOutput.Account, nil)
				printResult(debug, "", "iam-arn", *stsOutput.Arn, nil)
			}
			return nil
		},
		ModuleName: DefaultModuleName,
	},
	{
		Name: "s3:ListBuckets",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := s3.New(sess)
			output, err := svc.ListBuckets(&s3.ListBucketsInput{})
			return output, err
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return handleAWSError(debug, "s3:ListBuckets", err)
			}
			if s3Output, ok := output.(*s3.ListBucketsOutput); ok {
				for _, bucket := range s3Output.Buckets {
					printResult(debug, "", "s3:ListBuckets", fmt.Sprintf("Found S3 bucket: %s", *bucket.Name), nil)
				}
			}
			return nil
		},
		ModuleName: DefaultModuleName,
	},
	{
		Name: "ec2:DescribeInstances",
		Call: func(sess *session.Session) (interface{}, error) {
			var allInstances []*ec2.Instance
			for _, region := range Regions {
				sess.Config.Region = aws.String(region)
				svc := ec2.New(sess)
				input := &ec2.DescribeInstancesInput{}
				output, err := svc.DescribeInstances(input)
				if err != nil {
					return nil, err
				}
				for _, reservation := range output.Reservations {
					allInstances = append(allInstances, reservation.Instances...)
				}
			}
			return allInstances, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return handleAWSError(debug, "ec2:DescribeInstances", err)
			}
			if instances, ok := output.([]*ec2.Instance); ok {
				for _, instance := range instances {
					printResult(debug, "", "ec2:DescribeInstances", fmt.Sprintf("Found EC2 instance: %s", *instance.InstanceId), nil)
				}
			}
			return nil
		},
		ModuleName: DefaultModuleName,
	},
	{
		Name: "rds:DescribeDBInstances",
		Call: func(sess *session.Session) (interface{}, error) {
			var allDBInstances []*rds.DBInstance
			for _, region := range Regions {
				sess.Config.Region = aws.String(region)
				svc := rds.New(sess)
				output, err := svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{})
				if err != nil {
					return nil, err
				}
				allDBInstances = append(allDBInstances, output.DBInstances...)
			}
			return allDBInstances, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return handleAWSError(debug, "rds:DescribeDBInstances", err)
			}
			if dbInstances, ok := output.([]*rds.DBInstance); ok {
				for _, db := range dbInstances {
					printResult(debug, "", "rds:DescribeDBInstances", fmt.Sprintf("Found RDS instance: %s (%s)", *db.DBInstanceIdentifier, *db.AvailabilityZone), nil)
				}
			}
			return nil
		},
		ModuleName: DefaultModuleName,
	},
	{
		Name: "secretsmanager:ListSecrets",
		Call: func(sess *session.Session) (interface{}, error) {
			var allSecrets []*secretsmanager.SecretListEntry
			for _, region := range Regions {
				sess.Config.Region = aws.String(region)
				svc := secretsmanager.New(sess)
				output, err := svc.ListSecrets(&secretsmanager.ListSecretsInput{})
				if err != nil {
					return nil, err
				}
				allSecrets = append(allSecrets, output.SecretList...)
			}
			return allSecrets, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return handleAWSError(debug, "secretsmanager:ListSecrets", err)
			}
			if secrets, ok := output.([]*secretsmanager.SecretListEntry); ok {
				for _, secret := range secrets {
					printResult(debug, "", "secretsmanager:ListSecrets", fmt.Sprintf("Found secret: %s", *secret.Name), nil)
				}
			}
			return nil
		},
		ModuleName: DefaultModuleName,
	},
	{
		Name: "iam:ListUsers",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := iam.New(sess)
			output, err := svc.ListUsers(&iam.ListUsersInput{})
			return output, err
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return handleAWSError(debug, "iam:ListUsers", err)
			}
			if iamOutput, ok := output.(*iam.ListUsersOutput); ok {
				for _, user := range iamOutput.Users {
					printResult(debug, "", "iam:ListUsers", fmt.Sprintf("Found IAM user: %s", *user.UserName), nil)
				}
			}
			return nil
		},
		ModuleName: DefaultModuleName,
	},
	{
		Name: "appsync:ListGraphqlApis",
		Call: func(sess *session.Session) (interface{}, error) {
			var allApis []*appsync.GraphqlApi
			for _, region := range Regions {
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
				return handleAWSError(debug, "appsync:ListGraphqlApis", err)
			}
			if apis, ok := output.([]*appsync.GraphqlApi); ok {
				for _, api := range apis {
					printResult(debug, "", "appsync:ListGraphqlApis", fmt.Sprintf("Found AppSync API: %s", *api.Name), nil)
				}
			}
			return nil
		},
		ModuleName: DefaultModuleName,
	},
	{
		Name: "apigateway:RestApis",
		Call: func(sess *session.Session) (interface{}, error) {
			var allApis []*apigateway.RestApi
			for _, region := range Regions {
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
				return handleAWSError(debug, "apigateway:RestApis", err)
			}
			if apis, ok := output.([]*apigateway.RestApi); ok {
				for _, api := range apis {
					printResult(debug, "", "apigateway:RestApis", fmt.Sprintf("Found API Gateway: %s", *api.Name), nil)
				}
			}
			return nil
		},
		ModuleName: DefaultModuleName,
	},
	//{
	//	Name: "appconfig:ListApplications",
	//	Call: func(sess *session.Session) (interface{}, error) {
	//		var allApps []*appconfig.Application
	//		for _, region := range Regions {
	//			sess.Config.Region = aws.String(region)
	//			svc := appconfig.New(sess)
	//			input := &appconfig.ListApplicationsInput{}
	//			err := svc.ListApplicationsPages(input,
	//				func(page *appconfig.ListApplicationsOutput, lastPage bool) bool {
	//					allApps = append(allApps, page.Applications...)
	//					return !lastPage
	//				})
	//			if err != nil {
	//				return nil, err
	//			}
	//		}
	//		return allApps, nil
	//	},
	//	Process: func(output interface{}, err error, debug bool) error {
	//		if err != nil {
	//			return handleAWSError(debug, "appconfig:ListApplications", err)
	//		}
	//		if apps, ok := output.([]*appconfig.Application); ok {
	//			for _, app := range apps {
	//				printResult(debug, "", "appconfig:ListApplications", fmt.Sprintf("Found AppConfig Application: %s", *app.Name), nil)
	//			}
	//		}
	//		return nil
	//	},
	//	ModuleName: DefaultModuleName,
	//},
	//{
	//	Name: "dynamodb:ListTables",
	//	Call: func(sess *session.Session) (interface{}, error) {
	//		var allTables []*string
	//		for _, region := range Regions {
	//			regionSess, err := session.NewSession(&aws.Config{
	//				Region: aws.String(region),
	//			})
	//			if err != nil {
	//				return nil, err
	//			}
	//			svc := dynamodb.New(regionSess)
	//			output, err := svc.ListTables(&dynamodb.ListTablesInput{})
	//			if err != nil {
	//				return nil, err
	//			}
	//			allTables = append(allTables, output.TableNames...)
	//		}
	//		return allTables, nil
	//	},
	//	Process: func(output interface{}, err error, debug bool) error {
	//		if err != nil {
	//			return handleAWSError(debug, "dynamodb:ListTables", err)
	//		}
	//		if tables, ok := output.([]*string); ok {
	//			for _, table := range tables {
	//				printResult(debug, "", "dynamodb:ListTables", fmt.Sprintf("Found DynamoDB table: %s", *table), nil)
	//			}
	//		}
	//		return nil
	//	},
	//	ModuleName: DefaultModuleName,
	//},
}
