package rules

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
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
	"pkg.re/essentialkaos/ek.v1/httputil"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// contentTypes contains map file ext -> content type
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

	err := checkMockFile(mockFile)

	if err != nil {
		return nil, err
	}

	// We don't check errors, because file was checked before
	fd, _ := os.Open(mockFile)

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

	var section, id, source string

	for _, line := range data {
		if line[0:1] == "@" {
			section, id, source = parseSectionHeader(line)

			if section == "RESPONSE" && source != "" {
				resp := getResponse(rule, id)

				if httputil.IsURL(source) {
					resp.URL = source
				} else {
					resp.Headers["Content-Type"] = guessContentType(source)

					if service != "" {
						resp.File = ruleDir + "/" + service + "/" + source
					} else {
						resp.File = ruleDir + "/" + source
					}
				}
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
				return nil, fmt.Errorf("Can't parse file %s - section REQUEST is malformed", rule.Path)
			}

			if reqURL[0:1] != "/" {
				return nil, fmt.Errorf("Can't parse file %s - request url must start from /", rule.Path)
			}

			if strings.Contains(reqURL, "*") {
				urlStruct, err := url.Parse(reqURL)

				if err == nil {
					queryWildcard := getQueryWildcard(urlStruct.Query())

					if queryWildcard == "" {
						return nil, fmt.Errorf("Can't parse file %s - wildcard in REQUEST section is malformed", rule.Path)
					}

					rule.Wildcard = queryWildcard
				} else {
					return nil, fmt.Errorf("Can't parse file %s - can't parse query in REQUEST section", rule.Path)
				}
			}

			rule.Request = &Request{reqMethod, reqURL}

		case "RESPONSE":
			getResponse(rule, id).Content += line + "\n"

		case "CODE":
			code, err := strconv.Atoi(strings.TrimRight(line, " "))

			if err != nil {
				return nil, fmt.Errorf("Can't parse file %s - section CODE is malformed", rule.Path)
			}

			getResponse(rule, id).Code = code

		case "HEADERS":
			headerName, headerValue := parseHTTPHeader(line)

			if headerName == "" || headerValue == "" {
				return nil, fmt.Errorf("Can't parse file %s - section HEADERS is malformed", rule.Path)
			}

			getResponse(rule, id).Headers[headerName] = headerValue

		case "DELAY":
			delay, err := strconv.ParseFloat(strings.TrimRight(line, " "), 64)

			if err != nil {
				return nil, fmt.Errorf("Can't parse file %s - section DELAY is malformed", rule.Path)
			}

			getResponse(rule, id).Delay = delay

		case "AUTH":
			lpa := strings.Split(strings.TrimRight(line, " "), ":")

			if len(lpa) != 2 {
				return nil, fmt.Errorf("Can't parse file %s - section AUTH is malformed", rule.Path)
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

func checkMockFile(file string) error {
	switch {
	case fsutil.IsExist(file) == false:
		return fmt.Errorf("File %s is not exist", file)
	case fsutil.IsReadable(file) == false:
		return fmt.Errorf("File %s is not readable", file)
	case fsutil.IsNonEmpty(file) == false:
		return fmt.Errorf("File %s is empty", file)
	}

	return nil
}

func parseSectionHeader(header string) (string, string, string) {
	var slice []string

	var (
		section = ""
		id      = DEFAULT
		source  = ""
	)
	section = strings.Replace(header[1:], " ", "", -1)

	if strings.Contains(section, "<") {
		slice = strings.Split(section, "<")
		source, section = slice[1], slice[0]
	}

	if strings.Contains(section, ":") {
		slice = strings.Split(section, ":")
		id, section = slice[1], slice[0]
	}

	return strings.ToUpper(section), id, source
}

func parseRequestInfo(request string) (string, string) {
	requestSlice := strings.Split(request, " ")

	if len(requestSlice) < 2 {
		return "", ""
	}

	return strings.ToUpper(requestSlice[0]), requestSlice[1]
}

func parseHTTPHeader(header string) (string, string) {
	headerSlice := strings.Split(header, ":")

	if len(headerSlice) < 2 {
		return "", ""
	}

	return strings.TrimRight(headerSlice[0], " "), strings.TrimLeft(headerSlice[1], " ")
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
	fileExt := path.Ext(file)
	contentType, ok := contentTypes[fileExt]

	if ok {
		return contentType
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
