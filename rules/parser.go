package rules

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2015 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"pkg.re/essentialkaos/ek.v1/fsutil"
)

// ////////////////////////////////////////////////////////////////////////////////// //

var contentTypes = map[string]string{
	".json": "text/javascript",
	".txt":  "text/plain",
	".xml":  "text/xml",
	".csv":  "text/csv",
	".html": "text/html",
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Parse rule file
func Parse(ruleDir, service, dir, mock string) (*Rule, error) {
	mockFile := path.Join(ruleDir, service, dir, mock+".mock")

	if !fsutil.CheckPerms("FR", mockFile) {
		return NewRule(), fmt.Errorf("File %s is not readable or not exist", mockFile)
	}

	fd, err := os.Open(mockFile)

	if err != nil {
		return nil, err
	}

	defer fd.Close()

	reader := bufio.NewReader(fd)
	scanner := bufio.NewScanner(reader)

	var data []string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || strings.Replace(line, " ", "", -1) == "" {
			continue
		}

		if strings.HasPrefix(strings.Trim(line, " "), "#") {
			continue
		}

		data = append(data, line)
	}

	return parseRuleData(data, ruleDir, service, dir, mock)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func parseRuleData(data []string, ruleDir, service, dir, mock string) (*Rule, error) {
	var rule = NewRule()

	rule.Name = mock
	rule.FullName = path.Join(dir, mock)
	rule.Dir = dir
	rule.Service = service
	rule.Path = path.Join(ruleDir, service, dir, mock+".mock")

	var section, id, file string

	if len(data) == 0 {
		return rule, fmt.Errorf("Can't parse file %s - file is empty", rule.Path)
	}

	for _, line := range data {
		if line[0:1] == "@" {
			section, id, file = parseSectionHeader(line)

			if section == "RESPONSE" && file != "" {
				resp := getResponse(rule, id)
				resp.Headers["Content-Type"] = guessContentType(file)
				resp.File = ruleDir + "/" + service + "/" + file
			}

			continue
		}

		switch section {
		case "DESCRIPTION":
			rule.Desc += line

		case "HOST":
			rule.Host = strings.TrimRight(line, " ")

		case "REQUEST":
			reqMethod, reqURL := parseRequestInfo(line)

			if reqMethod == "" || reqURL == "" {
				return NewRule(), fmt.Errorf("Can't parse file %s - section REQUEST is malformed", rule.Path)
			}

			if reqURL[0:1] != "/" {
				return NewRule(), fmt.Errorf("Can't parse file %s - request url must start from /", rule.Path)
			}

			if strings.Contains(reqURL, "*") {
				urlStruct, err := url.Parse(reqURL)

				if err == nil {
					rule.Wildcard = getQueryWildcard(urlStruct.Query())
				}
			}

			rule.Request = &Request{reqMethod, reqURL}

		case "RESPONSE":
			getResponse(rule, id).Content += line + "\n"

		case "CODE":
			code, err := strconv.Atoi(strings.TrimRight(line, " "))

			if err != nil {
				return NewRule(), fmt.Errorf("Can't parse file %s - section CODE is malformed", rule.Path)
			}

			getResponse(rule, id).Code = code

		case "HEADERS":
			headerName, headerValue := parseHTTPHeader(line)

			if headerName == "" || headerValue == "" {
				return NewRule(), fmt.Errorf("Can't parse file %s - section HEADERS is malformed", rule.Path)
			}

			getResponse(rule, id).Headers[headerName] = headerValue

		case "DELAY":
			delay, err := strconv.ParseFloat(strings.TrimRight(line, " "), 64)

			if err != nil {
				return NewRule(), fmt.Errorf("Can't parse file %s - section DELAY is malformed", rule.Path)
			}

			getResponse(rule, id).Delay = delay

		case "AUTH":
			lpa := strings.Split(strings.TrimRight(line, " "), ":")

			if len(lpa) != 2 {
				return NewRule(), fmt.Errorf("Can't parse file %s - section AUTH is malformed", rule.Path)
			}

			rule.Auth.User, rule.Auth.Password = lpa[0], lpa[1]
		}
	}

	// If all sections in rule file is empty we create default response
	if len(rule.Responses) == 0 {
		rule.Responses[DEFAULT] = &Response{Headers: make(map[string]string)}
	}

	mtime, _ := fsutil.GetMTime(rule.Path)
	rule.ModTime = mtime

	return rule, nil
}

func parseSectionHeader(header string) (string, string, string) {
	var s []string
	var section, id, file string = "", DEFAULT, ""

	section = strings.Replace(header[1:], " ", "", -1)

	if strings.Contains(section, "<") {
		s = strings.Split(section, "<")
		file, section = s[1], s[0]
	}

	if strings.Contains(section, ":") {
		s = strings.Split(section, ":")
		id, section = s[1], s[0]
	}

	return strings.ToUpper(section), id, file
}

func parseRequestInfo(request string) (string, string) {
	s := strings.Split(request, " ")

	if len(s) < 2 {
		return "", ""
	}

	return strings.ToUpper(s[0]), s[1]
}

func parseHTTPHeader(header string) (string, string) {
	s := strings.Split(header, ":")

	if len(s) < 2 {
		return "", ""
	}

	return strings.TrimRight(s[0], " "), strings.TrimLeft(s[1], " ")
}

func getResponse(rule *Rule, id string) *Response {
	resp, ok := rule.Responses[id]

	if ok {
		return resp
	}

	resp = &Response{Headers: make(map[string]string)}
	rule.Responses[id] = resp

	return resp
}

func guessContentType(file string) string {
	ext := path.Ext(file)
	ct, ok := contentTypes[ext]

	if ok {
		return ct
	}

	return "text/plain"
}

func getQueryWildcard(values url.Values) string {
	var queryItems []string

	for itemName, item := range values {
		itemValue := strings.Join(item, "")

		switch itemValue {
		case "*":
			queryItems = append(queryItems, "~"+itemName)
		default:
			queryItems = append(queryItems, itemName+"="+itemValue)
		}
	}

	if len(queryItems) == 0 {
		return ""
	}

	sort.Strings(queryItems)

	return strings.Join(queryItems, ":")
}
