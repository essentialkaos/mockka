package listing

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2015 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"errors"
	"fmt"

	"github.com/essentialkaos/ek/fmtc"
	"github.com/essentialkaos/ek/netutil"

	"github.com/essentialkaos/mockka/core"
	"github.com/essentialkaos/mockka/rules"
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
			fmtc.Println("\n{y}No services and mocks are created.{!}\n")
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
		return fmt.Errorf("Service %s is not found.", service)
	}

	fmtc.Printf("\n{*r}%s{!} {s}(%d mocks){!}\n", service, len(rulesNames))

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

	host := rule.Host

	if host == "" {
		if core.Config.GetS(core.ConfListingHost) != "" {
			host = core.Config.GetS(core.ConfListingHost)
		} else {
			host = ip
		}
	}

	if core.Config.GetS(core.ConfListingScheme) != "" {
		host = core.Config.GetS(core.ConfListingScheme) + "://" + host
	}

	if core.Config.GetS(core.ConfListingPort) != "" {
		host = host + ":" + core.Config.GetS(core.ConfListingPort)
	}

	fmtc.Printf("  {s}%s %s%s{!}\n", rule.Request.Method, host, rule.Request.URL)
}
