package server

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2015 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"pkg.re/essentialkaos/ek.v1/fsutil"
	"pkg.re/essentialkaos/ek.v1/httputil"
	"pkg.re/essentialkaos/ek.v1/knf"
	"pkg.re/essentialkaos/ek.v1/kv"
	"pkg.re/essentialkaos/ek.v1/log"
	"pkg.re/essentialkaos/ek.v1/rand"
	"pkg.re/essentialkaos/ek.v1/req"
	"pkg.re/essentialkaos/ek.v1/system"

	"github.com/essentialkaos/mockka/rules"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	X_MOCKKA_CODE_OK     = 0
	X_MOCKKA_NO_RULE     = 1
	X_MOCKKA_NO_RESPONSE = 2
	X_MOCKKA_CANT_RENDER = 3
	X_MOCKKA_CANT_PROXY  = 4
	X_MOCKKA_FORBIDDEN   = 5
)

const ERROR_HTTP_CODE = 599

const (
	MAIN_LOG_DIR              = "main:log-dir"
	HTTP_IP                   = "http:ip"
	HTTP_PORT                 = "http:port"
	HTTP_READ_TIMEOUT         = "http:read-timeout"
	HTTP_WRITE_TIMEOUT        = "http:write-timeout"
	HTTP_MAX_HEADER_SIZE      = "http:max-header-size"
	PROCESSING_ALLOW_PROXYING = "processing:allow-proxying"
	ACCESS_USER               = "access:user"
	ACCESS_GROUP              = "access:group"
	ACCESS_MOCK_PERMS         = "access:mock-perms"
	ACCESS_LOG_PERMS          = "access:log-perms"
	ACCESS_DIR_PERMS          = "access:dir-perms"
)

// ////////////////////////////////////////////////////////////////////////////////// //

var (
	serverToken string
	observer    *rules.Observer
	stabber     *Stabber
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Start starts mockka HTTP server
func Start(obs *rules.Observer, serverName, customPort string) error {
	if obs == nil {
		return errors.New("Observer is not created")
	}

	observer = obs
	serverToken = serverName
	stabber = &Stabber{}

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

// basicHandler is handler for all requests
func basicHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		rule *rules.Rule
		resp *rules.Response
	)

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
		resp = rule.Responses[rules.DEFAULT]
	default:
		resp = getRandomResponse(rule)
	}

	var content = ""

	if r.Method != "HEAD" {
		if resp.URL == "" {
			content, err = renderTemplate(r, resp.Body())

			if err != nil {
				log.Error("Can't render response body: %v", err)
				addInfoHeader(w, r, X_MOCKKA_CANT_RENDER)
				w.WriteHeader(ERROR_HTTP_CODE)
				return
			}
		} else {
			if !knf.GetB(PROCESSING_ALLOW_PROXYING) {
				log.Error("Can't proxy request: proxying disabled in configuration file")
				addInfoHeader(w, r, X_MOCKKA_FORBIDDEN)
				w.WriteHeader(ERROR_HTTP_CODE)
				return
			}

			content, err = proxyRequest(r, rule, resp)

			if err != nil {
				log.Error("Can't proxy request: %v", err)
				addInfoHeader(w, r, X_MOCKKA_CANT_PROXY)
				w.WriteHeader(ERROR_HTTP_CODE)
				return
			}
		}
	}

	logRequestInfo(r, rule, resp, content)
	processRequest(w, r, rule, resp, content)
}

