package services

import (
	"github.com/MillerMedia/awtest/cmd/awtest/services/sns"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
)

func AllServices() []types.AWSService {
	var allServices []types.AWSService

	//allServices = append(allServices, sts.STSCalls...)
	//allServices = append(allServices, s3.S3Calls...)
	//allServices = append(allServices, ec2.EC2Calls...)
	//allServices = append(allServices, iam.IAMCalls...)
	//allServices = append(allServices, rds.RDSCalls...)
	allServices = append(allServices, sns.SNSCalls...)
	//allServices = append(allServices, sqs.SQSCalls...)
	//allServices = append(allServices, lambda.LambdaCalls...)
	//allServices = append(allServices, cloudwatch.CloudwatchCalls...)
	//allServices = append(allServices, secretsmanager.SecretsManagerCalls...)
	//allServices = append(allServices, appsync.AppSyncCalls...)
	//allServices = append(allServices, apigateway.APIGatewayCalls...)

	return allServices
}
