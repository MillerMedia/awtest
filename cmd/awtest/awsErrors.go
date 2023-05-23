package main

var awsErrorMessages = map[string]string{
	"UnauthorizedOperation": "You don't have permission to perform this operation.",
	"InvalidAccessKeyId":    "Invalid access key. Aborting scan.",
	"AccessDeniedException": "Access denied to this service.",
	"InvalidClientTokenId":  "The security token included in the request is invalid.",
}

type InvalidKeyError struct {
	message string
}

func (e *InvalidKeyError) Error() string {
	return e.message
}
