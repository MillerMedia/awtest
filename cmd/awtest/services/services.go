package services

import (
	"github.com/MillerMedia/awtest/cmd/awtest/services/apigateway"
	"github.com/MillerMedia/awtest/cmd/awtest/services/appsync"
	"github.com/MillerMedia/awtest/cmd/awtest/services/cloudwatch"
	"github.com/MillerMedia/awtest/cmd/awtest/services/dynamodb"
	"github.com/MillerMedia/awtest/cmd/awtest/services/ec2"
	"github.com/MillerMedia/awtest/cmd/awtest/services/glacier"
	"github.com/MillerMedia/awtest/cmd/awtest/services/iam"
	"github.com/MillerMedia/awtest/cmd/awtest/services/kms"
	"github.com/MillerMedia/awtest/cmd/awtest/services/lambda"
	"github.com/MillerMedia/awtest/cmd/awtest/services/rds"
	"github.com/MillerMedia/awtest/cmd/awtest/services/s3"
	"github.com/MillerMedia/awtest/cmd/awtest/services/secretsmanager"
	"github.com/MillerMedia/awtest/cmd/awtest/services/sns"
	"github.com/MillerMedia/awtest/cmd/awtest/services/sqs"
	"github.com/MillerMedia/awtest/cmd/awtest/services/sts"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
)

func AllServices() []types.AWSService {
	var allServices []types.AWSService

	allServices = append(allServices, sts.STSCalls...)
	allServices = append(allServices, s3.S3Calls...)
	allServices = append(allServices, ec2.EC2Calls...)
	allServices = append(allServices, iam.IAMCalls...)
	allServices = append(allServices, rds.RDSCalls...)
	allServices = append(allServices, sns.SNSCalls...)
	allServices = append(allServices, sqs.SQSCalls...)
	allServices = append(allServices, lambda.LambdaCalls...)
	allServices = append(allServices, cloudwatch.CloudwatchCalls...)
	allServices = append(allServices, secretsmanager.SecretsManagerCalls...)
	allServices = append(allServices, appsync.AppSyncCalls...)
	allServices = append(allServices, apigateway.APIGatewayCalls...)
	allServices = append(allServices, dynamodb.DynamoDBCalls...)
	allServices = append(allServices, glacier.GlacierCalls...)
	allServices = append(allServices, kms.KMSCalls...)

	return allServices
}
