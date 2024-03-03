package ivsrealtime

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ivsrealtime"
)

var IvsRealtimeCalls = []types.AWSService{
	{
		Name: "ivsRealtime:ListStages",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := ivsrealtime.New(sess)
			input := &ivsrealtime.ListStagesInput{}
			return svc.ListStages(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "ivsRealtime:ListStages", err)
			}
			if stagesOutput, ok := output.(*ivsrealtime.ListStagesOutput); ok {
				for _, stage := range stagesOutput.Stages {
					utils.PrintResult(debug, "", "ivs-realtime:ListStages", fmt.Sprintf("Stage: %s", *stage.Name), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
