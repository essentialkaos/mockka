package rules

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"io/ioutil"
	"time"

	"pkg.re/essentialkaos/ek.v3/timeutil"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const DEFAULT = "_default"

// ////////////////////////////////////////////////////////////////////////////////// //

type Rule struct {
	Name       string               // Mock file name
	FullName   string               // Dir + mock file name
	Service    string               // Mock file dir name
	Dir        string               // Inner dir
	Path       string               // Full path to file
	PrettyPath string               // Service + FullName
	Desc       string               // Mock description
	Auth       *Auth                // Basic auth login and pass
	Request    *Request             // Request method and url
	Responses  map[string]*Response // Responses map
	ModTime    time.Time            // Mock file mod time
	IsWildcard bool                 // Wildcard marker
}

type Auth struct {
	User     string // Username
	Password string // Password
}

type Request struct {
	Host   string // Request host
	Method string // Request method
	URL    string // Request URL
	NURL   string // Normalized (sorted) URL
	URI    string // URI (host + method + normalized url)
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

// String return string with request info
func (r *Request) String() string {
	if r == nil {
		return "Nil"
	}

	host := "-"

	if r.Host != "" {
		host = r.Host
	}

	return fmt.Sprintf(
		"Host: %s | Method: %s | URL: %s | NURL: %s | URI: %s",
		host, r.Method, r.URL, r.NURL, r.URI,
	)
}

// String return string with rule info
func (r *Rule) String() string {
	if r == nil {
		return "Nil"
	}

	auth := "-"
	dir := "-"

	if r.Auth != nil {
		auth = r.Auth.User + "/" + r.Auth.Password
	}

	if r.Dir != "" {
		dir = r.Dir
	}

	return fmt.Sprintf(
		"FullName: %s | Service: %s | Dir: %s | Path: %s | Desc: %t | Auth: %s | ResponsesN: %d | ModTime: %s | IsWildcard: %t",
		r.FullName, r.Service, dir, r.Path, r.Desc != "", auth, len(r.Responses),
		timeutil.Format(r.ModTime, "%Y/%m/%d %H:%M:%S"), r.IsWildcard,
	)
}

// String return string with response info
func (r *Response) String() string {
	if r == nil {
		return "Nil"
	}

	file := "-"
	url := "-"

	if r.File != "" {
		file = r.File
	}

	if r.URL != "" {
		url = r.URL
	}

	return fmt.Sprintf(
		"ContentSyms: %d | File: %s | URL: %s | Code: %d | HeadersNum: %d | Delay: %g | OverwriteFlag: %t",
		len(r.Content), file, url, r.Code, len(r.Headers), r.Delay, r.Overwrite,
	)
}

// Body return reponse body
func (r *Response) Body() string {
	if r == nil {
		return ""
	}

	if r.File != "" {
		body, err := ioutil.ReadFile(r.File)

		if err != nil {
			return ""
		}

		return string(body)
	}

	return r.Content
}

// ////////////////////////////////////////////////////////////////////////////////// //
