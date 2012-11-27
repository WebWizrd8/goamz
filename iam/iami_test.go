package iam_test

import (
	"flag"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/iam"
	. "launchpad.net/gocheck"
	"net/url"
)

var amazon = flag.Bool("amazon", false, "Enable tests against amazon server")

// AmazonServer represents an Amazon AWS server.
type AmazonServer struct {
	auth aws.Auth
}

func (s *AmazonServer) SetUp(c *C) {
	auth, err := aws.EnvAuth()
	if err != nil {
		c.Fatal(err)
	}
	s.auth = auth
}

var _ = Suite(&AmazonClientSuite{})

// AmazonClientSuite tests the client against a live AWS server.
type AmazonClientSuite struct {
	srv AmazonServer
	ClientTests
}

func (s *AmazonClientSuite) SetUpSuite(c *C) {
	if !*amazon {
		c.Skip("AmazonClientSuite tests not enabled")
	}
	s.srv.SetUp(c)
	s.iam = iam.New(s.srv.auth, aws.USEast)
}

// ClientTests defines integration tests designed to test the client.
// It is not used as a test suite in itself, but embedded within
// another type.
type ClientTests struct {
	iam *iam.IAM
}

func (s *ClientTests) TestCreateAndDeleteUser(c *C) {
	createResp, err := s.iam.CreateUser("gopher", "/gopher/")
	c.Assert(err, IsNil)
	getResp, err := s.iam.GetUser("gopher")
	c.Assert(err, IsNil)
	c.Assert(createResp.User, DeepEquals, getResp.User)
	_, err = s.iam.DeleteUser("gopher")
	c.Assert(err, IsNil)
}

func (s *ClientTests) TestCreateUserError(c *C) {
	_, err := s.iam.CreateUser("gopher", "/gopher/")
	c.Assert(err, IsNil)
	defer s.iam.DeleteUser("gopher")
	_, err = s.iam.CreateUser("gopher", "/")
	iamErr, ok := err.(*iam.Error)
	c.Assert(ok, Equals, true)
	c.Assert(iamErr.StatusCode, Equals, 409)
	c.Assert(iamErr.Code, Equals, "EntityAlreadyExists")
	c.Assert(iamErr.Message, Equals, "User with name gopher already exists.")
}

func (s *ClientTests) TestDeleteUserError(c *C) {
	_, err := s.iam.DeleteUser("gopher")
	iamErr, ok := err.(*iam.Error)
	c.Assert(ok, Equals, true)
	c.Assert(iamErr.StatusCode, Equals, 404)
	c.Assert(iamErr.Code, Equals, "NoSuchEntity")
	c.Assert(iamErr.Message, Equals, "The user with name gopher cannot be found.")
}

func (s *ClientTests) TestGetUserError(c *C) {
	_, err := s.iam.GetUser("gopher")
	iamErr, ok := err.(*iam.Error)
	c.Assert(ok, Equals, true)
	c.Assert(iamErr.StatusCode, Equals, 404)
	c.Assert(iamErr.Code, Equals, "NoSuchEntity")
	c.Assert(iamErr.Message, Equals, "The user with name gopher cannot be found.")
}

func (s *ClientTests) TestPutGetAndDeleteUserPolicy(c *C) {
	userResp, err := s.iam.CreateUser("gopher", "/gopher/")
	c.Assert(err, IsNil)
	defer s.iam.DeleteUser(userResp.User.Name)
	document := `{
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "arn:aws:s3:::8shsns19s90ajahadsj/*",
        "arn:aws:s3:::8shsns19s90ajahadsj"
      ]
    }
  ]
}`
	_, err = s.iam.PutUserPolicy(userResp.User.Name, "EverythingS3", document)
	c.Assert(err, IsNil)
	resp, err := s.iam.GetUserPolicy(userResp.User.Name, "EverythingS3")
	c.Assert(err, IsNil)
	c.Assert(resp.Policy.Name, Equals, "EverythingS3")
	c.Assert(resp.Policy.UserName, Equals, userResp.User.Name)
	gotDocument, err := url.QueryUnescape(resp.Policy.Document)
	c.Assert(err, IsNil)
	c.Assert(gotDocument, Equals, document)
	_, err = s.iam.DeleteUserPolicy(userResp.User.Name, "EverythingS3")
	c.Assert(err, IsNil)
	_, err = s.iam.GetUserPolicy(userResp.User.Name, "EverythingS3")
	c.Assert(err, NotNil)
}
