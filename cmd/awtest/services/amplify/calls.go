package amplify

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/amplify"
)

type AppWithBranches struct {
	App      *amplify.App
	Branches []*amplify.Branch
	Region   string
}

var AmplifyCalls = []types.AWSService{
	{
		Name: "amplify:ListApps",
		Call: func(sess *session.Session) (interface{}, error) {
			var allAppsWithBranches []AppWithBranches

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
				svc := amplify.New(regionSess)
				appsOutput, err := svc.ListApps(&amplify.ListAppsInput{})
				if err != nil {
					return nil, err
				}

				for _, app := range appsOutput.Apps {
					branchesOutput, err := svc.ListBranches(&amplify.ListBranchesInput{AppId: app.AppId})
					if err != nil {
						return nil, err
					}

					appWithBranches := AppWithBranches{
						App:      app,
						Branches: branchesOutput.Branches,
						Region:   region,
					}
					allAppsWithBranches = append(allAppsWithBranches, appWithBranches)
				}
			}
			return allAppsWithBranches, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "amplify:ListApps", err)
			}

			if appsWithBranches, ok := output.([]AppWithBranches); ok {
				for _, appWithBranches := range appsWithBranches {
					fmt.Println()
					appName := *appWithBranches.App.Name
					utils.PrintResult(debug, "", "amplify:ListApps", fmt.Sprintf("Found Amplify App: %s", appName), nil)

					if len(appWithBranches.Branches) > 0 {
						for _, branch := range appWithBranches.Branches {
							utils.PrintResult(debug, "", "amplify:ListBranches", fmt.Sprintf("Found Branch: %s (%s)", *branch.BranchName, appName), nil)
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
