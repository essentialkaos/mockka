package rules

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"io/ioutil"
	"time"
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

// ////////////////////////////////////////////////////////////////////////////////// //
