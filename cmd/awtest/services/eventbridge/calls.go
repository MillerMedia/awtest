package eventbridge

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
)

var EventbridgeCalls = []types.AWSService{
	{
		Name: "eventbridge:ListEventBuses",
		Call: func(sess *session.Session) (interface{}, error) {
			svc := eventbridge.New(sess)
			output, err := svc.ListEventBuses(&eventbridge.ListEventBusesInput{})
			if err != nil {
				return nil, err
			}
			return output.EventBuses, nil
		},
		Process: func(output interface{}, err error, debug bool) error {
			if err != nil {
				return utils.HandleAWSError(debug, "eventbridge:ListEventBuses", err)
			}
			if buses, ok := output.([]*eventbridge.EventBus); ok {
				for _, bus := range buses {
					utils.PrintResult(debug, "", "eventbridge:ListEventBuses", fmt.Sprintf("EventBridge bus: %s", *bus.Name), nil)
				}
			}
			return nil
		},
	},
}
