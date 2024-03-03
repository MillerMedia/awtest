package services

import (
	"github.com/MillerMedia/awtest/cmd/awtest/services/amplify"
	"github.com/MillerMedia/awtest/cmd/awtest/services/apigateway"
	"github.com/MillerMedia/awtest/cmd/awtest/services/appsync"
	"github.com/MillerMedia/awtest/cmd/awtest/services/batch"
	"github.com/MillerMedia/awtest/cmd/awtest/services/cloudformation"
	"github.com/MillerMedia/awtest/cmd/awtest/services/cloudfront"
	"github.com/MillerMedia/awtest/cmd/awtest/services/cloudtrail"
	"github.com/MillerMedia/awtest/cmd/awtest/services/cloudwatch"
	"github.com/MillerMedia/awtest/cmd/awtest/services/codepipeline"
	"github.com/MillerMedia/awtest/cmd/awtest/services/cognitoidentity"
	"github.com/MillerMedia/awtest/cmd/awtest/services/dynamodb"
	"github.com/MillerMedia/awtest/cmd/awtest/services/ec2"
	"github.com/MillerMedia/awtest/cmd/awtest/services/elasticbeanstalk"
	"github.com/MillerMedia/awtest/cmd/awtest/services/eventbridge"
	"github.com/MillerMedia/awtest/cmd/awtest/services/glacier"
	"github.com/MillerMedia/awtest/cmd/awtest/services/glue"
	"github.com/MillerMedia/awtest/cmd/awtest/services/iam"
	"github.com/MillerMedia/awtest/cmd/awtest/services/iot"
	"github.com/MillerMedia/awtest/cmd/awtest/services/ivs"
	"github.com/MillerMedia/awtest/cmd/awtest/services/ivschat"
	"github.com/MillerMedia/awtest/cmd/awtest/services/ivsrealtime"
	"github.com/MillerMedia/awtest/cmd/awtest/services/kms"
	"github.com/MillerMedia/awtest/cmd/awtest/services/lambda"
	"github.com/MillerMedia/awtest/cmd/awtest/services/rds"
	"github.com/MillerMedia/awtest/cmd/awtest/services/route53"
	"github.com/MillerMedia/awtest/cmd/awtest/services/s3"
	"github.com/MillerMedia/awtest/cmd/awtest/services/secretsmanager"
	"github.com/MillerMedia/awtest/cmd/awtest/services/ses"
	"github.com/MillerMedia/awtest/cmd/awtest/services/sns"
	"github.com/MillerMedia/awtest/cmd/awtest/services/sqs"
	"github.com/MillerMedia/awtest/cmd/awtest/services/sts"
	"github.com/MillerMedia/awtest/cmd/awtest/services/transcribe"
	"github.com/MillerMedia/awtest/cmd/awtest/services/waf"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
)

func AllServices() []types.AWSService {
	var allServices []types.AWSService

	allServices = append(allServices, sts.STSCalls...)
	allServices = append(allServices, amplify.AmplifyCalls...)
	allServices = append(allServices, apigateway.APIGatewayCalls...)
	allServices = append(allServices, appsync.AppSyncCalls...)
	allServices = append(allServices, batch.BatchCalls...)
	allServices = append(allServices, cloudformation.CloudFormationCalls...)
	allServices = append(allServices, cloudfront.CloudFrontCalls...)
	allServices = append(allServices, cloudtrail.CloudTrailCalls...)
	allServices = append(allServices, cloudwatch.CloudwatchCalls...)
	allServices = append(allServices, codepipeline.CodePipelineCalls...)
	allServices = append(allServices, cognitoidentity.CognitoIdentityCalls...)
	allServices = append(allServices, dynamodb.DynamoDBCalls...)
	allServices = append(allServices, ec2.EC2Calls...)
	allServices = append(allServices, elasticbeanstalk.ElasticBeanstalkCalls...)
	allServices = append(allServices, eventbridge.EventbridgeCalls...)
	allServices = append(allServices, glacier.GlacierCalls...)
	allServices = append(allServices, glue.GlueCalls...)
	allServices = append(allServices, iam.IAMCalls...)
	allServices = append(allServices, iot.IoTCalls...)
	allServices = append(allServices, ivs.IvsCalls...)
	allServices = append(allServices, ivschat.IvsChatCalls...)
	allServices = append(allServices, ivsrealtime.IvsRealtimeCalls...)
	allServices = append(allServices, kms.KMSCalls...)
	allServices = append(allServices, lambda.LambdaCalls...)
	allServices = append(allServices, rds.RDSCalls...)
	allServices = append(allServices, route53.Route53Calls...)
	allServices = append(allServices, s3.S3Calls...)
	allServices = append(allServices, secretsmanager.SecretsManagerCalls...)
	allServices = append(allServices, ses.SESCalls...)
	allServices = append(allServices, sns.SNSCalls...)
	allServices = append(allServices, sqs.SQSCalls...)
	allServices = append(allServices, transcribe.TranscribeCalls...)
	allServices = append(allServices, waf.WafCalls...)

	return allServices
}
