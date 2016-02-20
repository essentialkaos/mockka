package urlutil

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

type URLUtilSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&URLUtilSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *URLUtilSuite) TestMatch(c *C) {
	c.Assert(Match("/test", "/test"), Equals, true)
	c.Assert(Match("/*", "/test"), Equals, true)
	c.Assert(Match("/te*t", "/test"), Equals, true)
	c.Assert(Match("/test/*", "/test/something"), Equals, true)
	c.Assert(Match("/user?action=*", "/user?action=edit"), Equals, true)
	c.Assert(Match("/user?action=*&rnd=*", "/user?action=edit&rnd=123"), Equals, true)

	c.Assert(Match("", ""), Equals, false)
	c.Assert(Match("/test", "/testA"), Equals, false)
	c.Assert(Match("/testA", "/test"), Equals, false)
	c.Assert(Match("/testa", "/testb"), Equals, false)
	c.Assert(Match("/user?action=*", "/user?action=edit&rnd"), Equals, false)
	c.Assert(Match("/user?action=*", "/user?action=edit&rnd=123"), Equals, false)
}

func (s *URLUtilSuite) TestEquals(c *C) {
	c.Assert(EqualPatterns("/user*", "/users*"), Equals, true)
	c.Assert(EqualPatterns("/users/andy*", "/users/*"), Equals, true)
	c.Assert(EqualPatterns("/user?id=*", "/user?id*"), Equals, true)

	c.Assert(EqualPatterns("/users/*", "/user/*"), Equals, false)
	c.Assert(EqualPatterns("/user?id=*", "/user?id=*&rnd=*"), Equals, false)
}

func (s *URLUtilSuite) TestSort(c *C) {
	c.Assert(SortParams("/test"), Equals, "/test")
	c.Assert(SortParams("/test?a=1"), Equals, "/test?a=1")
	c.Assert(SortParams("/test/path/index.php?a=1&z=2&f=3&l=4"), Equals, "/test/path/index.php?a=1&f=3&l=4&z=2")
	c.Assert(SortParams("/test/path/index.php?a=1&z=2&f&k"), Equals, "/test/path/index.php?a=1&f&k&z=2")
	c.Assert(SortParams("/test?a=1&z=2#some_fragment"), Equals, "/test?a=1&z=2#some_fragment")
}
