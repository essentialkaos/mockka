package listing

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"errors"
	"fmt"

	"pkg.re/essentialkaos/ek.v1/fmtc"
	"pkg.re/essentialkaos/ek.v1/fmtutil"
	"pkg.re/essentialkaos/ek.v1/knf"
	"pkg.re/essentialkaos/ek.v1/netutil"

	"github.com/essentialkaos/mockka/rules"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	LISTING_SCHEME = "listing:scheme"
	LISTING_HOST   = "listing:host"
	LISTING_PORT   = "listing:port"
)

// ////////////////////////////////////////////////////////////////////////////////// //

var ip = netutil.GetIP()

// ////////////////////////////////////////////////////////////////////////////////// //

func List(observer *rules.Observer, service string) error {
	if observer == nil {
		return errors.New("Observer is not created")
	}

	observer.Load()

	if service != "" {
		err := listService(observer, service)

		if err != nil {
			return err
		}
	} else {
		services := observer.GetServices()

		if len(services) == 0 {
			fmtc.Println("\n{y}No services and mocks are created{!}\n")
			return nil
		}

		for _, serviceName := range services {
			err := listService(observer, serviceName)

			if err != nil {
				return nil
			}
		}
	}

	fmt.Println("")

	return nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

func listService(observer *rules.Observer, service string) error {
	rulesNames := observer.GetServiceRulesNames(service)

	if len(rulesNames) == 0 {
		return fmt.Errorf("Service %s is not found", service)
	}

	fmtc.Printf("\n{*r}%s{!} {s}(%s){!}\n", service, fmtutil.Pluralize(len(rulesNames), "mock", "mocks"))

	for _, ruleName := range rulesNames {
		rule := observer.GetRuleByName(service, ruleName)

		if rule == nil {
			continue
		}

		showRuleInfo(rule)
	}

	return nil
}

func showRuleInfo(rule *rules.Rule) {
	if rule.Desc == "" {
		fmtc.Printf("\n  {*}%s{!} {s}(Description is empty){!}\n", rule.FullName+".mock")
	} else {
		fmtc.Printf("\n  {*}%s{!} - %s\n", rule.FullName+".mock", rule.Desc)
	}

	host := rule.Request.Host

	if host == "" {
		if knf.GetS(LISTING_HOST) != "" {
			host = knf.GetS(LISTING_HOST)
		} else {
			host = ip
		}
	}

	if knf.GetS(LISTING_SCHEME) != "" {
		host = knf.GetS(LISTING_SCHEME) + "://" + host
	}

	if knf.GetS(LISTING_PORT) != "" {
		host = host + ":" + knf.GetS(LISTING_PORT)
	}

	fmtc.Printf("  {s}%s %s%s{!}\n", rule.Request.Method, host, rule.Request.URL)
}
