<p align="center">
  <img src="https://readme-typing-svg.demolab.com?font=Fira+Code&weight=700&size=28&pause=1000&color=FF9900&center=true&vCenter=true&width=500&lines=AWTest" alt="AWTest" />
</p>

<p align="center">
<pre align="center">
     ___  _      ________        __
    /   || | /| / /_  __/__ ___ / /_
   / /| || |/ |/ / / / / -_|_-&lt;/ __/
  /_/ |_||__/|__/ /_/  \__/___/\__/
</pre>
</p>

<p align="center">
  <strong>AWS Credential Permission Scanner for Security Assessments</strong>
</p>

<p align="center">
  <a href="https://github.com/MillerMedia/awtest/releases/latest"><img src="https://img.shields.io/github/v/release/MillerMedia/awtest?color=ff9900&style=flat-square" alt="Latest Release"></a>
  <a href="https://github.com/MillerMedia/awtest/actions/workflows/test.yml"><img src="https://img.shields.io/github/actions/workflow/status/MillerMedia/awtest/test.yml?label=tests&style=flat-square" alt="Tests"></a>
  <a href="https://github.com/MillerMedia/awtest/blob/master/LICENSE"><img src="https://img.shields.io/github/license/MillerMedia/awtest?style=flat-square" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/MillerMedia/awtest"><img src="https://goreportcard.com/badge/github.com/MillerMedia/awtest?style=flat-square" alt="Go Report Card"></a>
</p>

---

AWTest quickly enumerates the permissions of AWS credentials by performing read-only list/describe operations across **46 AWS services**. Built for pentesters, red teamers, and cloud security assessors.

## Features

- **Broad AWS Coverage** -- 46 services scanned including S3, EC2, IAM, Lambda, EKS, RDS, DynamoDB, and more
- **Multiple Output Formats** -- Text, JSON, YAML, CSV, and table output
- **File Export** -- Write results directly to a file with `--output-file`
- **Service Filtering** -- Include or exclude specific services with `--services` and `--exclude-services`
- **Configurable Timeouts** -- Set scan duration limits with `--timeout`
- **Concurrent Scanning** -- Tune parallelism with `--concurrency`
- **Session Token Support** -- Works with temporary credentials (STS)
- **Cross-Platform** -- Pre-built binaries for macOS, Linux, and Windows

## Installation

### Homebrew (macOS/Linux)

```sh
brew install --cask MillerMedia/tap/awtest
```

### Go Install

Requires Go 1.19+:

```sh
go install github.com/MillerMedia/awtest/cmd/awtest@latest
```

### Binary Download

