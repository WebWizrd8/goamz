package s3

import (
	"launchpad.net/goamz/aws"
)

var originalStrategy = attempts

func SetAttemptStrategy(s *aws.AttemptStrategy) {
	if s == nil {
		attempts = originalStrategy
	} else {
		attempts = *s
	}
}

func Sign(auth aws.Auth, method, path string, params, headers map[string][]string) {
	sign(auth, method, path, params, headers)
}
