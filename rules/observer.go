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
	"net/http"
	"path"
	"sort"
	"strings"
	"time"

	"pkg.re/essentialkaos/ek.v1/fsutil"
	"pkg.re/essentialkaos/ek.v1/httputil"
	"pkg.re/essentialkaos/ek.v1/log"
)

// ////////////////////////////////////////////////////////////////////////////////// //

type Observer struct {
	AutoHead bool

	uriMap  map[string]*Rule            // method+url -> rule
	pathMap map[string]*Rule            // full path -> rule
	wcMap   map[string]*Rule            // Wilcard string -> rule
	nameMap map[string]map[string]*Rule // service -> full name (with dir) -> rule
	errMap  map[string]bool             // full name -> has error
	srvMap  map[string]bool             // service name -> true

	ruleDir string
	works   bool
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Create new observer struct
func NewObserver(ruleDir string) *Observer {
	return &Observer{
		ruleDir: ruleDir,
		uriMap:  make(map[string]*Rule),
		pathMap: make(map[string]*Rule),
		wcMap:   make(map[string]*Rule),
		nameMap: make(map[string]map[string]*Rule),
		errMap:  make(map[string]bool),
		srvMap:  make(map[string]bool),
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Start observer
func (obs *Observer) Start(checkDelay int) {
	if obs.works {
		return
	}

	obs.works = true

	go obs.watch(checkDelay)
}

// Load rules
func (obs *Observer) Load() []string {
	var messages []string

	for _, r := range obs.uriMap {
		if !fsutil.IsExist(r.Path) {
			delete(obs.uriMap, r.URI())
			delete(obs.errMap, r.URI())
			delete(obs.pathMap, r.Path)
			delete(obs.nameMap[r.Service], r.FullName)

			if r.Wildcard != "" {
				delete(obs.wcMap, r.WilcardURI())
			}

			// If no one rule found for service, remove it's own map
			if len(obs.nameMap[r.Service]) == 0 {
				delete(obs.nameMap, r.Service)
			}

			messages = append(messages, fmt.Sprintf("Rule %s unloaded (mock file deleted)", path.Join(r.Service, r.FullName)))

			continue
		}

		mtime, _ := fsutil.GetMTime(r.Path)

		if r.ModTime.UnixNano() != mtime.UnixNano() {
			rule, err := Parse(obs.ruleDir, r.Service, r.Dir, r.Name)

			if err != nil {
				messages = append(messages, fmt.Sprintf("[ERROR] Can't parse rule file - %s", err.Error()))
				continue
			}

			// URI can be changed, remove rule from uri map anyway
			delete(obs.uriMap, r.URI())
			delete(obs.errMap, r.URI())

			if r.Wildcard != "" {
				delete(obs.wcMap, r.WilcardURI())
			}

			obs.uriMap[rule.URI()] = rule
			obs.pathMap[rule.Path] = rule

			if rule.Wildcard != "" {
				obs.wcMap[rule.WilcardURI()] = rule
			}

			messages = append(messages, fmt.Sprintf("Rule %s reloaded", path.Join(rule.Service, rule.FullName)))
		}
	}

	dl, err := ioutil.ReadDir(obs.ruleDir)

	if err != nil {
		messages = append(messages, fmt.Sprintf("Can't list directory with rules (%s)", obs.ruleDir))
	}

	for _, di := range dl {
		// Ignore all files in rules directory
		if !di.IsDir() {
			continue
		}

		service := di.Name()
		messages = append(messages, obs.checkDir(service, "")...)
	}

	return messages
}

// Get rule struct by request struct
func (obs *Observer) GetRule(r *http.Request) *Rule {
	var rule *Rule

	autoHead := obs.AutoHead && r.Method == "HEAD"

	rule = findRule(obs.uriMap, r, false, autoHead)

	if rule != nil {
		return rule
	}

	if len(obs.wcMap) != 0 {
		rule = findRule(obs.wcMap, r, true, autoHead)

		if rule != nil {
			return rule
		}
	}

	return nil
}

// Get rule by full name (i.e. service/dir/mock>)
func (obs *Observer) GetRuleByName(service, name string) *Rule {
	if !obs.srvMap[service] {
		return nil
	}

	return obs.nameMap[service][name]
}

// Get services names list
func (obs *Observer) GetServices() []string {
	var result []string

	if len(obs.srvMap) == 0 {
		return result
	}

	for service := range obs.srvMap {
		result = append(result, service)
	}

	sort.Strings(result)

	return result
}

// Get rules full names (with dirs)
func (obs *Observer) GetServiceRulesNames(service string) []string {
	var result []string

	if !obs.srvMap[service] {
		return result
	}

	for name := range obs.nameMap[service] {
		result = append(result, name)
	}

	sort.Strings(result)

	return result
}

// ////////////////////////////////////////////////////////////////////////////////// //

func (obs *Observer) checkDir(service, dir string) []string {
	var messages []string

	rl, err := ioutil.ReadDir(path.Join(obs.ruleDir, service, dir))

	if err != nil {
		messages = append(messages, fmt.Sprintf("[ERROR] %s", err.Error()))
		return messages
	}

	for _, ri := range rl {
		filename := ri.Name()

		if ri.IsDir() && filename[0:1] != "." {
			messages = append(messages, obs.checkDir(service, path.Join(dir, filename))...)
			continue
		}

		// Ignore all files without .mock extension
		if path.Ext(filename) != ".mock" {
			continue
		}

		fullpath := path.Join(obs.ruleDir, service, dir, filename)
		_, readed := obs.pathMap[fullpath]

		if readed {
			continue
		}

		rule, err := Parse(obs.ruleDir, service, dir, strings.Replace(filename, ".mock", "", -1))

		if err != nil {
			if obs.errMap[rule.URI()] != true {
				messages = append(messages, fmt.Sprintf("[ERROR] Can't parse rule file - %s", err.Error()))
				obs.errMap[rule.URI()] = true
			}

			continue
		}

		if obs.uriMap[rule.URI()] != nil || obs.wcMap[rule.WilcardURI()] != nil {
			if obs.errMap[rule.URI()] != true {
				messages = append(messages, fmt.Sprintf("[ERROR] Can't apply rule from %s - rule already exist for given method/url pair", path.Join(rule.Service, rule.FullName)))
				obs.errMap[rule.URI()] = true
			}

			continue
		}

		delete(obs.errMap, rule.URI())

		obs.uriMap[rule.URI()] = rule
		obs.pathMap[rule.Path] = rule
		obs.srvMap[service] = true

		if rule.Wildcard != "" {
			obs.wcMap[rule.WilcardURI()] = rule
		}

		if obs.nameMap[service] == nil {
			obs.nameMap[service] = make(map[string]*Rule)
		}

		obs.nameMap[service][rule.FullName] = rule

		messages = append(messages, fmt.Sprintf("Rule %s loaded", path.Join(rule.Service, rule.FullName)))
	}

	return messages
}

func (obs *Observer) watch(checkDelay int) {
	for {
		messages := obs.Load()

		if len(messages) != 0 {
			for _, message := range messages {
				log.Info(message)
			}
		}

		time.Sleep(time.Duration(checkDelay) * time.Second)
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

func findRule(data map[string]*Rule, r *http.Request, wildcard, autoHead bool) *Rule {
	var result *Rule
	var uri string

	host := httputil.GetRequestHost(r)

	if !wildcard {
		uri = getSortedRequestURI(r)
	} else {
		uri = getQueryWildcard(r.URL.Query())
	}

	result = getRule(data, host, r.Method, uri)

	if result != nil {
		return result
	}

	if !autoHead {
		return nil
	}

	for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
		result = getRule(data, host, method, uri)

		if result != nil {
			return result
		}
	}

	return nil
}

func getRule(data map[string]*Rule, host, method, uri string) *Rule {
	var result *Rule

	result = data[host+":"+method+":"+uri]

	if result != nil {
		return result
	}

	result = data[":"+method+":"+uri]

	return result
}

func getSortedRequestURI(r *http.Request) string {
	if !strings.Contains(r.RequestURI, "?") {
		return r.RequestURI
	}

	query := r.URL.Query()
	result := r.URL.Path + "?"

	var sortedQuery []string

	for qp := range query {
		sortedQuery = append(sortedQuery, qp)
	}

	sort.Strings(sortedQuery)

	for _, qp := range sortedQuery {
		result += qp + "=" + strings.Join(query[qp], "") + "&"
	}

	return result[0 : len(result)-1]
}
