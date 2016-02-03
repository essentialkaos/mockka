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
	Name      string               // Mock file name
	FullName  string               // Dir + mock file name
	Service   string               // Mock file dir name
	Dir       string               // Inner dir
	Path      string               // Full path to file
	Desc      string               // Mock description
	Host      string               // Host
	Auth      *Auth                // Basic auth login and pass
	Request   *Request             // Request method and url
	Responses map[string]*Response // Responses map
	ModTime   time.Time            // Mock file mod time
	Wildcard  string               // Wildcard string
}

type Auth struct {
	User     string // Username
	Password string // Password
}

type Request struct {
	Method string // Request method
	URL    string // Request URL
}

type Response struct {
	Content   string            // Static content
	File      string            // Path to file with content
	URL       string            // URL for request proxying
	Code      int               // Status code
	Headers   map[string]string // Map with headers
	Delay     float64           // Response delay
	Overwrite bool              // Proxying overwrite mode flag
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

// Body return reponse body
func (r *Response) Body() string {
	if r.File != "" {
		body, err := ioutil.ReadFile(r.File)

		if err != nil {
			return ""
		}

		return string(body)
	}

	return r.Content
}

// URI return rule URI
func (r *Rule) URI() string {
	return r.Host + ":" + r.Request.Method + ":" + getSortedURI(r.Request.URL)
}

// WilcardURI return rule wildcard uri
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
