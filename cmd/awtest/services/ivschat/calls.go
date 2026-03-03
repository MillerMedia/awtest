package ivschat

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ivschat"
)

var IvsChatCalls = []types.AWSService{
	{
		Name: "ivsChat:ListRooms",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := ivschat.New(sess)
			input := &ivschat.ListRoomsInput{}
			return svc.ListRooms(input)
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "ivsChat:ListRooms", err)
			}
			if roomsOutput, ok := output.(*ivschat.ListRoomsOutput); ok {
				for _, room := range roomsOutput.Rooms {
					utils.PrintResult(debug, "", "ivsChat:ListRooms", fmt.Sprintf("Room: %s", *room.Name), nil)
				}
			}
			return nil
		},
		ModuleName: types.DefaultModuleName,
	},
}
