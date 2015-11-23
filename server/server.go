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
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/essentialkaos/ek/fsutil"
	"github.com/essentialkaos/ek/httputil"
	"github.com/essentialkaos/ek/rand"
	"github.com/essentialkaos/ek/system"
	"github.com/essentialkaos/ek/timeutil"

	"github.com/essentialkaos/mockka/core"
	"github.com/essentialkaos/mockka/rules"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	XMockkaCodeOk     = 0
	XMockkaNoRule     = 1
	XMockkaNoResponse = 2
	XMockkaCantRender = 3
)

const ErrorHTTPCode = 599

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

	port := core.Config.GetS(core.ConfHTTPPort)

	if customPort != "" {
		port = customPort
	}

	server := &http.Server{
		Addr:           core.Config.GetS(core.ConfHTTPIp) + ":" + port,
		Handler:        http.NewServeMux(),
		ReadTimeout:    time.Duration(core.Config.GetI(core.ConfHTTPReadTimeout)) * time.Second,
		WriteTimeout:   time.Duration(core.Config.GetI(core.ConfHTTPWriteTimeout)) * time.Second,
		MaxHeaderBytes: core.Config.GetI(core.ConfHTTPMaxHeaderSize),
	}

	server.Handler.(*http.ServeMux).HandleFunc("/", basicHandler)

	log.Printf("Mockka HTTP server started on %s:%s\n", core.Config.GetS(core.ConfHTTPIp), port)

	return server.ListenAndServe()
}

func basicHandler(w http.ResponseWriter, r *http.Request) {
	var rule *rules.Rule
	var resp *rules.Response

	w.Header().Set("Server", serverToken)

	rule = observer.GetRule(r)

	if rule == nil {
		log.Printf("[ERROR] Can't find rule for request %s (%s)\n", r.URL.String(), r.Method)
		addInfoHeader(w, r, XMockkaNoRule)
		w.WriteHeader(ErrorHTTPCode)
		return
	}

	switch len(rule.Responses) {
	case 0:
		log.Printf("[ERROR] Can't find response for request %s (%s)\n", r.URL.String(), r.Method)
		addInfoHeader(w, r, XMockkaNoResponse)
		w.WriteHeader(ErrorHTTPCode)
		return
	case 1:
		resp = rule.Responses[rules.DefaultID]
	default:
		resp = getRandomResponse(rule)
	}

	content, err := renderTemplate(r, resp.Body())

	if err != nil {
		log.Printf("[ERROR] Can't render response body - %s\n", err.Error())
		addInfoHeader(w, r, XMockkaCantRender)
		w.WriteHeader(ErrorHTTPCode)
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

	addInfoHeader(w, r, XMockkaCodeOk)

	w.WriteHeader(code)
	w.Write([]byte(content))
}

func logRequestInfo(req *http.Request, rule *rules.Rule, resp *rules.Response, content string) {
	filePath := core.Config.GetS(core.ConfMainLogDir) + "/" + rule.Service + ".log"
	requredPermChange := !fsutil.IsExist(filePath)

	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		log.Printf("[ERROR] %s\n", err.Error())
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
	if core.Config.HasProp(core.ConfAccessUser) || core.Config.HasProp(core.ConfAccessGroup) {
		logOwnerUID, logOwnerGID, _ := fsutil.GetOwner(logPath)

		if core.Config.HasProp(core.ConfAccessUser) {
			userInfo, err := system.LookupUser(core.Config.GetS(core.ConfAccessUser))

			if err != nil {
				logOwnerUID = userInfo.UID
			}
		}

		if core.Config.HasProp(core.ConfAccessGroup) {
			groupInfo, err := system.LookupGroup(core.Config.GetS(core.ConfAccessGroup))

			if err != nil {
				logOwnerGID = groupInfo.GID
			}
		}

		os.Chown(logPath, logOwnerUID, logOwnerGID)
	}

	os.Chmod(logPath, core.Config.GetM(core.ConfAccessLogPerms))
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
