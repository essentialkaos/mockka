package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"pkg.re/essentialkaos/ek.v1/arg"
	"pkg.re/essentialkaos/ek.v1/fmtc"
	"pkg.re/essentialkaos/ek.v1/fsutil"
	"pkg.re/essentialkaos/ek.v1/knf"
	"pkg.re/essentialkaos/ek.v1/log"
	"pkg.re/essentialkaos/ek.v1/signal"
	"pkg.re/essentialkaos/ek.v1/system"
	"pkg.re/essentialkaos/ek.v1/usage"

	"github.com/essentialkaos/mockka/generator"
	"github.com/essentialkaos/mockka/listing"
	"github.com/essentialkaos/mockka/rules"
	"github.com/essentialkaos/mockka/server"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	APP  = "Mockka"
	VER  = "1.7.0"
	DESC = "Utility for mockking HTTP API's"
)

const (
	ARG_CONFIG   = "c:config"
	ARG_PORT     = "p:port"
	ARG_DAEMON   = "d:daemon"
	ARG_NO_COLOR = "nc:no-color"
	ARG_HELP     = "h:help"
	ARG_VER      = "v:version"
)

const (
	MIN_PORT          = 1024
	MAX_PORT          = 65535
	MIN_READ_TIMEOUT  = 1
	MAX_READ_TIMEOUT  = 120
	MIN_WRITE_TIMEOUT = 1
	MAX_WRITE_TIMEOUT = 120
	MIN_HEADER_SIZE   = 1024
	MAX_HEADER_SIZE   = 10 * 1024 * 1024
	MIN_CHECK_DELAY   = 1
	MAX_CHECK_DELAY   = 3600
)

const (
	MAIN_DIR                  = "main:dir"
	DATA_RULE_DIR             = "data:rule-dir"
	DATA_LOG_DIR              = "data:log-dir"
	DATA_LOG_TYPE             = "data:log-type"
	DATA_CHECK_DELAY          = "data:check-delay"
	HTTP_IP                   = "http:ip"
	HTTP_PORT                 = "http:port"
	HTTP_READ_TIMEOUT         = "http:read-timeout"
	HTTP_WRITE_TIMEOUT        = "http:write-timeout"
	HTTP_MAX_HEADER_SIZE      = "http:max-header-size"
	PROCESSING_AUTO_HEAD      = "processing:auto-head"
	PROCESSING_ALLOW_PROXYING = "processing:allow-proxying"
	LOG_DIR                   = "log:dir"
	LOG_FILE                  = "log:file"
	LOG_PERMS                 = "log:perms"
	LOG_LEVEL                 = "log:level"
	ACCESS_USER               = "access:user"
	ACCESS_GROUP              = "access:group"
	ACCESS_MOCK_PERMS         = "access:mock-perms"
	ACCESS_LOG_PERMS          = "access:log-perms"
	ACCESS_MOCK_DIR_PERMS     = "access:mock-dir-perms"
	ACCESS_LOG_DIR_PERMS      = "access:log-dir-perms"
	LISTING_SCHEME            = "listing:scheme"
	LISTING_HOST              = "listing:host"
	LISTING_PORT              = "listing:port"
	TEMPLATE_PATH             = "template:path"
)

const (
	COMMAND_RUN  = "run"
	COMMAND_LIST = "list"
	COMMAND_MAKE = "make"
)

// ////////////////////////////////////////////////////////////////////////////////// //

var argMap = arg.Map{
	ARG_CONFIG:   &arg.V{Value: "/etc/mockka.conf"},
	ARG_PORT:     &arg.V{Type: arg.BOOL, Min: MIN_PORT, Max: MAX_PORT},
	ARG_DAEMON:   &arg.V{Type: arg.BOOL},
	ARG_NO_COLOR: &arg.V{Type: arg.BOOL},
	ARG_HELP:     &arg.V{Type: arg.BOOL, Alias: "u:usage"},
	ARG_VER:      &arg.V{Type: arg.BOOL, Alias: "ver"},
}

