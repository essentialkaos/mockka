package rules

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"testing"

	. "pkg.re/check.v1"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type ParserTest struct {
	TempDir string
}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&ParserTest{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *ParserTest) TestParsingError(c *C) {
	var err error

	_, err = Parse("../testdata", "", "", "test0")

	c.Assert(err, Not(IsNil))
	c.Assert(err.Error(), Equals, "File ../testdata/test0.mock is not readable or not exist")

	_, err = Parse("../testdata", "", "", "error_empty")

	c.Assert(err, Not(IsNil))
	c.Assert(err.Error(), Equals, "Can't parse file ../testdata/error_empty.mock - file is empty")

	_, err = Parse("../testdata", "", "", "error_resp1")

	c.Assert(err, Not(IsNil))
	c.Assert(err.Error(), Equals, "Can't parse file ../testdata/error_resp1.mock - section REQUEST is malformed")

	_, err = Parse("../testdata", "", "", "error_resp2")

	c.Assert(err, Not(IsNil))
	c.Assert(err.Error(), Equals, "Can't parse file ../testdata/error_resp2.mock - request url must start from /")

	_, err = Parse("../testdata", "", "", "error_code")

	c.Assert(err, Not(IsNil))
	c.Assert(err.Error(), Equals, "Can't parse file ../testdata/error_code.mock - section CODE is malformed")

	_, err = Parse("../testdata", "", "", "error_headers")

	c.Assert(err, Not(IsNil))
	c.Assert(err.Error(), Equals, "Can't parse file ../testdata/error_headers.mock - section HEADERS is malformed")

	_, err = Parse("../testdata", "", "", "error_delay")

	c.Assert(err, Not(IsNil))
	c.Assert(err.Error(), Equals, "Can't parse file ../testdata/error_delay.mock - section DELAY is malformed")

	_, err = Parse("../testdata", "", "", "error_auth")

	c.Assert(err, Not(IsNil))
	c.Assert(err.Error(), Equals, "Can't parse file ../testdata/error_auth.mock - section AUTH is malformed")
}

func (s *ParserTest) TestParsing(c *C) {
	var (
		rule *Rule
		err  error
	)

	rule, err = Parse("../testdata", "test1", "dir1", "test")

	c.Assert(rule, Not(IsNil))
	c.Assert(err, IsNil)

	c.Assert(rule.Name, Equals, "test")
	c.Assert(rule.FullName, Equals, "dir1/test")
	c.Assert(rule.Dir, Equals, "dir1")
	c.Assert(rule.Service, Equals, "test1")
	c.Assert(rule.Path, Equals, "../testdata/test1/dir1/test.mock")

	c.Assert(rule.Desc, Equals, "Test mock file")
	c.Assert(rule.Host, Equals, "test.domain")
	c.Assert(rule.Request.Method, Equals, "GET")
	c.Assert(rule.Request.URL, Equals, "/test?rnd=123")
	c.Assert(rule.Auth.User, Equals, "user1")
	c.Assert(rule.Auth.Password, Equals, "password1")
	c.Assert(rule.Responses[DEFAULT].Body(), Equals, "{\"test\":123}\n")
	c.Assert(rule.Responses[DEFAULT].Code, Equals, 200)
	c.Assert(rule.Responses[DEFAULT].Headers["Content-Type"], Equals, "application/json")
	c.Assert(rule.Responses[DEFAULT].Delay, Equals, 12.3)
}

func (s *ParserTest) TestMultiresponseParsing(c *C) {
	var (
		rule *Rule
		err  error
	)

	rule, err = Parse("../testdata", "", "", "multi_resp")

	c.Assert(rule, Not(IsNil))
	c.Assert(err, IsNil)

	c.Assert(rule.Responses["1"].Body(), Equals, "{\"test\":1}\n")
	c.Assert(rule.Responses["1"].Code, Equals, 200)
	c.Assert(rule.Responses["1"].Headers["X-Header"], Equals, "1")
	c.Assert(rule.Responses["1"].Delay, Equals, 0.0)

	c.Assert(rule.Responses["2"].Body(), Equals, "{\"test\":2}\n")
	c.Assert(rule.Responses["2"].Code, Equals, 404)
	c.Assert(rule.Responses["2"].Headers["X-Header"], Equals, "2")
	c.Assert(rule.Responses["2"].Delay, Equals, 5.5)
}

func (s *ParserTest) TestFileResponseParsing(c *C) {
	var (
		rule *Rule
		err  error
	)

	rule, err = Parse("../testdata", "", "", "file_resp")

	c.Assert(rule, Not(IsNil))
	c.Assert(err, IsNil)

	c.Assert(rule.Responses["1"].Body(), Equals, "{\"test\":1}\n")
	c.Assert(rule.Responses["2"].Body(), Equals, "{\"test\":2}\n")
}
