package iot

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iot"
)

var IoTCalls = []types.AWSService{
	{
		Name: "iot:GetStatistics",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := iot.New(sess)
			input := &iot.GetStatisticsInput{}
			return svc.GetStatistics(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "iot:GetStatistics", err)
			}
			if statistics, ok := output.(*iot.GetStatisticsOutput); ok {
				utils.PrintResult(debug, "", "iot:GetStatistics", fmt.Sprintf("Statistics: %v", statistics.Statistics), nil)
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "iot:ListThings",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := iot.New(sess)
			input := &iot.ListThingsInput{}
			return svc.ListThings(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "iot:ListThings", err)
			}
			if things, ok := output.(*iot.ListThingsOutput); ok {
				for _, thing := range things.Things {
					utils.PrintResult(debug, "", "iot:ListThings", fmt.Sprintf("Found Thing: %s", *thing.ThingName), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "iot:ListPolicies",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := iot.New(sess)
			input := &iot.ListPoliciesInput{}
			return svc.ListPolicies(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "iot:ListPolicies", err)
			}
			if policies, ok := output.(*iot.ListPoliciesOutput); ok {
				for _, policy := range policies.Policies {
					utils.PrintResult(debug, "", "iot:ListPolicies", fmt.Sprintf("Found Policy: %s", *policy.PolicyName), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "iot:ListCertificates",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := iot.New(sess)
			input := &iot.ListCertificatesInput{}
			return svc.ListCertificates(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "iot:ListCertificates", err)
			}
			if certificates, ok := output.(*iot.ListCertificatesOutput); ok {
				for _, cert := range certificates.Certificates {
					utils.PrintResult(debug, "", "iot:ListCertificates", fmt.Sprintf("Found Certificate: %s", *cert.CertificateArn), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
