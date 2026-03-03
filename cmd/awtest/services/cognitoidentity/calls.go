package cognitoidentity

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
)

var CognitoIdentityCalls = []types.AWSService{
	{
		Name: "cognitoidentity:ListIdentityPools",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := cognitoidentity.New(sess)
			input := &cognitoidentity.ListIdentityPoolsInput{
				MaxResults: aws.Int64(60), // You can adjust this as needed
			}
			output, err := svc.ListIdentityPools(input)
			if err != nil {
				return nil, err
			}
			return output.IdentityPools, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "cognitoidentity:ListIdentityPools", err)
			}
			if identityPools, ok := output.([]*cognitoidentity.IdentityPoolShortDescription); ok {
				for _, pool := range identityPools {
					utils.PrintResult(debug, "", "cognitoidentity:ListIdentityPools", fmt.Sprintf("Found Identity Pool: %s (%s)", *pool.IdentityPoolName, *pool.IdentityPoolId), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
