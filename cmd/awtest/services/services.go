package services

import (
	"github.com/MillerMedia/awtest/cmd/awtest/services/amplify"
	"github.com/MillerMedia/awtest/cmd/awtest/services/apigateway"
	"github.com/MillerMedia/awtest/cmd/awtest/services/appsync"
	"github.com/MillerMedia/awtest/cmd/awtest/services/cloudfront"
	"github.com/MillerMedia/awtest/cmd/awtest/services/cloudwatch"
	"github.com/MillerMedia/awtest/cmd/awtest/services/cognitoidentity"
	"github.com/MillerMedia/awtest/cmd/awtest/services/dynamodb"
	"github.com/MillerMedia/awtest/cmd/awtest/services/ec2"
	"github.com/MillerMedia/awtest/cmd/awtest/services/elasticbeanstalk"
	"github.com/MillerMedia/awtest/cmd/awtest/services/glacier"
	"github.com/MillerMedia/awtest/cmd/awtest/services/iam"
	"github.com/MillerMedia/awtest/cmd/awtest/services/iot"
	"github.com/MillerMedia/awtest/cmd/awtest/services/kms"
	"github.com/MillerMedia/awtest/cmd/awtest/services/lambda"
	"github.com/MillerMedia/awtest/cmd/awtest/services/rds"
	"github.com/MillerMedia/awtest/cmd/awtest/services/s3"
	"github.com/MillerMedia/awtest/cmd/awtest/services/secretsmanager"
	"github.com/MillerMedia/awtest/cmd/awtest/services/ses"
	"github.com/MillerMedia/awtest/cmd/awtest/services/sns"
	"github.com/MillerMedia/awtest/cmd/awtest/services/sqs"
	"github.com/MillerMedia/awtest/cmd/awtest/services/sts"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
)

func AllServices() []types.AWSService {
	var allServices []types.AWSService

	allServices = append(allServices, sts.STSCalls...)
	allServices = append(allServices, amplify.AmplifyCalls...)
	allServices = append(allServices, apigateway.APIGatewayCalls...)
	allServices = append(allServices, appsync.AppSyncCalls...)
	allServices = append(allServices, cloudfront.CloudFrontCalls...)
	allServices = append(allServices, cloudwatch.CloudwatchCalls...)
	allServices = append(allServices, cognitoidentity.CognitoIdentityCalls...)
	allServices = append(allServices, dynamodb.DynamoDBCalls...)
	allServices = append(allServices, ec2.EC2Calls...)
	allServices = append(allServices, elasticbeanstalk.ElasticBeanstalkCalls...)
	allServices = append(allServices, glacier.GlacierCalls...)
	allServices = append(allServices, iam.IAMCalls...)
	allServices = append(allServices, iot.IoTCalls...)
	allServices = append(allServices, kms.KMSCalls...)
	allServices = append(allServices, lambda.LambdaCalls...)
	allServices = append(allServices, rds.RDSCalls...)
	allServices = append(allServices, s3.S3Calls...)
	allServices = append(allServices, secretsmanager.SecretsManagerCalls...)
	allServices = append(allServices, ses.SESCalls...)
	allServices = append(allServices, sns.SNSCalls...)
	allServices = append(allServices, sqs.SQSCalls...)

	return allServices
}
