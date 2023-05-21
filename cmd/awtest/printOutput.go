package main

import (
	"fmt"
	"github.com/logrusorgru/aurora"
)

// Determines the severity based on the error received. For simplicity, we'll classify service
// call errors as high severity, and successful calls as info severity.
func determineSeverity(err error) string {
	if err != nil {
		return "high"
	} else {
		return "info"
	}
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

func printResult(debug bool, moduleName string, method string, result string, err error) {
	severity := determineSeverity(err)

	if moduleName == "" {
		moduleName = DefaultModuleName
	}

	message := colorizeMessage(moduleName, method, severity, result)
	if debug {
		fmt.Println(message)
	}
}
