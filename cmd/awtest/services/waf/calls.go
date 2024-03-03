package waf

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/waf"
)

var WafCalls = []types.AWSService{
	{
		Name: "waf:ListWebACLs",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := waf.New(sess)
			input := &waf.ListWebACLsInput{}
			return svc.ListWebACLs(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "waf:ListWebACLs", err)
			}
			if webAcls, ok := output.(*waf.ListWebACLsOutput); ok {
				for _, webAcl := range webAcls.WebACLs {
					utils.PrintResult(debug, "", "waf:ListWebACLs", fmt.Sprintf("WebACL: %s", *webAcl.Name), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
