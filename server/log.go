package server

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"
	"strings"
	"time"

	"pkg.re/essentialkaos/ek.v3/kv"
	"pkg.re/essentialkaos/ek.v3/timeutil"
)

// ////////////////////////////////////////////////////////////////////////////////// //

type LogRecord struct {
	Date            time.Time `json:"date"`
	Mock            string    `json:"mock"`
	RemoteAdress    string    `json:"remote_adress"`
	RequestHost     string    `json:"request_host"`
	Method          string    `json:"method"`
	Request         string    `json:"request"`
	Query           []*kv.KV  `json:"query"`
	RequestHeaders  []*kv.KV  `json:"request_headers"`
	ResponseHeaders []*kv.KV  `json:"response_headers"`
	RequestBody     string    `json:"request_body"`
	ResponseBody    string    `json:"response_body"`
	ResponseURL     string    `json:"response_url"`
	Cookies         []string  `json:"cookies"`
	StatusCode      int       `json:"status_code"`
	StatusDesc      string    `json:"status_desc"`
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Write write log record to file
func (lr *LogRecord) Write(file string) error {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer fd.Close()

	date := timeutil.Format(lr.Date, "%Y/%m/%d %T")

	fmt.Fprintf(fd, "-- %s -----------------------------------------------------------------\n\n", date)
	fmt.Fprintf(fd, "  %-24s %s\n", "Mock:", lr.Mock)

	if lr.RemoteAdress != "" {
		fmt.Fprintf(fd, "  %-24s %s\n", "Remote Adress:", lr.RemoteAdress)
	}

	if lr.RequestHost != "" {
		fmt.Fprintf(fd, "  %-24s %s\n", "Request Host:", lr.RequestHost)
	}

	fmt.Fprintf(fd, "  %-24s %s %s\n", "Request:", lr.Method, lr.Request)

	if lr.ResponseURL != "" {
		fmt.Fprintf(fd, "  %-24s %s\n", "Response URL:", lr.ResponseURL)
	}

	fmt.Fprintf(fd, "  %-24s %d %s\n", "Status Code:", lr.StatusCode, lr.StatusDesc)

	if len(lr.RequestHeaders) != 0 {
		fmt.Fprintf(fd, "\n+ HEADERS\n\n")

		for _, k := range lr.RequestHeaders {
			fmt.Fprintf(fd, "  %-24s %s\n", k.Key+":", k.String())
		}
	}

	if len(lr.Cookies) != 0 {
		fmt.Fprintf(fd, "\n+ COOKIES\n\n")

		for _, c := range lr.Cookies {
			fmt.Fprintf(fd, "  %s\n", c)
		}
	}

	if lr.Method == "GET" {
		if len(lr.Query) != 0 {
			fmt.Fprintf(fd, "\n+ QUERY\n\n")

			for _, q := range lr.Query {
				fmt.Fprintf(fd, "  %-24s %s\n", q.Key+":", q.String())
			}
		}
	}

	if lr.RequestBody != "" {
		fmt.Fprintf(fd, "\n+ REQUEST BODY\n\n")
		fmt.Fprintf(fd, lr.RequestBody)

		if !strings.HasSuffix(lr.RequestBody, "\n") {
			fmt.Fprintln(fd, "")
		}
	}

	if lr.ResponseBody != "" {
		fmt.Fprintf(fd, "\n+ RESPONSE BODY\n\n")
		fmt.Fprintf(fd, lr.ResponseBody)

		if !strings.HasSuffix(lr.ResponseBody, "\n") {
			fmt.Fprintln(fd, "")
		}
	}

	if len(lr.ResponseHeaders) != 0 {
		fmt.Fprintf(fd, "\n+ RESPONSE HEADERS\n\n")

		for _, h := range lr.ResponseHeaders {
			fmt.Fprintf(fd, "  %-24s %s\n", h.Key+":", h.String())
		}
	}

	fmt.Fprintf(fd, "\n\n")

	return nil
}
