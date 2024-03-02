package route53

import (
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

var Route53Calls = []types.AWSService{
	{
		Name: "route53:ListHostedZones",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := route53.New(sess)
			output, err := svc.ListHostedZones(&route53.ListHostedZonesInput{})

			if err != nil {
				return nil, err
			} else {
				return output, nil
			}
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "route53:ListHostedZones", err)
			}
			if zones, ok := output.(*route53.ListHostedZonesOutput); ok {
				for _, zone := range zones.HostedZones {
					utils.PrintResult(debug, "", "route53:ListHostedZones", *zone.Name, nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "route53:ListHealthChecks",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := route53.New(sess)
			output, err := svc.ListHealthChecks(&route53.ListHealthChecksInput{})

			if err != nil {
				return nil, err
			} else {
				return output, nil
			}
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return err
			}
			if checks, ok := output.(*route53.ListHealthChecksOutput); ok {
				for _, check := range checks.HealthChecks {
					utils.PrintResult(debug, "", "route53:ListHealthChecks", *check.HealthCheckConfig.IPAddress, nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
