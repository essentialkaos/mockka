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
	"strings"
	"text/template"
	"time"

	"pkg.re/essentialkaos/ek.v3/crypto"
	"pkg.re/essentialkaos/ek.v3/fsutil"
	"pkg.re/essentialkaos/ek.v3/httputil"
	"pkg.re/essentialkaos/ek.v3/knf"
	"pkg.re/essentialkaos/ek.v3/kv"
	"pkg.re/essentialkaos/ek.v3/log"
	"pkg.re/essentialkaos/ek.v3/mathutil"
	"pkg.re/essentialkaos/ek.v3/path"
	"pkg.re/essentialkaos/ek.v3/rand"
	"pkg.re/essentialkaos/ek.v3/req"
	"pkg.re/essentialkaos/ek.v3/system"

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
	DATA_LOG_DIR              = "data:log-dir"
	DATA_LOG_TYPE             = "data:log-type"
	HTTP_IP                   = "http:ip"
	HTTP_PORT                 = "http:port"
	HTTP_READ_TIMEOUT         = "http:read-timeout"
	HTTP_WRITE_TIMEOUT        = "http:write-timeout"
	HTTP_MAX_HEADER_SIZE      = "http:max-header-size"
	HTTP_MAX_DELAY            = "http:max-delay"
	PROCESSING_ALLOW_PROXYING = "processing:allow-proxying"
	ACCESS_USER               = "access:user"
	ACCESS_GROUP              = "access:group"
	ACCESS_MOCK_PERMS         = "access:mock-perms"
	ACCESS_LOG_PERMS          = "access:log-perms"
	ACCESS_MOCK_DIR_PERMS     = "access:mock-dir-perms"
	ACCESS_LOG_DIR_PERMS      = "access:log-dir-perms"
)

// ////////////////////////////////////////////////////////////////////////////////// //

var (
	serverToken string
	observer    *rules.Observer
	stabber     *Stabber
)

var errorDesc = map[int]string{
	X_MOCKKA_NO_RULE:     "RuleNotFound",
	X_MOCKKA_NO_RESPONSE: "ResponseNotFound",
	X_MOCKKA_CANT_RENDER: "CantRenderTemplate",
	X_MOCKKA_CANT_PROXY:  "CantProxyRequest",
	X_MOCKKA_FORBIDDEN:   "ForbidenAction",
}

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

	uuid := crypto.GenUUID()

	log.Debug("Request: %s → %s%v (%s)", r.Method, r.Host, r.URL, uuid)

	w.Header().Set("Server", serverToken)

	rule = observer.GetRule(r)

	if rule == nil {
		log.Error("Can't find rule for request %s → %s%s", r.Method, r.Host, r.URL.String())
		writeError(w, r, X_MOCKKA_NO_RULE)
		return
	}

	switch len(rule.Responses) {
	case 0:
		log.Error("Can't find rule for request %s → %s%s", r.Method, r.Host, r.URL.String())
		writeError(w, r, X_MOCKKA_NO_RESPONSE)
		return
	case 1:
		resp = rule.Responses[rules.DEFAULT]
	default:
		resp = getRandomResponse(rule)
	}

	var responseContent string
	var bodyData []byte

	if r.Method != "HEAD" {
		if resp.URL == "" {
			responseContent, err = renderTemplate(r, resp.Body())

			if err != nil {
				log.Error("Can't render response body: %v", err)
				writeError(w, r, X_MOCKKA_CANT_RENDER)
				return
			}
		} else {
			if !knf.GetB(PROCESSING_ALLOW_PROXYING) {
				log.Error("Can't proxy request: proxying disabled in configuration file")
				writeError(w, r, X_MOCKKA_FORBIDDEN)
				return
			}

			responseContent, bodyData, resp, err = proxyRequest(r, rule, resp)

			if err != nil {
				log.Error("Can't proxy request: %v", err)
				writeError(w, r, X_MOCKKA_CANT_PROXY)
				return
			}
		}
	}

	log.Debug("<%s:RULE> → %v", uuid, rule)
	log.Debug("<%s:REQ>  → %v", uuid, rule.Request)
	log.Debug("<%s:RESP> → %v", uuid, resp)

	logRequestInfo(r, rule, resp, responseContent, bodyData)
	processRequest(w, r, rule, resp, responseContent)
}

// processRequest process http request and use found rule for formating output data
func processRequest(w http.ResponseWriter, r *http.Request, rule *rules.Rule, resp *rules.Response, responseContent string) {
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
		delay := mathutil.BetweenF(resp.Delay, 0.0, knf.GetF(HTTP_MAX_DELAY, 60.0)) * float64(time.Second)
		time.Sleep(time.Duration(delay))
	}

	for k, v := range headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(code)
	w.Write([]byte(responseContent))
}

// logRequestInfo create log file and write record with info about request and reponse
// into this file
func logRequestInfo(req *http.Request, rule *rules.Rule, resp *rules.Response, responseContent string, bodyData []byte) {
	logPath, err := getLogStore(rule)

	if err != nil {
		log.Error(err.Error())
		return
	}

	requredPermChange := !fsutil.IsExist(logPath)

	err = makeLogRecord(req, rule, resp, responseContent, bodyData).Write(logPath)

	if err != nil {
		log.Error(err.Error())
		return
	}

	if requredPermChange {
		updatePerms(logPath, knf.GetM(ACCESS_LOG_PERMS, 0644))
	}
}