// ////////////////////////////////////////////////////////////////////////////////// //

func Init() {
	var err error
	var errs []error

	runtime.GOMAXPROCS(1)

	if len(os.Args) <= 1 {
		showUsage()
		return
	}

	args, errs := arg.Parse(argMap)

	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Println(err.Error())
		}

		os.Exit(1)
	}

	if arg.GetB(ARG_NO_COLOR) {
		fmtc.DisableColors = true
	}

	if arg.GetB(ARG_VER) {
		showAbout()
		return
	}

	if len(args) == 0 || arg.GetB(ARG_HELP) {
		showUsage()
		return
	}

	err = knf.Global(arg.GetS(ARG_CONFIG))

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	errs = validateConfig()

	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Println(err.Error())
		}

		os.Exit(1)
	}

	setupLog()

	if arg.GetB(ARG_DAEMON) {
		registerSignalHandlers()
	}

	execCommand(args)
}

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
		if config.GetS(prop) == "" {
			return nil
		}

		if !system.IsUserExist(config.GetS(prop)) {
			return fmt.Errorf("User defined in %s is not exist.", prop)
		}

		return nil
	}

	var groupChecker = func(config *knf.Config, prop string, value interface{}) error {
		if config.GetS(prop) == "" {
			return nil
		}

		if !system.IsGroupExist(config.GetS(prop)) {
			return fmt.Errorf("Group defined in %s is not exist.", prop)
		}

		return nil
	}

	return knf.Validate([]*knf.Validator{
		&knf.Validator{DATA_RULE_DIR, knf.Empty, nil},
		&knf.Validator{DATA_LOG_DIR, knf.Empty, nil},
		&knf.Validator{DATA_CHECK_DELAY, knf.Empty, nil},
		&knf.Validator{HTTP_PORT, knf.Empty, nil},
		&knf.Validator{HTTP_READ_TIMEOUT, knf.Empty, nil},
		&knf.Validator{HTTP_WRITE_TIMEOUT, knf.Empty, nil},
		&knf.Validator{HTTP_MAX_HEADER_SIZE, knf.Empty, nil},
		&knf.Validator{ACCESS_MOCK_PERMS, knf.Empty, nil},
		&knf.Validator{ACCESS_LOG_PERMS, knf.Empty, nil},
		&knf.Validator{ACCESS_MOCK_DIR_PERMS, knf.Empty, nil},
		&knf.Validator{ACCESS_LOG_DIR_PERMS, knf.Empty, nil},

		&knf.Validator{DATA_RULE_DIR, permsChecker, "DRX"},
		&knf.Validator{DATA_LOG_DIR, permsChecker, "DRX"},

		&knf.Validator{DATA_CHECK_DELAY, knf.Less, MIN_CHECK_DELAY},
		&knf.Validator{DATA_CHECK_DELAY, knf.Greater, MAX_CHECK_DELAY},
		&knf.Validator{HTTP_PORT, knf.Less, MIN_PORT},
		&knf.Validator{HTTP_PORT, knf.Greater, MAX_PORT},
		&knf.Validator{HTTP_READ_TIMEOUT, knf.Less, MIN_READ_TIMEOUT},
		&knf.Validator{HTTP_READ_TIMEOUT, knf.Greater, MAX_READ_TIMEOUT},
		&knf.Validator{HTTP_WRITE_TIMEOUT, knf.Less, MIN_WRITE_TIMEOUT},
		&knf.Validator{HTTP_WRITE_TIMEOUT, knf.Greater, MAX_WRITE_TIMEOUT},
		&knf.Validator{HTTP_MAX_HEADER_SIZE, knf.Less, MIN_HEADER_SIZE},
		&knf.Validator{HTTP_MAX_HEADER_SIZE, knf.Greater, MAX_HEADER_SIZE},

		&knf.Validator{ACCESS_USER, userChecker, nil},
		&knf.Validator{ACCESS_GROUP, groupChecker, nil},
	})
}

