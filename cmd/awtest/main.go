package main

import (
	"flag"
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/services"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
	"strings"
)

const Version = "v0.3.0"

func main() {
	fmt.Println("     /\\ \\        / /__   __|      | |")
	fmt.Println("    /  \\ \\  /\\  / /   | | ___  ___| |_")
	fmt.Println("   / /\\ \\ \\/  \\/ /    | |/ _ \\/ __| __|")
	fmt.Println("  / ____ \\  /\\  /     | |  __/\\__ \\ |_")
	fmt.Println(" /_/    \\_\\/  \\/      |_|\\___||___/\\__|")
	fmt.Println("----------------------------------------")
	fmt.Println("Version:", Version)
	fmt.Println("----------------------------------------")

	awsAccessKeyID := flag.String("access-key-id", "", "AWS Access Key ID")
	awsSecretAccessKey := flag.String("secret-access-key", "", "AWS Secret Access Key")
	awsSessionToken := flag.String("session-token", "", "AWS Session Token (optional)")
	awsRegion := flag.String("region", "us-west-2", "AWS Region")

	awsAccessKeyIDAbbr := flag.String("aki", "", "Abbreviated AWS Access Key ID")
	awsSecretAccessKeyAbbr := flag.String("sak", "", "Abbreviated AWS Secret Access Key")
	awsSessionTokenAbbr := flag.String("st", "", "Abbreviated AWS Session Token")

	debug := flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()

	if *awsAccessKeyIDAbbr != "" {
		awsAccessKeyID = awsAccessKeyIDAbbr
	}
	if *awsSecretAccessKeyAbbr != "" {
		awsSecretAccessKey = awsSecretAccessKeyAbbr
	}
	if *awsSessionTokenAbbr != "" {
		awsSessionToken = awsSessionTokenAbbr
	}

	if *awsAccessKeyID == "" {
		*awsAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if *awsSecretAccessKey == "" {
		*awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	if *awsSessionToken == "" {
		*awsSessionToken = os.Getenv("AWS_SESSION_TOKEN")
	}

	// Check if AWS_PROFILE is set and no parameters were provided
	awsProfile := os.Getenv("AWS_PROFILE")
	var sess *session.Session
	var err error

	if *awsAccessKeyID == "" || *awsSecretAccessKey == "" {
		if awsProfile != "" {
			// Use the AWS_PROFILE if set and no access keys are provided
			sess, err = session.NewSessionWithOptions(session.Options{
				Profile:           awsProfile,
				SharedConfigState: session.SharedConfigEnable,
				Config: aws.Config{
					Region: aws.String(*awsRegion),
				},
			})
		} else {
			// Fall back to default shared config if no profile is set
			sess, err = session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
				Config: aws.Config{
					Region: aws.String(*awsRegion),
				},
			})
		}
		if err != nil {
			fmt.Println("Failed to create session with shared config: ", err)
			return
		}
	} else {
		// Check if the access key starts with 'ASIA'
		if strings.HasPrefix(*awsAccessKeyID, "ASIA") && *awsSessionToken != "" {
			// Use the session token as well
			sess, _ = session.NewSession(&aws.Config{
				Region:      aws.String(*awsRegion),
				Credentials: credentials.NewStaticCredentials(*awsAccessKeyID, *awsSecretAccessKey, *awsSessionToken),
			})
		} else {
			// If keys are provided, use them to create session without session token
			sess, _ = session.NewSession(&aws.Config{
				Region:      aws.String(*awsRegion),
				Credentials: credentials.NewStaticCredentials(*awsAccessKeyID, *awsSecretAccessKey, ""),
			})
		}
	}

	if *debug {
		fmt.Println("Debug mode enabled")
		fmt.Println("-----------------------------")
		fmt.Println("Using the following AWS configuration:")

		// Get credentials from session if they are not provided explicitly
		creds, err := sess.Config.Credentials.Get()
		if err != nil {
			fmt.Println("Failed to retrieve credentials from session: ", err)
		} else {
			fmt.Println("Access Key ID:", creds.AccessKeyID)
			fmt.Println("Secret Access Key:", utils.MaskSecret(creds.SecretAccessKey))
			if creds.SessionToken != "" {
				fmt.Println("Session Token:", utils.MaskSecret(creds.SessionToken))
			}
		}

		if awsProfile != "" {
			fmt.Println("Profile:", awsProfile)
		}
		fmt.Println("Region:", *awsRegion)
		fmt.Println("-----------------------------")
	}

	for _, service := range services.AllServices() {
		output, err := service.Call(sess)
		if err := service.Process(output, err, *debug); err != nil {
			// Check if the error is InvalidKeyError and exit if so
			if _, ok := err.(*types.InvalidKeyError); ok {
				os.Exit(1)
			}
		}
	}
}
