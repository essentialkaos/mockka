package server

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2015 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"pkg.re/essentialkaos/ek.v1/fsutil"
	"pkg.re/essentialkaos/ek.v1/httputil"
	"pkg.re/essentialkaos/ek.v1/knf"
	"pkg.re/essentialkaos/ek.v1/log"
	"pkg.re/essentialkaos/ek.v1/rand"
	"pkg.re/essentialkaos/ek.v1/system"
	"pkg.re/essentialkaos/ek.v1/timeutil"

	"github.com/essentialkaos/mockka/rules"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	X_MOCKKA_CODE_OK     = 0
	X_MOCKKA_NO_RULE     = 1
	X_MOCKKA_NO_RESPONSE = 2
	X_MOCKKA_CANT_RENDER = 3
)

const ERROR_HTTP_CODE = 599

const (
	MAIN_LOG_DIR         = "main:log-dir"
	HTTP_IP              = "http:ip"
	HTTP_PORT            = "http:port"
	HTTP_READ_TIMEOUT    = "http:read-timeout"
	HTTP_WRITE_TIMEOUT   = "http:write-timeout"
	HTTP_MAX_HEADER_SIZE = "http:max-header-size"
	ACCESS_USER          = "access:user"
	ACCESS_GROUP         = "access:group"
	ACCESS_MOCK_PERMS    = "access:mock-perms"
	ACCESS_LOG_PERMS     = "access:log-perms"
	ACCESS_DIR_PERMS     = "access:dir-perms"
)

// ////////////////////////////////////////////////////////////////////////////////// //

type stabber struct {
	request *http.Request
}

func (s *stabber) RandomNum(start, end int) int {
	if start >= end {
		return start
	}

	d := end - start
	r := rand.Int(d)

	return start + r
}

func (s *stabber) RandomString(length int) string {
	return rand.String(length)
}

func (s *stabber) Query(name string) string {
	if s.request == nil {
		return ""
	}

	query := s.request.URL.Query()

	return strings.Join(query[name], " ")
}

func (s *stabber) IsQuery(name, value string) bool {
	return s.Query(name) == value
}

func (s *stabber) Header(name string) string {
	if s.request == nil {
		return ""
	}

	headers := s.request.Header

	return strings.Join(headers[name], " ")
}

func (s *stabber) IsHeader(name, value string) bool {
	return s.Header(name) == value
}

// ////////////////////////////////////////////////////////////////////////////////// //

var stab *stabber
var serverToken string
var observer *rules.Observer

// ////////////////////////////////////////////////////////////////////////////////// //

func Start(obs *rules.Observer, serverName, customPort string) error {
	if obs == nil {
		return errors.New("Observer is not created")
	}

	observer = obs
	stab = &stabber{}
	serverToken = serverName

	port := knf.GetS(HTTP_PORT)

	if customPort != "" {
		port = customPort
	}

	server := &http.Server{
		Addr:           knf.GetS(HTTP_IP) + ":" + port,
		Handler:        http.NewServeMux(),
		ReadTimeout:    time.Duration(knf.GetI(HTTP_READ_TIMEOUT)) * time.Second,
		WriteTimeout:   time.Duration(knf.GetI(HTTP_WRITE_TIMEOUT)) * time.Second,
		MaxHeaderBytes: knf.GetI(HTTP_MAX_HEADER_SIZE),
	}

	server.Handler.(*http.ServeMux).HandleFunc("/", basicHandler)

	log.Aux("Mockka HTTP server started on %s:%s\n", knf.GetS(HTTP_IP), port)

	return server.ListenAndServe()
}

func basicHandler(w http.ResponseWriter, r *http.Request) {
	var rule *rules.Rule
	var resp *rules.Response

	w.Header().Set("Server", serverToken)

	rule = observer.GetRule(r)

	if rule == nil {
		log.Error("Can't find rule for request %s (%s)", r.URL.String(), r.Method)
		addInfoHeader(w, r, X_MOCKKA_NO_RULE)
		w.WriteHeader(ERROR_HTTP_CODE)
		return
	}

	switch len(rule.Responses) {
	case 0:
		log.Error("Can't find response for request %s (%s)", r.URL.String(), r.Method)
		addInfoHeader(w, r, X_MOCKKA_NO_RESPONSE)
		w.WriteHeader(ERROR_HTTP_CODE)
		return
	case 1:
		resp = rule.Responses[rules.DefaultID]
	default:
		resp = getRandomResponse(rule)
	}

	content, err := renderTemplate(r, resp.Body())

	if err != nil {
		log.Error("Can't render response body: %v", err.Error())
		addInfoHeader(w, r, X_MOCKKA_CANT_RENDER)
		w.WriteHeader(ERROR_HTTP_CODE)
		return
	}

	logRequestInfo(r, rule, resp, content)
	processRequest(w, r, rule, resp, content)
}