Download pre-built binaries from [GitHub Releases](https://github.com/MillerMedia/awtest/releases):

| Platform | File |
|---|---|
| macOS (Intel) | `awtest_<version>_darwin_amd64.tar.gz` |
| macOS (Apple Silicon) | `awtest_<version>_darwin_arm64.tar.gz` |
| Linux (amd64) | `awtest_<version>_linux_amd64.tar.gz` |
| Linux (arm64) | `awtest_<version>_linux_arm64.tar.gz` |
| Windows | `awtest_<version>_windows_amd64.zip` |

## Usage

### Scan using current AWS CLI profile

```bash
awtest
```

### Scan with explicit credentials

```bash
awtest --aki=AKIAEXAMPLE --sak=YourSecretAccessKey
```

### Scan with temporary credentials (STS)

```bash
awtest --aki=ASIAEXAMPLE --sak=YourSecretKey --st=YourSessionToken
```

### Output as JSON to a file

```bash
awtest --format=json --output-file=results.json
```

### Scan only specific services

```bash
awtest --services=s3,ec2,iam,lambda
```

### Exclude noisy services

```bash
awtest --exclude-services=cloudwatch,cloudtrail
```

### Example Output

```
[AWTest] [user-id] [info] AKIABCDEFGHIJKLMNO
[AWTest] [account-number] [info] 123456789012
[AWTest] [iam-arn] [info] arn:aws:iam::123456789012:user/exampleUser
[AWTest] [s3:ListBuckets] [info] Found S3 bucket: example-bucket-1
[AWTest] [s3:ListBuckets] [info] Found S3 bucket: example-bucket-2
[AWTest] [ec2:DescribeInstances] [info] Found EC2 instance: i-0abcdef1234567890
[AWTest] [iam:ListUsers] [info] Found IAM user: exampleUser1
[AWTest] [lambda:ListFunctions] [info] Found Lambda function: myFunction
[AWTest] [eks:ListClusters] [info] Found EKS cluster: production
[AWTest] [rds:DescribeDBInstances] [info] Found RDS instance: mydb
[AWTest] [appsync:ListGraphqlApis] [info] Error: Access denied to this service.
```

## Flags

| Flag | Description | Default |
|---|---|---|
| `--aki`, `--access-key-id` | AWS Access Key ID | |
| `--sak`, `--secret-access-key` | AWS Secret Access Key | |
| `--st`, `--session-token` | AWS Session Token | |
| `--region` | AWS Region | `us-west-2` |
| `--format` | Output format: `text`, `json`, `yaml`, `csv`, `table` | `text` |
| `--output-file` | Write output to file | |
| `--services` | Include only specific services (comma-separated) | all |
| `--exclude-services` | Exclude specific services (comma-separated) | none |
| `--timeout` | Maximum scan duration (e.g., `5m`, `300s`) | `5m` |
| `--concurrency` | Number of concurrent service scans | `1` |
| `--quiet` | Suppress info messages, show only findings | `false` |
| `--debug` | Enable debug output | `false` |
| `--version` | Print version and build info | |

## Supported AWS Services (46)

<details>
<summary>Click to expand full service list</summary>

| Service | API Calls |
|---|---|
| Amplify | ListApps |
| API Gateway | GetRestApis |
| AppSync | ListGraphqlApis |
| Batch | DescribeComputeEnvironments |
| Certificate Manager (ACM) | ListCertificates |
| CloudFormation | ListStacks |
| CloudFront | ListDistributions |
| CloudTrail | DescribeTrails |
| CloudWatch | DescribeAlarms |
| CodePipeline | ListPipelines |
| Cognito Identity | ListIdentityPools |
| Cognito User Pools | ListUserPools |
| Config | DescribeConfigRules, DescribeDeliveryChannels |
| DynamoDB | ListTables |
| EC2 | DescribeInstances |
| ECS | ListClusters |
| EFS | DescribeFileSystems |
| EKS | ListClusters |
| ElastiCache | DescribeCacheClusters |
| Elastic Beanstalk | DescribeApplications |
| EventBridge | ListEventBuses |
| Fargate | ListTasks |
| Glacier | ListVaults |
| Glue | GetDatabases |
| IAM | ListUsers, ListRoles, ListGroups |
| IoT | ListThings |
| IVS | ListChannels |
| IVS Chat | ListRooms |
| IVS Realtime | ListStages |
| KMS | ListKeys |
| Lambda | ListFunctions |
| RDS | DescribeDBInstances |
| Redshift | DescribeClusters |
| Rekognition | ListCollections |
| Route53 | ListHostedZones |
| S3 | ListBuckets |
| Secrets Manager | ListSecrets |
| SES | ListIdentities |
| SNS | ListTopics |
| SQS | ListQueues |
| Step Functions | ListStateMachines |
| STS | GetCallerIdentity |
| Systems Manager (SSM) | DescribeParameters |
| Transcribe | ListTranscriptionJobs |
| VPC | DescribeVpcs, DescribeSubnets, DescribeSecurityGroups |
| WAF | ListWebACLs |

</details>

## Contributing

Contributions are welcome! If you have suggestions, bug reports, or ideas for improvement, feel free to open an issue or submit a pull request.

## Support the Project

If you find this project helpful, please consider supporting its development:

[![Buy me a coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/yOd1JU9MQe)

## License

This project is licensed under the [MIT License](LICENSE).
