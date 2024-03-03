package transcribe

import (
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/transcribeservice"
)

var TranscribeCalls = []types.AWSService{
	{
		Name: "transcribe:ListTranscriptionJobs",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := transcribeservice.New(sess)
			input := &transcribeservice.ListTranscriptionJobsInput{}
			return svc.ListTranscriptionJobs(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "transcribe:ListTranscriptionJobs", err)
			}

			if jobsOutput, ok := output.(*transcribeservice.ListTranscriptionJobsOutput); ok {
				for _, job := range jobsOutput.TranscriptionJobSummaries {
					utils.PrintResult(debug, "", "transcribe:ListTranscriptionJobs", *job.TranscriptionJobName, nil)
				}
			}

			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "transcribe:ListVocabularies",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := transcribeservice.New(sess)
			input := &transcribeservice.ListVocabulariesInput{}
			return svc.ListVocabularies(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "transcribe:ListVocabularies", err)
			}

			if vocabOutput, ok := output.(*transcribeservice.ListVocabulariesOutput); ok {
				for _, vocab := range vocabOutput.Vocabularies {
					utils.PrintResult(debug, "", "transcribe:ListVocabularies", *vocab.VocabularyName, nil)
				}
			}

			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "transcribe:ListLanguageModels",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := transcribeservice.New(sess)
			input := &transcribeservice.ListLanguageModelsInput{}
			return svc.ListLanguageModels(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "transcribe:ListLanguageModels", err)
			}

			if langModelOutput, ok := output.(*transcribeservice.ListLanguageModelsOutput); ok {
				for _, langModel := range langModelOutput.Models {
					utils.PrintResult(debug, "", "transcribe:ListLanguageModels", *langModel.ModelName, nil)
				}
			}

			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "transcribe:StartTranscriptionJob",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := transcribeservice.New(sess)
			input := &transcribeservice.StartTranscriptionJobInput{}
			return svc.StartTranscriptionJob(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "transcribe:StartTranscriptionJob", err)
			}

			if jobOutput, ok := output.(*transcribeservice.StartTranscriptionJobOutput); ok {
				utils.PrintResult(debug, "", "transcribe:StartTranscriptionJob", *jobOutput.TranscriptionJob.TranscriptionJobName, nil)
			}

			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
