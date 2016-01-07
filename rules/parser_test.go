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

type ParseSuite struct {
	TempDir string
}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&ParseSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *ParseSuite) TestParsingError(c *C) {
	var err error

	_, err = Parse("../testdata", "", "", "test0")

	c.Assert(err, Not(IsNil))
	c.Assert(err.Error(), Equals, "File ../testdata/test0.mock is not exist")

	_, err = Parse("../testdata", "", "", "error_empty")

	c.Assert(err, Not(IsNil))
	c.Assert(err.Error(), Equals, "File ../testdata/error_empty.mock is empty")

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

func (s *ParseSuite) TestParsing(c *C) {
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

func (s *ParseSuite) TestMultiresponseParsing(c *C) {
	var (
		rule *Rule
		err  error
	)

	rule, err = Parse("../testdata", "", "", "multi_resp")

	c.Assert(rule, Not(IsNil))
	c.Assert(err, IsNil)

	c.Assert(rule.URI(), Equals, ":GET:/test?id=314&simple=&user=bob")

	c.Assert(rule.Responses["1"].Body(), Equals, "{\"test\":1}\n")
	c.Assert(rule.Responses["1"].Code, Equals, 200)
	c.Assert(rule.Responses["1"].Headers["X-Header"], Equals, "1")
	c.Assert(rule.Responses["1"].Delay, Equals, 0.0)

	c.Assert(rule.Responses["2"].Body(), Equals, "{\"test\":2}\n")
	c.Assert(rule.Responses["2"].Code, Equals, 404)
	c.Assert(rule.Responses["2"].Headers["X-Header"], Equals, "2")
	c.Assert(rule.Responses["2"].Delay, Equals, 5.5)
}

func (s *ParseSuite) TestFileResponseParsing(c *C) {
	var (
		rule *Rule
		err  error
	)

	rule, err = Parse("../testdata", "", "", "file_resp")

	c.Assert(rule, Not(IsNil))
	c.Assert(err, IsNil)

	c.Assert(rule.Responses["1"].File, Equals, "../testdata/json/resp1.json")
	c.Assert(rule.Responses["1"].Body(), Equals, "{\"test\":1}\n")
	c.Assert(rule.Responses["2"].File, Equals, "../testdata/json/resp2.json")
	c.Assert(rule.Responses["2"].Body(), Equals, "{\"test\":2}\n")
	c.Assert(rule.Responses["3"].File, Equals, "../testdata/json/resp3.json")
	c.Assert(rule.Responses["3"].Body(), Equals, "")
}

func (s *ParseSuite) TestURLResponseParsing(c *C) {
	var (
		rule *Rule
		err  error
	)

	rule, err = Parse("../testdata", "", "", "url_resp")

	c.Assert(rule, Not(IsNil))
	c.Assert(err, IsNil)

	c.Assert(rule.Responses["1"].URL, Equals, "http://www.domain.com/api/users")
	c.Assert(rule.Responses["2"].URL, Equals, "https://www.domain.com/api/users?limit=20")
}

func (s *ParseSuite) TestWildcardRuleParsing(c *C) {
	var (
		rule *Rule
		err  error
	)

	rule, err = Parse("../testdata", "", "", "wildcard")

	c.Assert(rule, Not(IsNil))
	c.Assert(err, IsNil)

	c.Assert(rule.Wildcard, Equals, "type=1:~username")
	c.Assert(rule.WilcardURI(), Equals, ":GET:type=1:~username")
}
