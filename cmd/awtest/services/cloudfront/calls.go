package cloudfront

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
)

type DistributionWithOrigins struct {
	Distribution *cloudfront.DistributionSummary
	Origins      []*cloudfront.Origin
	Region       string
}

var CloudFrontCalls = []types.AWSService{
	{
		Name: "cloudfront:ListDistributions",
		Call: func(sess *session.Session) (interface{}, error) {
			var allDistributionsWithOrigins []DistributionWithOrigins

			originalConfig := sess.Config
			for _, region := range types.Regions {
				regionConfig := &aws.Config{
					Region:      aws.String(region),
					Credentials: originalConfig.Credentials,
				}
				regionSess, err := session.NewSession(regionConfig)
				if err != nil {
					return nil, err
				}
				svc := cloudfront.New(regionSess)
				distributionsOutput, err := svc.ListDistributions(&cloudfront.ListDistributionsInput{})
				if err != nil {
					return nil, err
				}

				for _, distribution := range distributionsOutput.DistributionList.Items {
					distributionWithOrigins := DistributionWithOrigins{
						Distribution: distribution,
						Origins:      distribution.Origins.Items,
						Region:       region,
					}
					allDistributionsWithOrigins = append(allDistributionsWithOrigins, distributionWithOrigins)
				}
			}
			return allDistributionsWithOrigins, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "cloudfront:ListDistributions", err)
			}

			if distributionsWithOrigins, ok := output.([]DistributionWithOrigins); ok {
				for _, distributionWithOrigins := range distributionsWithOrigins {
					fmt.Println()
					distributionID := *distributionWithOrigins.Distribution.Id
					utils.PrintResult(debug, "", "cloudfront:ListDistributions", fmt.Sprintf("Found Distribution: %s", distributionID), nil)

					if len(distributionWithOrigins.Origins) > 0 {
						for _, origin := range distributionWithOrigins.Origins {
							utils.PrintResult(debug, "", "cloudfront:ListOrigins", fmt.Sprintf("Found Origin: %s (%s)", *origin.Id, distributionID), nil)
						}
					}
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	// You can add more methods here similar to the one above
}
