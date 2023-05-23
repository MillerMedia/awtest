# AWTest

AWTest is a tool for pentesting found AWS credentials.

## Features
* Performs list operations on various AWS services to check permissions for AWS access key id / secret access key pairs.

## Installation

Run the following command to install the latest version -

```sh
go install -v github.com/MillerMedia/awtest/cmd/awtest@latest
```

<details>
  <summary>Brew</summary>

  ```sh
  brew tap MillerMedia/awtest
  brew install awtest
  ```

</details>

## Usage

#### Example command
```bash
awtest --aki=YourAccessKeyID --sak=YourSecretAccessKey
```

#### Example Output

```bash
[AWTest] [user-id] [info] AKIABCDEFGHIJKLMNO
[AWTest] [account-number] [info] 123456789012
[AWTest] [iam-arn] [info] arn:aws:iam::123456789012:user/exampleUser
[AWTest] [s3:ListBuckets] [info] Found S3 bucket: example-bucket-1
[AWTest] [s3:ListBuckets] [info] Found S3 bucket: example-bucket-2
[AWTest] [s3:ListBuckets] [info] Found S3 bucket: example-bucket-3
[AWTest] [ec2:DescribeInstances] [info] Found EC2 instance: i-0abcdef1234567890
[AWTest] [iam:ListUsers] [info] Found IAM user: exampleUser1
[AWTest] [iam:ListUsers] [info] Found IAM user: exampleUser2
[AWTest] [sns:ListTopics] [info] Found SNS topic: arn:aws:sns:us-east-1:123456789012:ExampleTopic1
[AWTest] [sns:ListTopics] [info] Found SNS topic: arn:aws:sns:us-east-1:123456789012:ExampleTopic2
[AWTest] [lambda:ListFunctions] [info] Found Lambda function: exampleFunction1
[AWTest] [lambda:ListFunctions] [info] Found Lambda function: exampleFunction2
[AWTest] [cloudwatch:DescribeAlarms] [info] Found CloudWatch alarm: exampleAlarm1
[AWTest] [cloudwatch:DescribeAlarms] [info] Found CloudWatch alarm: exampleAlarm2
[AWTest] [appsync:ListGraphqlApis] [info] Error: Access denied to this service.
[AWTest] [apigateway:RestApis] [info] Error: Access denied to this service.
```

#### Flags/Options
```angular2html
-access-key-id string
    AWS Access Key ID
-aki string
    Abbreviated AWS Access Key ID
-debug
    Enable debug mode
-region string
    AWS Region (default "us-west-2")
-sak string
    Abbreviated AWS Secret Access Key
-secret-access-key string
    AWS Secret Access Key

```

## Contributing

I welcome contributions from the community! If you have any suggestions, bug reports, or ideas for improvement, feel free to open an issue or submit a pull request.

## Support the project

If you find this project helpful and would like to support its development, please consider donating:

[![Buy me a coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/yOd1JU9MQe)

## License

This project is licensed under the MIT License.