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
	"strconv"
	"strings"

	"pkg.re/essentialkaos/ek.v3/fsutil"
	"pkg.re/essentialkaos/ek.v3/httputil"

	"github.com/essentialkaos/mockka/urlutil"
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

// ParsePath parse path of rule file and return service name,
// mock name (without extension) and inner dir
func ParsePath(path string) (string, string, string) {
	pathSlice := strings.Split(path, "/")
	pathItems := len(pathSlice)

	switch pathItems {
	case 1:
		return path, "", ""
	case 2:
		return pathSlice[0], pathSlice[1], ""
	default:
		return pathSlice[0], pathSlice[pathItems-1], strings.Join(pathSlice[1:pathItems-1], "/")
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

func parseRuleData(data []string, ruleDir, service, dir, mock string) (*Rule, error) {
	var rule = NewRule()

	rule.Dir = dir
	rule.Name = mock
	rule.Service = service
	rule.FullName = path.Join(dir, mock)
	rule.PrettyPath = path.Join(rule.Service, rule.FullName)
	rule.Path = path.Join(ruleDir, service, dir, mock+".mock")
	rule.Request = &Request{}

	var section, id, source string
	var overwrite bool

	for _, line := range data {
		if line[0:1] == "@" {
			section, id, source, overwrite = parseSectionHeader(line)

			if section == "RESPONSE" && source != "" {
				resp := getResponse(rule, id)

				if httputil.IsURL(source) {
					resp.Overwrite = overwrite
					resp.URL = source
				} else {
					resp.Headers["Content-Type"] = guessContentType(source)

					if service != "" {
						resp.File = path.Join(ruleDir, service, source)
					} else {
						resp.File = path.Join(ruleDir, source)
					}
				}
			}

			continue
		}

		switch section {
		case "DESCRIPTION":
			rule.Desc += line

		case "HOST":
			rule.Request.Host = strings.TrimRight(line, " ")

		case "REQUEST":
			reqMethod, reqURL := parseRequestInfo(line)

			if reqMethod == "" || reqURL == "" {
				return nil, fmt.Errorf("Can't parse file %s - section REQUEST is malformed", rule.Path)
			}

			if reqURL[0:1] != "/" {
				return nil, fmt.Errorf("Can't parse file %s - request url must start from /", rule.Path)
			}

			if strings.Contains(reqURL, "*") {
				_, err := url.Parse(reqURL)

				if err == nil {
					rule.IsWildcard = strings.Contains(reqURL, "*")
				} else {
					return nil, fmt.Errorf("Can't parse file %s - can't parse query in REQUEST section", rule.Path)
				}
			}

			rule.Request.Method, rule.Request.URL = reqMethod, reqURL

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

	rule.Request.NURL = urlutil.SortParams(rule.Request.URL)
	rule.Request.URI = rule.Request.Host + ":" + rule.Request.Method + ":" + rule.Request.NURL

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

func parseSectionHeader(header string) (string, string, string, bool) {
	var slice []string

	var (
		section   = ""
		id        = DEFAULT
		source    = ""
		overwrite = false
	)

	section = strings.Replace(header[1:], " ", "", -1)

	if strings.Contains(section, "<<") {
		slice = strings.Split(section, "<<")
		source, section, overwrite = slice[1], slice[0], true
	} else if strings.Contains(section, "<") {
		slice = strings.Split(section, "<")
		source, section = slice[1], slice[0]
	}

	if strings.Contains(section, ":") {
		slice = strings.Split(section, ":")
		id, section = slice[1], slice[0]
	}

	return strings.ToUpper(section), id, source, overwrite
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
