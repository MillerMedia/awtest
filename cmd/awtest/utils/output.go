package utils

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/logrusorgru/aurora"
)

const (
	ResetColor   = "\033[0m"
	DisplayColor = "\033[33m"
)

// Determines the severity based on the error received. For simplicity, we'll classify service
// call errors as high severity, and successful calls as info severity.
func determineSeverity(err error) string {
	//if err != nil {
	//	return "high"
	//} else {
	//	return "info"
	//}

	return "info"
}

func colorizeMessage(moduleName string, method string, severity string, result string) string {
	moduleNameColored := aurora.BrightGreen(moduleName).String()
	methodColored := aurora.BrightBlue(method).String()
	var severityColored string

	if severity == "high" {
		severityColored = aurora.Red(severity).String()
	} else {
		severityColored = aurora.Blue(severity).String()
	}

	coloredMessage := fmt.Sprintf("[%s] [%s] [%s] %s", moduleNameColored, methodColored, severityColored, result)
	return coloredMessage
}

func PrintResult(debug bool, moduleName string, method string, result string, err error) {
	severity := determineSeverity(err)

	if moduleName == "" {
		moduleName = types.DefaultModuleName
	}

	message := colorizeMessage(moduleName, method, severity, result)

	fmt.Println(message)
}

func HandleAWSError(debug bool, callName string, err error) error {
	if awsErr, ok := err.(awserr.Error); ok {
		prettyMsg, exists := types.AwsErrorMessages[awsErr.Code()]
		if !exists {
			prettyMsg = awsErr.Message()
		}

		if awsErr.Code() == types.InvalidAccessKeyId || awsErr.Code() == types.InvalidClientTokenId {
			PrintResult(debug, "", callName, fmt.Sprintf("Error: %s", prettyMsg), err)
			return &types.InvalidKeyError{prettyMsg}
		} else {
			PrintResult(debug, "", callName, fmt.Sprintf("Error: %s", prettyMsg), err)
		}
	} else {
		PrintResult(debug, "", callName, fmt.Sprintf("Error: %s", err.Error()), err)
	}
	return nil
}

func ColorizeItem(input string) string {
	return DisplayColor + input + ResetColor
}
