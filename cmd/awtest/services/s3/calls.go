package s3

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var S3Calls = []types.AWSService{
	{
		Name: "s3:ListBuckets",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := s3.New(sess)
			output, err := svc.ListBuckets(&s3.ListBucketsInput{})
			return output, err
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "s3:ListBuckets", err)
			}
			if s3Output, ok := output.(*s3.ListBucketsOutput); ok {
				for _, bucket := range s3Output.Buckets {
					utils.PrintResult(debug, "", "s3:ListBuckets", fmt.Sprintf("Found S3 bucket: %s", *bucket.Name), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