func processRequest(w http.ResponseWriter, r *http.Request, rule *rules.Rule, resp *rules.Response, content string) {
	var defResp *rules.Response
	var headers map[string]string
	var ok bool

	var code = 200

	if rule.Auth.User != "" && rule.Auth.Password != "" {
		auth := r.Header.Get("Authorization")
		login, password, ok := parseBasicAuth(auth)

		if !ok || login != rule.Auth.User || password != rule.Auth.Password {
			w.WriteHeader(401)
			return
		}
	}

	if resp.Code == 0 {
		defResp, ok = rule.Responses[rules.DefaultID]

		if ok && defResp.Code != 0 {
			code = defResp.Code
		}
	} else {
		code = resp.Code
	}

	if len(resp.Headers) == 0 {
		defResp, ok = rule.Responses[rules.DefaultID]

		if ok && len(defResp.Headers) != 0 {
			headers = defResp.Headers
		}
	} else {
		headers = resp.Headers
	}

	if resp.Delay > 0 {
		delay := resp.Delay * float64(time.Second)
		time.Sleep(time.Duration(delay))
	}

	for k, v := range headers {
		w.Header().Set(k, v)
	}

	addInfoHeader(w, r, X_MOCKKA_CODE_OK)

	w.WriteHeader(code)
	w.Write([]byte(content))
}

func logRequestInfo(req *http.Request, rule *rules.Rule, resp *rules.Response, content string) {
	filePath := knf.GetS(MAIN_LOG_DIR) + "/" + rule.Service + ".log"
	requredPermChange := !fsutil.IsExist(filePath)

	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		log.Error(err.Error())
		return
	}

	defer fd.Close()

	makeLogRecord(fd, rule.Service, req, rule, resp, content)

	if requredPermChange {
		updatePerms(filePath)
	}
}

func renderTemplate(req *http.Request, content string) (string, error) {
	templ, err := template.New("body").Parse(content)

	if err != nil {
		return "", err
	}

	var bf bytes.Buffer

	stab.request = req
	ct := template.Must(templ, nil)
	ct.Execute(&bf, stab)
	stab.request = nil

	return string(bf.Bytes()), nil
}

func getRandomResponse(rule *rules.Rule) *rules.Response {
	var ids []string

	for id := range rule.Responses {
		if id == rules.DefaultID {
			continue
		}

		ids = append(ids, id)
	}

	return rule.Responses[ids[rand.Int(len(ids)-1)]]
}