func setupLog() {
	levels := map[string]int{
		"debug": log.DEBUG,
		"info":  log.INFO,
		"warn":  log.WARN,
		"error": log.ERROR,
		"crit":  log.CRIT,
	}

	log.MinLevel(levels[strings.ToLower(knf.GetS(LOG_LEVEL, "debug"))])

	if arg.GetB(ARG_DAEMON) {
		err := log.Set(knf.GetS(LOG_FILE), knf.GetM(LOG_PERMS, 0644))

		if err != nil {
			fmt.Printf("Can't setup logger: %s\n", err.Error())
			os.Exit(1)
		}
	}
}

func execCommand(args []string) {
	command := args[0]

	switch command {
	case COMMAND_RUN:
		runServer()

	case COMMAND_MAKE:
		makeMock(args[1:])

	case COMMAND_LIST:
		listMocks(args[1:])

	default:
		printError(fmt.Sprintf("Unknown command %s", command))
		os.Exit(1)
	}
}

func registerSignalHandlers() {
	signal.Handlers{
		signal.INT:  intSignalHandler,
		signal.TERM: termSignalHandler,
		signal.HUP:  hupSignalHandler,
	}.TrackAsync()
}

func runServer() {
	observer := rules.NewObserver(knf.GetS(DATA_RULE_DIR))
	observer.AutoHead = knf.GetB(PROCESSING_AUTO_HEAD)
	observer.Start(knf.GetI(DATA_CHECK_DELAY))

	err := server.Start(observer, APP+"/"+VER, arg.GetS(ARG_PORT))

	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}
}

func makeMock(args []string) {
	var name = ""

	if len(args) != 0 {
		name = args[0]
	}

	err := generator.Make(name)

	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}
}

func listMocks(args []string) {
	var service = ""

	// Suppress observer logging
	log.Set(os.DevNull, 0)

	if len(args) != 0 {
		service = args[0]
	}

	observer := rules.NewObserver(knf.GetS(DATA_RULE_DIR))
	err := listing.List(observer, service)

	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}
}

func printError(message string) {
	fmt.Printf("\n%s\n\n", message)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func intSignalHandler() {
	log.Info("Received INT signal, shutdown...")
	os.Exit(0)
}

func termSignalHandler() {
	log.Info("Received TERM signal, shutdown...")
	os.Exit(0)
}

func hupSignalHandler() {
	log.Info("Received HUP signal, log reopened")
	log.Reopen()
}

// ////////////////////////////////////////////////////////////////////////////////// //

func showUsage() {
	info := usage.NewInfo("")

	info.AddCommand(COMMAND_RUN, "Run mockka server")
	info.AddCommand(COMMAND_MAKE, "Create mock file from template", "name")
	info.AddCommand(COMMAND_LIST, "Show list of exist rules")

	info.AddOption(ARG_CONFIG, "Path to config file", "file")
	info.AddOption(ARG_PORT, "Overwrite port", fmt.Sprintf("%d-%d", MIN_PORT, MAX_PORT))
	info.AddOption(ARG_DAEMON, "Run server in daemon mode")
	info.AddOption(ARG_NO_COLOR, "Disable colors in output")
	info.AddOption(ARG_HELP, "Show this help message")
	info.AddOption(ARG_VER, "Show version")

	info.AddExample("-c /path/to/mockka.conf run")
	info.AddExample("-c /path/to/mockka.conf make service1/test1")
	info.AddExample("-c /path/to/mockka.conf list")

	info.Render()
}

func showAbout() {
	about := &usage.About{
		App:     APP,
		Version: VER,
		Release: ".beta1",
		Desc:    DESC,
		Year:    2009,
		Owner:   "ESSENTIAL KAOS",
		License: "Essential Kaos Open Source License <https://essentialkaos.com/ekol?en>",
	}

	about.Render()
}
