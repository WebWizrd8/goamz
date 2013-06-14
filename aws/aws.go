//
// goamz - Go packages to interact with the Amazon Web Services.
//
//   https://wiki.ubuntu.com/goamz
//
// Copyright (c) 2011 Canonical Ltd.
//
// Written by Gustavo Niemeyer <gustavo.niemeyer@canonical.com>
//
package aws

import (
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Defines the valid signers
const (
	V2Signature = iota
	V4Signature = iota
)

// Defines the service endpoint and correct Signer implementation to use
// to sign requests for this endpoint
type ServiceInfo struct {
	Endpoint string
	Signer   uint
}

// Region defines the URLs where AWS services may be accessed.
//
// See http://goo.gl/d8BP1 for more details.
type Region struct {
	Name                 string // the canonical name of this region.
	EC2Endpoint          string
	S3Endpoint           string
	S3BucketEndpoint     string // Not needed by AWS S3. Use ${bucket} for bucket name.
	S3LocationConstraint bool   // true if this region requires a LocationConstraint declaration.
	S3LowercaseBucket    bool   // true if the region requires bucket names to be lower case.
	SDBEndpoint          string
	SNSEndpoint          string
	SQSEndpoint          string
	IAMEndpoint          string
	DynamoDBEndpoint     string
	CloudWatchEndpoint   ServiceInfo
}

var Regions = map[string]Region{
	APNortheast.Name:  APNortheast,
	APSoutheast.Name:  APSoutheast,
	APSoutheast2.Name: APSoutheast2,
	EUWest.Name:       EUWest,
	USEast.Name:       USEast,
	USWest.Name:       USWest,
	USWest2.Name:      USWest2,
	SAEast.Name:       SAEast,
}

// Designates a signer interface suitable for signing AWS requests, params
// should be appropriately encoded for the request before signing.
//
// A signer should be initialized with Auth and the appropriate endpoint.
type Signer interface {
	Sign(method, path string, params map[string]string)
}

// An AWS Service interface with the API to query the AWS service
//
// Supplied as an easy way to mock out service calls during testing.
type AWSService interface {
	// Queries the AWS service at a given method/path with the params and
	// returns an http.Response and error
	Query(method, path string, params map[string]string) (*http.Response, error)
	// Builds an error given an XML payload in the http.Response, can be used
	// to process an error if the status code is not 200 for example.
	BuildError(r *http.Response) error
}

// Implements a Server Query/Post API to easily query AWS services and build
// errors when desired
type Service struct {
	service ServiceInfo
	signer  Signer
}

// Create a new AWS server to handle making requests
func NewService(auth Auth, service ServiceInfo) *Service {
	var s Signer
	if service.Signer == V2Signature {
		s = &V2Signer{auth, service}
	}
	return &Service{signer: s}
}

func (s *Service) Query(method, path string, params map[string]string) (resp *http.Response, err error) {
	params["Timestamp"] = time.Now().UTC().Format(time.RFC3339)
	u, err := url.Parse(s.service.Endpoint)
	if err != nil {
		return nil, err
	}
	u.Path = path

	s.signer.Sign(method, path, params)
	if method == "GET" {
		u.RawQuery = multimap(params).Encode()
		resp, err = http.Get(u.String())
	} else if method == "POST" {
		resp, err = http.PostForm(u.String(), multimap(params))
	}
	return
}

func (s *Service) BuildError(r *http.Response) error {
	errors := xmlErrors{}
	xml.NewDecoder(r.Body).Decode(&errors)
	var err Error
	if len(errors.Errors) > 0 {
		err = errors.Errors[0]
	}
	err.RequestId = errors.RequestId
	err.StatusCode = r.StatusCode
	if err.Message == "" {
		err.Message = r.Status
	}
	return &err
}

type Error struct {
	StatusCode int
	Code       string
	Message    string
	RequestId  string
}

func (err *Error) Error() string {
	return err.Message
}

type xmlErrors struct {
	RequestId string
	Errors    []Error `xml:"Errors>Error"`
}

type Auth struct {
	AccessKey, SecretKey string
}

// ResponseMetadata
type ResponseMetadata struct {
	RequestId string // A unique ID for tracking the request
}

type BaseResponse struct {
	ResponseMetadata ResponseMetadata
}

type AWSError struct {
	Type    string
	Code    string
	Message string
}

type ErrorResponse struct {
	Error     AWSError
	RequestId string // A unique ID for tracking the request
}

var unreserved = make([]bool, 128)
var hex = "0123456789ABCDEF"

func init() {
	// RFC3986
	u := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567890-_.~"
	for _, c := range u {
		unreserved[c] = true
	}
}

func multimap(p map[string]string) url.Values {
	q := make(url.Values, len(p))
	for k, v := range p {
		q[k] = []string{v}
	}
	return q
}

// EnvAuth creates an Auth based on environment information.
// The AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment
// variables are used.
func EnvAuth() (auth Auth, err error) {
	auth.AccessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	auth.SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	if auth.AccessKey == "" {
		err = errors.New("AWS_ACCESS_KEY_ID not found in environment")
	}
	if auth.SecretKey == "" {
		err = errors.New("AWS_SECRET_ACCESS_KEY not found in environment")
	}
	return
}

// Encode takes a string and URI-encodes it in a way suitable
// to be used in AWS signatures.
func Encode(s string) string {
	encode := false
	for i := 0; i != len(s); i++ {
		c := s[i]
		if c > 127 || !unreserved[c] {
			encode = true
			break
		}
	}
	if !encode {
		return s
	}
	e := make([]byte, len(s)*3)
	ei := 0
	for i := 0; i != len(s); i++ {
		c := s[i]
		if c > 127 || !unreserved[c] {
			e[ei] = '%'
			e[ei+1] = hex[c>>4]
			e[ei+2] = hex[c&0xF]
			ei += 3
		} else {
			e[ei] = c
			ei += 1
		}
	}
	return string(e[:ei])
}