func makeLogRecord(fd *os.File, service string, req *http.Request, rule *rules.Rule, resp *rules.Response, content string) {
	date := timeutil.Format(time.Now(), "%Y/%m/%d %T")

	fd.WriteString(fmt.Sprintf("-- %s -----------------------------------------------------------------\n\n", date))
	fd.WriteString(fmt.Sprintf("  %-24s %s\n", "Mock:", rule.Path))

	xForwardedFor := req.Header.Get("X-Forwarded-For")
	xRealIP := req.Header.Get("X-Real-Ip")

	switch {
	case xRealIP != "":
		fd.WriteString(fmt.Sprintf("  %-24s %s\n", "Remote Adress:", xRealIP))
	case xForwardedFor != "":
		fd.WriteString(fmt.Sprintf("  %-24s %s\n", "Remote Adress:", strings.Split(xForwardedFor, ",")[0]))
	default:
		fd.WriteString(fmt.Sprintf("  %-24s %s\n", "Remote Adress:", req.RemoteAddr))
	}

	if rule.Host != "" {
		fd.WriteString(fmt.Sprintf("  %-24s %s\n", "Request Host:", rule.Host))
	}

	fd.WriteString(fmt.Sprintf("  %-24s %s %s\n", "Request:", req.Method, req.RequestURI))
	fd.WriteString(fmt.Sprintf("  %-24s %d %s\n", "Status Code:", resp.Code, httputil.GetDescByCode(resp.Code)))

	fd.WriteString("\n+ HEADERS\n\n")

	reqHeaders := getSortedReqHeaders(req.Header)

	for _, k := range reqHeaders {
		fd.WriteString(fmt.Sprintf("  %-24s %s\n", k+":", strings.Join(req.Header[k], " ")))
	}

	cookies := req.Cookies()

	if len(cookies) != 0 {
		fd.WriteString("\n+ COOKIES\n\n")

		for _, c := range cookies {
			fd.WriteString(fmt.Sprintf("  %s\n", "Request:", c.String()))
		}
	}

	req.ParseForm()

	if req.Method == "GET" {
		query := req.URL.Query()

		if len(query) != 0 {
			fd.WriteString("\n+ QUERY\n\n")

			sortedQuery := getSortedQuery(query)

			for _, k := range sortedQuery {
				fd.WriteString(fmt.Sprintf("  %-24s %s\n", k+":", strings.Join(query[k], " ")))
			}
		}
	} else {
		if len(req.Form) != 0 {
			fd.WriteString("\n+ FORM DATA\n\n")

			for k, v := range req.Form {
				fd.WriteString(fmt.Sprintf("  %-24s %s\n", k+":", strings.Join(v, " ")))
			}
		}
	}

	body, err := ioutil.ReadAll(req.Body)

	if err == nil && len(body) != 0 {
		fd.WriteString("\n+ REQUEST BODY\n\n")
		fd.Write(body)
		fd.WriteString("\n")
	}

	if content != "" {
		fd.WriteString("\n+ RESPONSE BODY\n\n")
		fd.WriteString(content)
	}

	fd.WriteString("\n+ RESPONSE HEADERS\n\n")

	respHeaders := getSortedRespHeaders(resp.Headers)

	for _, k := range respHeaders {
		fd.WriteString(fmt.Sprintf("  %-24s %s\n", k+":", resp.Headers[k]))
	}

	fd.WriteString("\n\n")
}

func updatePerms(logPath string) {
	if knf.HasProp(ACCESS_USER) || knf.HasProp(ACCESS_GROUP) {
		logOwnerUID, logOwnerGID, _ := fsutil.GetOwner(logPath)

		if knf.HasProp(ACCESS_USER) {
			userInfo, err := system.LookupUser(knf.GetS(ACCESS_USER))

			if err != nil {
				logOwnerUID = userInfo.UID
			}
		}

		if knf.HasProp(ACCESS_GROUP) {
			groupInfo, err := system.LookupGroup(knf.GetS(ACCESS_GROUP))

			if err != nil {
				logOwnerGID = groupInfo.GID
			}
		}

		os.Chown(logPath, logOwnerUID, logOwnerGID)
	}

	os.Chmod(logPath, knf.GetM(ACCESS_LOG_PERMS))
}

func parseBasicAuth(auth string) (string, string, bool) {
	if !strings.HasPrefix(auth, "Basic ") {
		return "", "", false
	}

	c, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))

	if err != nil {
		return "", "", false
	}

	cs := string(c)
	s := strings.IndexByte(cs, ':')

	if s < 0 {
		return "", "", false
	}

	return cs[:s], cs[s+1:], true
}

func getSortedReqHeaders(headers http.Header) []string {
	var result []string

	for n := range headers {
		result = append(result, n)
	}

	sort.Strings(result)

	return result
}

func getSortedRespHeaders(headers map[string]string) []string {
	var result []string

	for n := range headers {
		result = append(result, n)
	}

	sort.Strings(result)

	return result
}

func getSortedQuery(query url.Values) []string {
	var result []string

	for n := range query {
		result = append(result, n)
	}

	sort.Strings(result)

	return result
}

func addInfoHeader(w http.ResponseWriter, r *http.Request, code int) {
	if r.Header.Get("Mockka") != "" {
		w.Header().Add("X-Mockka-Code", strconv.Itoa(code))
	}
}
