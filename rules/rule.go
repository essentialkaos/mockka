package rules

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"io/ioutil"
	"net/url"
	"sort"
	"strings"
	"time"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const DEFAULT = "_default"

// ////////////////////////////////////////////////////////////////////////////////// //

type Rule struct {
	Name      string               // mock file name
	FullName  string               // dir + mock file name
	Service   string               // mock file dir name
	Dir       string               // inner dir
	Path      string               // full path to file
	Desc      string               // mock description
	Host      string               // host
	Auth      *Auth                // basic auth login and pass
	Request   *Request             // request method and url
	Responses map[string]*Response // responses map
	ModTime   time.Time            // mock file mod time
	Wildcard  string               // Wildcard string
}

type Auth struct {
	User     string
	Password string
}

type Request struct {
	Method string
	URL    string
}

type Response struct {
	Content string
	File    string
	Code    int
	Headers map[string]string
	Delay   float64
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Create new rule struct
func NewRule() *Rule {
	return &Rule{
		Auth:      &Auth{},
		Request:   &Request{},
		Responses: make(map[string]*Response),
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Get reponse body
func (r *Response) Body() string {
	if r.Content == "" {
		body, err := ioutil.ReadFile(r.File)

		if err != nil {
			return ""
		}

		return string(body)
	}

	return r.Content
}

// Get rule URI
func (r *Rule) URI() string {
	return r.Host + ":" + r.Request.Method + ":" + getSortedURI(r.Request.URL)
}

// Get rule wildcard uri
func (r *Rule) WilcardURI() string {
	return r.Host + ":" + r.Request.Method + ":" + r.Wildcard
}

// ////////////////////////////////////////////////////////////////////////////////// //

func getSortedURI(uri string) string {
	if !strings.Contains(uri, "?") {
		return uri
	}

	u, err := url.Parse(uri)

	if err != nil {
		return uri
	}

	query := u.Query()
	result := u.Path + "?"

	var sortedQuery []string

	for qp := range query {
		sortedQuery = append(sortedQuery, qp)
	}

	sort.Strings(sortedQuery)

	for _, qp := range sortedQuery {
		result += qp + "=" + strings.Join(query[qp], "") + "&"
	}

	return result[0 : len(result)-1]
}
