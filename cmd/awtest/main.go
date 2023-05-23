package main

import (
	"flag"
	"fmt"
	"github.com/MillerMedia/AWTest/cmd/awtest/services"
	"github.com/MillerMedia/AWTest/cmd/awtest/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
)

func main() {
	awsAccessKeyID := flag.String("access-key-id", "", "AWS Access Key ID")
	awsSecretAccessKey := flag.String("secret-access-key", "", "AWS Secret Access Key")
	awsRegion := flag.String("region", "us-west-2", "AWS Region")

	awsAccessKeyIDAbbr := flag.String("aki", "", "Abbreviated AWS Access Key ID")
	awsSecretAccessKeyAbbr := flag.String("sak", "", "Abbreviated AWS Secret Access Key")

	debug := flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()

	if *awsAccessKeyIDAbbr != "" {
		awsAccessKeyID = awsAccessKeyIDAbbr
	}
	if *awsSecretAccessKeyAbbr != "" {
		awsSecretAccessKey = awsSecretAccessKeyAbbr
	}

	// If keys are not provided through flags, try to get them from the environment
	if *awsAccessKeyID == "" {
		*awsAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if *awsSecretAccessKey == "" {
		*awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}

	var sess *session.Session
	var err error

	// If keys are not provided through flags or environment, fallback to shared config file
	if *awsAccessKeyID == "" || *awsSecretAccessKey == "" {
		sess, err = session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Config:            aws.Config{Region: aws.String(*awsRegion)},
		})
		if err != nil {
			fmt.Println("Failed to create session with shared config: ", err)
			return
		}
	} else {
		// If keys are provided, use them to create session
		sess, _ = session.NewSession(&aws.Config{
			Region:      aws.String(*awsRegion),
			Credentials: credentials.NewStaticCredentials(*awsAccessKeyID, *awsSecretAccessKey, ""),
		})
	}

	if *debug {
		fmt.Println("Debug mode enabled")
		fmt.Println("Using the following AWS configuration:")
		fmt.Println("Access Key ID:", *awsAccessKeyID)
		fmt.Println("Region:", *awsRegion)
	}

	for _, service := range services.AllServices() {
		output, err := service.Call(sess)
		if err := service.Process(output, err, *debug); err != nil {
			// Check if the error is InvalidKeyError and exit if so
			if _, ok := err.(*types.InvalidKeyError); ok {
				os.Exit(1)
			}
			// Otherwise, just continue to the next service
		}
	}
}
