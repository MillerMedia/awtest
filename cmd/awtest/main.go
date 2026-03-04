package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MillerMedia/awtest/cmd/awtest/formatters"
	"github.com/MillerMedia/awtest/cmd/awtest/services"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

const Version = "v0.3.0"

func main() {
	awsAccessKeyID := flag.String("access-key-id", "", "AWS Access Key ID")
	awsSecretAccessKey := flag.String("secret-access-key", "", "AWS Secret Access Key")
	awsSessionToken := flag.String("session-token", "", "AWS Session Token (optional)")
	awsRegion := flag.String("region", "us-west-2", "AWS Region")

	awsAccessKeyIDAbbr := flag.String("aki", "", "Abbreviated AWS Access Key ID")
	awsSecretAccessKeyAbbr := flag.String("sak", "", "Abbreviated AWS Secret Access Key")
	awsSessionTokenAbbr := flag.String("st", "", "Abbreviated AWS Session Token")

	outputFormat := flag.String("format", "text", "Output format: text, json, yaml, csv, table")
	outputFile := flag.String("output-file", "", "Write output to file instead of stdout")
	quiet := flag.Bool("quiet", false, "Suppress informational messages, show only findings")

	debug := flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()

	utils.Quiet = *quiet

	if !*quiet {
		fmt.Println("     /\\ \\        / /__   __|      | |")
		fmt.Println("    /  \\ \\  /\\  / /   | | ___  ___| |_")
		fmt.Println("   / /\\ \\ \\/  \\/ /    | |/ _ \\/ __| __|")
		fmt.Println("  / ____ \\  /\\  /     | |  __/\\__ \\ |_")
		fmt.Println(" /_/    \\_\\/  \\/      |_|\\___||___/\\__|")
		fmt.Println("----------------------------------------")
		fmt.Println("Version:", Version)
		fmt.Println("----------------------------------------")
	}

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

	// Validate format flag early
	formatter, err := getFormatter(*outputFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	startTime := time.Now()

	var results []types.ScanResult
	for _, service := range services.AllServices() {
		if !*quiet {
			fmt.Fprintf(os.Stderr, "Scanning %s...\n", service.Name)
		}
		output, err := service.Call(sess)
		serviceResults := service.Process(output, err, *debug)
		results = append(results, serviceResults...)
	}

	summary := types.GenerateSummary(results, startTime)

	// For text format to stdout, results are already printed by Process() methods
	// Unless quiet mode is set — then we need to use the formatter
	if *outputFormat == "text" && *outputFile == "" && !*quiet {
		printTextSummary(summary)
		return
	}

	// For all other cases, use formatter
	// Quiet mode suppresses summary (AC6) — use Format() instead of FormatWithSummary()
	var formatted string
	if *quiet {
		formatted, err = formatter.Format(results)
	} else {
		formatted, err = formatter.FormatWithSummary(results, summary)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, []byte(formatted), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Output written to %s\n", *outputFile)
	} else {
		fmt.Print(formatted)
	}
}

// printTextSummary prints a scan summary to stderr for the default text+stdout path.
func printTextSummary(summary types.ScanSummary) {
	fmt.Fprintf(os.Stderr, "========================================\n")
	fmt.Fprintf(os.Stderr, "Scan Summary\n")
	fmt.Fprintf(os.Stderr, "========================================\n")
	fmt.Fprintf(os.Stderr, "Timestamp:          %s\n", summary.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(os.Stderr, "Duration:           %s\n", summary.ScanDuration)
	fmt.Fprintf(os.Stderr, "Total Services:     %d\n", summary.TotalServices)
	fmt.Fprintf(os.Stderr, "Accessible:         %d\n", summary.AccessibleServices)
	fmt.Fprintf(os.Stderr, "Access Denied:      %d\n", summary.AccessDeniedServices)
	fmt.Fprintf(os.Stderr, "Resources Found:    %d\n", summary.TotalResources)
	fmt.Fprintf(os.Stderr, "========================================\n")
}

// getFormatter returns the appropriate OutputFormatter for the given format string.
func getFormatter(format string) (formatters.OutputFormatter, error) {
	switch strings.ToLower(format) {
	case "text":
		return formatters.NewTextFormatter(), nil
	case "json":
		return formatters.NewJSONFormatter(), nil
	case "yaml":
		return formatters.NewYAMLFormatter(), nil
	case "csv":
		return formatters.NewCSVFormatter(), nil
	case "table":
		return formatters.NewTableFormatter(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s (supported: text, json, yaml, csv, table)", format)
	}
}
