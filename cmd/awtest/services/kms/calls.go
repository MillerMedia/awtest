package kms

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

var KMSCalls = []types.AWSService{
	{
		Name: "kms:ListKeys",
		Call: func(sess *session.Session) (interface{}, error) {
			var allKeys []*kms.KeyListEntry
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := kms.New(sess)
				output, err := svc.ListKeys(&kms.ListKeysInput{})
				if err != nil {
					return nil, err
				}
				allKeys = append(allKeys, output.Keys...)
			}
			return allKeys, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "kms:ListKeys", err)
			}
			if keys, ok := output.([]*kms.KeyListEntry); ok {
				for _, key := range keys {
					utils.PrintResult(debug, "", "kms:ListKeys", fmt.Sprintf("KMS key: %s", utils.ColorizeItem(*key.KeyId)), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