// processRequest process http request and use found rule for formating output data
func processRequest(w http.ResponseWriter, r *http.Request, rule *rules.Rule, resp *rules.Response, content string) {
	var defResp *rules.Response
	var headers map[string]string
	var ok bool

	var code = 200

	if rule.Auth.User != "" && rule.Auth.Password != "" {
		login, password, hasAuth := r.BasicAuth()

		if !hasAuth || login != rule.Auth.User || password != rule.Auth.Password {
			w.WriteHeader(401)
			return
		}
	}

	if resp.Code == 0 {
		defResp, ok = rule.Responses[rules.DEFAULT]

		if ok && defResp.Code != 0 {
			code = defResp.Code
		}
	} else {
		code = resp.Code
	}

	if len(resp.Headers) == 0 {
		defResp, ok = rule.Responses[rules.DEFAULT]

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

// logRequestInfo create log file and write record with info about request and reponse
// into this file
func logRequestInfo(req *http.Request, rule *rules.Rule, resp *rules.Response, content string) {
	filePath := knf.GetS(MAIN_LOG_DIR) + "/" + rule.Service + ".log"
	requredPermChange := !fsutil.IsExist(filePath)

	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		log.Error(err.Error())
		return
	}

	defer fd.Close()

	err = makeLogRecord(req, rule, resp, content).Write(fd)

	if err != nil {
		log.Error(err.Error())
		return
	}

	if requredPermChange {
		updatePerms(filePath)
	}
}

// renderTemplate render output body template
func renderTemplate(req *http.Request, content string) (string, error) {
	templ, err := template.New("").Parse(content)

	if err != nil {
		return "", err
	}

	var bf bytes.Buffer

	stabber.request = req
	ct := template.Must(templ, nil)
	ct.Execute(&bf, stabber)
	stabber.request = nil

	return string(bf.Bytes()), nil
}

// getRandomResponse return random response from list of possible response bodies or default
// if response only one
func getRandomResponse(rule *rules.Rule) *rules.Response {
	var ids []string

	for id := range rule.Responses {
		if id == rules.DEFAULT {
			continue
		}

		ids = append(ids, id)
	}

	return rule.Responses[ids[rand.Int(len(ids)-1)]]
}

func makeLogRecord(req *http.Request, rule *rules.Rule, resp *rules.Response, content string) *LogRecord {
	record := &LogRecord{Date: time.Now()}

	record.Mock = rule.Path

	xForwardedFor := req.Header.Get("X-Forwarded-For")
	xRealIP := req.Header.Get("X-Real-Ip")

	switch {
	case xRealIP != "":
		record.RemoteAdress = xRealIP
	case xForwardedFor != "":
		record.RemoteAdress = strings.Split(xForwardedFor, ",")[0]
	default:
		record.RemoteAdress = req.RemoteAddr
	}

	if rule.Host != "" {
		record.RequestHost = rule.Host
	}

	record.ResponseURL = resp.URL

	record.Method = req.Method
	record.Request = req.RequestURI
	record.StatusCode = resp.Code
	record.StatusDesc = httputil.GetDescByCode(resp.Code)

	if len(req.Header) != 0 {
		record.RequestHeaders = getSortedValues(req.Header)
	}

	cookies := req.Cookies()

	if len(cookies) != 0 {
		record.Cookies = getSortedCookies(cookies)
	}

	req.ParseForm()

	if req.Method == "GET" {
		query := req.URL.Query()

		if len(query) != 0 {
			record.Query = getSortedValues(query)
		}
	} else {
		if len(req.Form) != 0 {
			record.FormData = getSortedValues(req.Form)
		}
	}

	body, err := ioutil.ReadAll(req.Body)

	if err == nil && len(body) != 0 {
		record.RequestBody = string(body[:])
	}

	if content != "" {
		record.ResponseBody = content
	}

	if len(resp.Headers) != 0 {
		record.ResponseHeaders = getSortedRespHeaders(resp.Headers)
	}

	return record
}

// updatePerms change permissions for log file/dir
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

// getSortedRespHeaders return sorted response headers
func getSortedRespHeaders(headers map[string]string) []*kv.KV {
	var result []*kv.KV

	for n, v := range headers {
		result = append(result, &kv.KV{n, v})
	}

	kv.Sort(result)

	return result
}

// getSortedValues return sorted request data
func getSortedValues(data map[string][]string) []*kv.KV {
	var result []*kv.KV

	for n, v := range data {
		result = append(result, &kv.KV{n, strings.Join(v, " ")})
	}

	kv.Sort(result)

	return result
}

// getSortedCookies return sorted cookies slice
func getSortedCookies(cookies []*http.Cookie) []string {
	var result []string

	for _, v := range cookies {
		result = append(result, v.String())
	}

	sort.Strings(result)

	return result
}

// addInfoHeader adds special header with error code
func addInfoHeader(w http.ResponseWriter, r *http.Request, code int) {
	if r.Header.Get("Mockka") != "" {
		w.Header().Add("X-Mockka-Code", strconv.Itoa(code))
	}
}

// proxyRequest used for proxying request
func proxyRequest(r *http.Request, rule *rules.Rule, resp *rules.Response) (string, error) {
	request := req.Request{
		Method: rule.Request.Method,
		URL:    resp.URL,
	}

	if len(r.Header) != 0 {
		headers := make(map[string]string)

		for n, v := range r.Header {
			headers[n] = strings.Join(v, " ")
		}

		request.Headers = headers
	}

	if r.Body != nil {
		request.Body = r.Body
	}

	prResp, err := request.Do()

	if err != nil {
		return "", err
	}

	return prResp.String(), nil
}
