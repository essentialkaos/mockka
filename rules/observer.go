package rules

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"pkg.re/essentialkaos/ek.v1/fsutil"
	"pkg.re/essentialkaos/ek.v1/httputil"
	"pkg.re/essentialkaos/ek.v1/log"
	"pkg.re/essentialkaos/ek.v1/path"

	"github.com/essentialkaos/mockka/urlutil"
)

// ////////////////////////////////////////////////////////////////////////////////// //

type RuleMap map[string]*Rule

type Observer struct {
	AutoHead bool

	uriMap  RuleMap            // host+method+url -> rule
	pathMap RuleMap            // full path -> rule
	wcMap   RuleMap            // full path -> rule (only wildcard)
	nameMap map[string]RuleMap // service -> full name (with dir) -> rule
	errMap  map[string]bool    // full name -> has error
	srvMap  map[string]bool    // service name -> true

	ruleDir string
	works   bool
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Create new observer struct
func NewObserver(ruleDir string) *Observer {
	return &Observer{
		ruleDir: ruleDir,
		uriMap:  make(RuleMap),
		pathMap: make(RuleMap),
		wcMap:   make(RuleMap),
		nameMap: make(map[string]RuleMap),
		errMap:  make(map[string]bool),
		srvMap:  make(map[string]bool),
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Start start observer
func (obs *Observer) Start(checkDelay int) {
	if obs.works {
		return
	}

	obs.works = true

	go obs.watch(time.Duration(checkDelay) * time.Second)
}

// Load load and parse all rules
func (obs *Observer) Load() bool {
	var ok = true

	for _, r := range obs.uriMap {
		if !fsutil.IsExist(r.Path) {
			delete(obs.uriMap, r.Request.URI)
			delete(obs.wcMap, r.Path)
			delete(obs.errMap, r.Path)
			delete(obs.pathMap, r.Path)
			delete(obs.nameMap[r.Service], r.FullName)

			// If no one rule found for service, remove it's own map
			if len(obs.nameMap[r.Service]) == 0 {
				delete(obs.nameMap, r.Service)
			}

			log.Info("Rule %s unloaded (mock file deleted)", r.PrettyPath)

			continue
		}

		mtime, _ := fsutil.GetMTime(r.Path)

		if r.ModTime.UnixNano() != mtime.UnixNano() {
			rule, err := Parse(obs.ruleDir, r.Service, r.Dir, r.Name)

			if err != nil {
				log.Error("Can't parse rule file: %v", err)
				ok = false
				continue
			}

			// URI can be changed, remove rule from uri map anyway
			delete(obs.uriMap, r.Request.URI)
			delete(obs.errMap, r.Path)

			if r.IsWildcard {
				delete(obs.wcMap, r.Path)
			}

			obs.uriMap[rule.Request.URI] = rule
			obs.pathMap[rule.Path] = rule

			if rule.IsWildcard {
				obs.wcMap[rule.Path] = rule
			}

			log.Info("Rule %s reloaded", rule.PrettyPath)
		}
	}

	dl, err := ioutil.ReadDir(obs.ruleDir)

	if err != nil {
		log.Error("Can't list directory with rules (%s)", obs.ruleDir)
	}

	for _, di := range dl {

		// Ignore all files in rules directory
		if !di.IsDir() {
			continue
		}

		service := di.Name()

		if !obs.checkDir(service, "") && ok {
			ok = false
		}
	}

	return ok
}

// GetRule return rule for request
func (obs *Observer) GetRule(r *http.Request) *Rule {
	autoHead := obs.AutoHead && r.Method == "HEAD"
	return findRule(obs.uriMap, obs.wcMap, r, autoHead)
}

// GetRuleByName return rule by full name (i.e. service/dir/mock>)
func (obs *Observer) GetRuleByName(service, name string) *Rule {
	if !obs.srvMap[service] {
		return nil
	}

	return obs.nameMap[service][name]
}

// GetServices return services names list
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

// GetServiceRulesNames return rules full names (with dirs)
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

func (obs *Observer) checkDir(service, dir string) bool {
	var ok = true

	rl, err := ioutil.ReadDir(path.Join(obs.ruleDir, service, dir))

	if err != nil {
		log.Error(err.Error())
		return false
	}

RULELOOP:
	for _, ri := range rl {
		filename := ri.Name()

		if ri.IsDir() && !path.IsDotfile(filename) {
			if !obs.checkDir(service, path.Join(dir, filename)) {
				ok = false
			}

			continue
		}

		// Ignore all files without .mock extension (backup files, temporary files)
		if path.Ext(filename) != ".mock" {
			continue
		}

		fullpath := path.Join(obs.ruleDir, service, dir, filename)

		// Skip rule if it already successfully parsed
		if obs.pathMap[fullpath] != nil {
			continue
		}

		rule, err := Parse(obs.ruleDir, service, dir, strings.Replace(filename, ".mock", "", -1))

		if err != nil {
			if obs.errMap[rule.Path] != true {
				log.Error("Can't parse rule %s: %v", rule.PrettyPath, err)
				obs.errMap[rule.Path] = true
				ok = false
			}

			continue
		}

		for _, r := range obs.wcMap {
			if r.Request.Method != rule.Request.Method {
				continue
			}

			if r.Request.Host != rule.Request.Host {
				continue
			}

			if urlutil.EqualPatterns(r.Request.NURL, rule.Request.NURL) {
				if obs.errMap[rule.Path] != true {
					log.Error("Rule intersection: rule %s and rule %s have same result urls", rule.PrettyPath, r.PrettyPath)
					obs.errMap[rule.Path] = true
					ok = false
				}

				continue RULELOOP
			}
		}

		delete(obs.errMap, rule.Path)

		obs.uriMap[rule.Request.URI] = rule
		obs.pathMap[rule.Path] = rule
		obs.srvMap[service] = true

		if rule.IsWildcard {
			obs.wcMap[rule.Path] = rule
		}

		if obs.nameMap[service] == nil {
			obs.nameMap[service] = make(RuleMap)
		}

		obs.nameMap[service][rule.FullName] = rule

		log.Info("Rule %s loaded", rule.PrettyPath)
	}

	return ok
}

func (obs *Observer) watch(checkDelay time.Duration) {
	for {
		obs.Load()
		time.Sleep(checkDelay)
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

func findRule(uriMap, wcMap RuleMap, r *http.Request, autoHead bool) *Rule {
	var result *Rule

	host := httputil.GetRequestHost(r)
	uri := urlutil.SortURLParams(r.URL)

	log.Debug("Request: %s%s", host, uri)

	result = getRule(uriMap, host, r.Method, uri)

	if result != nil {
		return result
	}

	if autoHead {
		for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
			result = getRule(uriMap, host, method, uri)

			if result != nil {
				return result
			}
		}
	}

	if len(wcMap) == 0 {
		return nil
	}

	for _, rule := range wcMap {
		if !autoHead && rule.Request.Method != r.Method {
			continue
		}

		if rule.Request.Host != "" && host != rule.Request.Host {
			continue
		}

		// For matching we use normalized url (with sorted get params)
		if urlutil.Match(rule.Request.NURL, uri) {
			return rule
		}
	}

	return nil
}

func getRule(ruleMap RuleMap, host, method, uri string) *Rule {
	var result *Rule

	result = ruleMap[host+":"+method+":"+uri]

	if result != nil {
		return result
	}

	result = ruleMap[":"+method+":"+uri]

	return result
}
