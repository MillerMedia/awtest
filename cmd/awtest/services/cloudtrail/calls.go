package cloudtrail

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
)

var CloudTrailCalls = []types.AWSService{
	{
		Name: "cloudtrail:DescribeTrails",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := cloudtrail.New(sess)
			input := &cloudtrail.DescribeTrailsInput{}
			return svc.DescribeTrails(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "cloudtrail:DescribeTrails", err)
			}
			if trails, ok := output.(*cloudtrail.DescribeTrailsOutput); ok {
				for _, trail := range trails.TrailList {
					utils.PrintResult(debug, "", "cloudtrail:DescribeTrails", fmt.Sprintf("Trail: %s", *trail.Name), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "cloudtrail:ListTrails",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := cloudtrail.New(sess)
			input := &cloudtrail.ListTrailsInput{}
			return svc.ListTrails(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "cloudtrail:ListTrails", err)
			}
			if trailsOutput, ok := output.(*cloudtrail.ListTrailsOutput); ok {
				for _, trail := range trailsOutput.Trails {
					utils.PrintResult(debug, "", "cloudtrail:ListTrails", fmt.Sprintf("Trail: %s", *trail.Name), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
