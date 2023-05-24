package glacier

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"
)

var GlacierCalls = []types.AWSService{
	{
		Name: "glacier:ListVaults",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := glacier.New(sess)
			output, err := svc.ListVaults(&glacier.ListVaultsInput{})
			if err != nil {
				return nil, err
			}
			return output.VaultList, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "glacier:ListVaults", err)
			}
			if vaults, ok := output.([]*glacier.DescribeVaultOutput); ok {
				for _, vault := range vaults {
					utils.PrintResult(debug, "", "glacier:ListVaults", fmt.Sprintf("Glacier vault: %s", utils.ColorizeItem(*vault.VaultName)), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
