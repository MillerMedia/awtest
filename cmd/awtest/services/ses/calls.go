package ses

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

var SESCalls = []types.AWSService{
	{
		Name: "ses:ListIdentities",
		Call: func(sess *session.Session) (interface{}, error) {
			originalConfig := sess.Config
			var allIdentities []string
			for _, region := range types.Regions {
				regionConfig := &aws.Config{
					Region:      aws.String(region),
					Credentials: originalConfig.Credentials,
				}
				regionSess, err := session.NewSession(regionConfig)
				if err != nil {
					return nil, err
				}
				svc := ses.New(regionSess)
				output, err := svc.ListIdentities(&ses.ListIdentitiesInput{})
				if err != nil {
					return nil, err
				}
				allIdentities = append(allIdentities, aws.StringValueSlice(output.Identities)...)
			}
			return allIdentities, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "ses:ListIdentities", err)
			}
			if identities, ok := output.([]string); ok {
				for _, identity := range identities {
					utils.PrintResult(debug, "", "ses:ListIdentities", fmt.Sprintf("Found Identity: %s", identity), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
