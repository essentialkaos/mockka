package core

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2015 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"errors"
	"fmt"

	"github.com/essentialkaos/ek/fsutil"
	"github.com/essentialkaos/ek/knf"
	"github.com/essentialkaos/ek/system"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	MinPort         = 1024
	MaxPort         = 65535
	MinReadTimeout  = 1
	MaxReadTimeout  = 120
	MinWriteTimeout = 1
	MaxWriteTimeout = 120
	MinHeaderSize   = 1024
	MaxHeaderSize   = 10 * 1024 * 1024
	MinCheckDelay   = 1
	MaxCheckDelay   = 3600
)

const (
	ConfMainDir           = "main:dir"
	ConfMainRuleDir       = "main:rule-dir"
	ConfMainLogDir        = "main:log-dir"
	ConfMainCheckDelay    = "main:check-delay"
	ConfHTTPIp            = "http:ip"
	ConfHTTPPort          = "http:port"
	ConfHTTPReadTimeout   = "http:read-timeout"
	ConfHTTPWriteTimeout  = "http:write-timeout"
	ConfHTTPMaxHeaderSize = "http:max-header-size"
	ConfLogFile           = "log:file"
	ConfLogPerms          = "log:perms"
	ConfLogLevel          = "log:min-level"
	ConfAccessUser        = "access:user"
	ConfAccessGroup       = "access:group"
	ConfAccessMockPerms   = "access:mock-perms"
	ConfAccessLogPerms    = "access:log-perms"
	ConfAccessDirPerms    = "access:dir-perms"
	ConfListingScheme     = "listing:scheme"
	ConfListingHost       = "listing:host"
	ConfListingPort       = "listing:port"
	ConfTemplatePath      = "template:path"
)

// ////////////////////////////////////////////////////////////////////////////////// //

var Config *knf.Config
var initialized = false

// ////////////////////////////////////////////////////////////////////////////////// //

func Init(confFile string) []error {
	var err error

	if initialized {
		return []error{errors.New("Core already initialized.")}
	}

	if !fsutil.IsExist(confFile) && fsutil.IsExist("mockka.conf") {
		confFile = "mockka.conf"
	}

	Config, err = knf.Read(confFile)

	if err != nil {
		return []error{err}
	}

	return validateConfig()
}

// ////////////////////////////////////////////////////////////////////////////////// //

func validateConfig() []error {
	var permsChecker = func(config *knf.Config, prop string, value interface{}) error {
		if !fsutil.CheckPerms(value.(string), config.GetS(prop)) {
			switch value.(string) {
			case "DRX":
				return fmt.Errorf("Property %s must be path to readable directory.", prop)
			}
		}

		return nil
	}

	var userChecker = func(config *knf.Config, prop string, value interface{}) error {
		if Config.GetS(prop) == "" {
			return nil
		}

		if !system.IsUserExist(config.GetS(prop)) {
			return fmt.Errorf("User defined in %s is not exist.", prop)
		}

		return nil
	}

	var groupChecker = func(config *knf.Config, prop string, value interface{}) error {
		if Config.GetS(prop) == "" {
			return nil
		}

		if !system.IsGroupExist(Config.GetS(prop)) {
			return fmt.Errorf("Group defined in %s is not exist.", prop)
		}

		return nil
	}

	return Config.Validate([]*knf.Validator{
		&knf.Validator{ConfMainRuleDir, knf.Empty, nil},
		&knf.Validator{ConfMainLogDir, knf.Empty, nil},
		&knf.Validator{ConfMainCheckDelay, knf.Empty, nil},
		&knf.Validator{ConfHTTPPort, knf.Empty, nil},
		&knf.Validator{ConfHTTPReadTimeout, knf.Empty, nil},
		&knf.Validator{ConfHTTPWriteTimeout, knf.Empty, nil},
		&knf.Validator{ConfHTTPMaxHeaderSize, knf.Empty, nil},
		&knf.Validator{ConfAccessMockPerms, knf.Empty, nil},
		&knf.Validator{ConfAccessLogPerms, knf.Empty, nil},
		&knf.Validator{ConfAccessDirPerms, knf.Empty, nil},

		&knf.Validator{ConfMainRuleDir, permsChecker, "DRX"},
		&knf.Validator{ConfMainLogDir, permsChecker, "DRX"},

		&knf.Validator{ConfMainCheckDelay, knf.Less, MinCheckDelay},
		&knf.Validator{ConfMainCheckDelay, knf.Greater, MaxCheckDelay},
		&knf.Validator{ConfHTTPPort, knf.Less, MinPort},
		&knf.Validator{ConfHTTPPort, knf.Greater, MaxPort},
		&knf.Validator{ConfHTTPReadTimeout, knf.Less, MinReadTimeout},
		&knf.Validator{ConfHTTPReadTimeout, knf.Greater, MaxReadTimeout},
		&knf.Validator{ConfHTTPWriteTimeout, knf.Less, MinWriteTimeout},
		&knf.Validator{ConfHTTPWriteTimeout, knf.Greater, MaxWriteTimeout},
		&knf.Validator{ConfHTTPMaxHeaderSize, knf.Less, MinHeaderSize},
		&knf.Validator{ConfHTTPMaxHeaderSize, knf.Greater, MaxHeaderSize},

		&knf.Validator{ConfAccessUser, userChecker, nil},
		&knf.Validator{ConfAccessGroup, groupChecker, nil},
	})
}

// ////////////////////////////////////////////////////////////////////////////////// //
