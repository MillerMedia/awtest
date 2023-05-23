package types

import "github.com/aws/aws-sdk-go/aws/session"

const DefaultModuleName = "AWTest"
const InvalidAccessKeyId = "InvalidAccessKeyId"
const InvalidClientTokenId = "InvalidClientTokenId"

var Regions = []string{
	"us-east-1",
	"us-east-2",
	"us-west-1",
	"us-west-2",
	// Add more regions as needed...
}

type AWSService struct {
	Name       string
	Call       func(*session.Session) (interface{}, error)
	Process    func(interface{}, error, bool) error
	ModuleName string
}

var AwsErrorMessages = map[string]string{
	"UnauthorizedOperation": "You don't have permission to perform this operation.",
	"InvalidAccessKeyId":    "Invalid access key. Aborting scan.",
	"AccessDeniedException": "Access denied to this service.",
	"InvalidClientTokenId":  "The security token included in the request is invalid. Aborting scan.",
}

type InvalidKeyError struct {
	Message string
}

func (e *InvalidKeyError) Error() string {
	return e.Message
}
