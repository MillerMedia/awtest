package codepipeline

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codepipeline"
)

var CodePipelineCalls = []types.AWSService{
	{
		Name: "codepipeline:ListPipelines",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := codepipeline.New(sess)
			input := &codepipeline.ListPipelinesInput{}
			return svc.ListPipelines(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "codepipeline:ListPipelines", err)
			}
			if pipelines, ok := output.(*codepipeline.ListPipelinesOutput); ok {
				for _, pipeline := range pipelines.Pipelines {
					utils.PrintResult(debug, "", "codepipeline:ListPipelines", fmt.Sprintf("Pipeline: %s", *pipeline.Name), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
