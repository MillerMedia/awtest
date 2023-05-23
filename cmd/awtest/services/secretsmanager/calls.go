package secretsmanager

import (
	"fmt"
	"github.com/MillerMedia/AWTest/cmd/awtest/types"
	"github.com/MillerMedia/AWTest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

var SecretsManagerCalls = []types.AWSService{
	{
		Name: "secretsmanager:ListSecrets",
		Call: func(sess *session.Session) (interface{}, error) {
			var allSecrets []*secretsmanager.SecretListEntry
			for _, region := range types.Regions {
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
				return utils.HandleAWSError(debug, "secretsmanager:ListSecrets", err)
			}
			if secrets, ok := output.([]*secretsmanager.SecretListEntry); ok {
				for _, secret := range secrets {
					utils.PrintResult(debug, "", "secretsmanager:ListSecrets", fmt.Sprintf("Found secret: %s", *secret.Name), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