// renderTemplate render output body template
func renderTemplate(req *http.Request, responseContent string) (string, error) {
	templ, err := template.New("").Parse(responseContent)

	if err != nil {
		return "", err
	}

	var bf bytes.Buffer

	stabber.request = req
	ct := template.Must(templ, nil)
	err = ct.Execute(&bf, stabber)
	stabber.request = nil

	if err != nil {
		return "", err
	}

	return bf.String(), nil
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

// makeLogRecord create log record struct
func makeLogRecord(req *http.Request, rule *rules.Rule, resp *rules.Response, responseContent string, bodyData []byte) *LogRecord {
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

	if rule.Request.Host != "" {
		record.RequestHost = rule.Request.Host
	}

	record.ResponseURL = resp.URL

	record.Method = req.Method
	record.Request = req.RequestURI

	record.StatusCode = 200

	if resp.Code == 0 {
		defResp, ok := rule.Responses[rules.DEFAULT]

		if ok && defResp.Code != 0 {
			record.StatusCode = defResp.Code
		}
	} else {
		record.StatusCode = resp.Code
	}

	record.StatusDesc = httputil.GetDescByCode(record.StatusCode)

	if len(req.Header) != 0 {
		record.RequestHeaders = getSortedValues(req.Header)
	}

	cookies := req.Cookies()

	if len(cookies) != 0 {
		record.Cookies = getSortedCookies(cookies)
	}

	if req.Method == "GET" {
		query := req.URL.Query()

		if len(query) != 0 {
			record.Query = getSortedValues(query)
		}
	}

	if req.ContentLength > 0 {
		if len(bodyData) != 0 {
			record.RequestBody = string(bodyData[:])
		} else {
			body, err := ioutil.ReadAll(req.Body)

			if err == nil && len(body) != 0 {
				record.RequestBody = string(body[:])
			}
		}
	}

	if responseContent != "" {
		record.ResponseBody = responseContent
	}

	if len(resp.Headers) != 0 {
		record.ResponseHeaders = getSortedRespHeaders(resp.Headers)
	}

	return record
}

// getLogStore create directory for log file and return full path to log
func getLogStore(rule *rules.Rule) (string, error) {
	if knf.GetS(DATA_LOG_TYPE, "united") == "united" {
		return path.Join(knf.GetS(DATA_LOG_DIR), rule.Service+".log"), nil
	}

	logDir := path.Join(rule.Service, rule.Dir)
	logDirSlice := strings.Split(logDir, "/")

	for i := 1; i < len(logDirSlice)+1; i++ {
		pathPart := path.Join(logDirSlice[0:i]...)
		logDirPath := path.Join(knf.GetS(DATA_LOG_DIR), pathPart)

		if fsutil.IsExist(logDirPath) {
			continue
		}

		err := os.Mkdir(logDirPath, 0775)

		if err != nil {
			return "", err
		}

		updatePerms(logDirPath, knf.GetM(ACCESS_LOG_DIR_PERMS, 0775))
	}

	return path.Join(knf.GetS(DATA_LOG_DIR), logDir, rule.Name+".log"), nil
}

// updatePerms change permissions for log file/dir
func updatePerms(path string, perms os.FileMode) {
	if knf.HasProp(ACCESS_USER) || knf.HasProp(ACCESS_GROUP) {
		logOwnerUID, logOwnerGID, _ := fsutil.GetOwner(path)

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

		os.Chown(path, logOwnerUID, logOwnerGID)
	}

	os.Chmod(path, perms)
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
func writeError(w http.ResponseWriter, r *http.Request, code int) {
	w.Header().Add("X-Mockka-Error", errorDesc[code])
	w.WriteHeader(ERROR_HTTP_CODE)
}

// proxyRequest used for proxying request
func proxyRequest(r *http.Request, rule *rules.Rule, resp *rules.Response) (string, []byte, *rules.Response, error) {
	var (
		err  error
		body []byte
	)

	request := req.Request{
		Method: rule.Request.Method,
		URL:    resp.URL,
	}

	// Append headers from initial request
	if len(r.Header) != 0 {
		headers := make(map[string]string)

		for n, v := range r.Header {
			headers[n] = strings.Join(v, " ")
		}

		request.Headers = headers
	}

	// If we have request body send it
	// Because request.Body is ReadCloser we must read body data
	// for logging and return it from this method
	if r.ContentLength > 0 {
		body, err = ioutil.ReadAll(r.Body)

		if err != nil {
			return "", nil, resp, err
		}

		request.Body = body
	}

	respData, err := request.Do()

	if err != nil {
		return "", nil, resp, err
	}

	resultResp := resp

	// If overwrite flag set for response, we return headers and
	// status code from proxied request
	if resp.Overwrite {
		resultResp = &rules.Response{
			Delay:   resp.Delay,
			Code:    respData.StatusCode,
			Headers: headersToMap(respData.Header),
		}
	}

	return respData.String(), body, resultResp, nil
}

// headersToMap convert headers to map with strings
func headersToMap(headers http.Header) map[string]string {
	var result = make(map[string]string)

	for k, v := range headers {
		result[k] = strings.Join(v, " ")
	}

	return result
}
