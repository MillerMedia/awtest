package ivs

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ivs"
)

var IvsCalls = []types.AWSService{
	{
		Name: "ivs:ListChannels",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := ivs.New(sess)
			input := &ivs.ListChannelsInput{}
			return svc.ListChannels(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "ivs:ListChannels", err)
			}
			if channels, ok := output.(*ivs.ListChannelsOutput); ok {
				for _, channel := range channels.Channels {
					utils.PrintResult(debug, "", "ivs:ListChannels", fmt.Sprintf("Channel: %s", *channel.Arn), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "ivs:ListStreams",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := ivs.New(sess)
			input := &ivs.ListStreamsInput{}
			return svc.ListStreams(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "ivs:ListStreams", err)
			}
			if streamsOutput, ok := output.(*ivs.ListStreamsOutput); ok {
				for _, stream := range streamsOutput.Streams {
					utils.PrintResult(debug, "", "ivs:ListStreams", fmt.Sprintf("Stream: %s", *stream.StreamId), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
	{
		Name: "ivs:ListStreamKeys",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := ivs.New(sess)
			input := &ivs.ListStreamKeysInput{}
			return svc.ListStreamKeys(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "ivs:ListStreamKeys", err)
			}
			if streamKeysOutput, ok := output.(*ivs.ListStreamKeysOutput); ok {
				for _, streamKey := range streamKeysOutput.StreamKeys {
					utils.PrintResult(debug, "", "ivs:ListStreamKeys", fmt.Sprintf("StreamKey: %s", *streamKey.Arn), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
