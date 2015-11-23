package main

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2015 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"ek/arg"
	"ek/fmtc"
	"ek/usage"
	"fmt"
	"log"
	"mockka/core"
	"mockka/generator"
	"mockka/listing"
	"mockka/rules"
	"mockka/server"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	AppName = "Mockka"
	AppVer  = "1.6.2"
	AppDesc = "Utility for mockking HTTP API's"
)

const (
	ArgConfig  = "c:config"
	ArgPort    = "p:port"
	ArgDaemon  = "d:daemon"
	ArgNoColor = "nc:no-color"
	ArgHelp    = "h:help"
	ArgVer     = "v:version"
)

const (
	CommandRun  = "run"
	CommandList = "list"
	CommandMake = "make"
)

// ////////////////////////////////////////////////////////////////////////////////// //

var argList = map[string]*arg.V{
	ArgConfig:  &arg.V{Value: "/etc/mockka.conf"},
	ArgPort:    &arg.V{Type: arg.Int, Min: core.MinPort, Max: core.MaxPort},
	ArgDaemon:  &arg.V{Type: arg.Bool},
	ArgNoColor: &arg.V{Type: arg.Bool},
	ArgHelp:    &arg.V{Type: arg.Bool, Alias: "u:usage"},
	ArgVer:     &arg.V{Type: arg.Bool, Alias: "ver"},
}

var logFd *os.File

// ////////////////////////////////////////////////////////////////////////////////// //

func main() {
	runtime.GOMAXPROCS(1)

	if len(os.Args) <= 1 {
		showUsage()
		return
	}

	args, errs := arg.Parse(argList)

	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Println(err.Error())
		}

		os.Exit(1)
	}

	if arg.GetB(ArgNoColor) {
		fmtc.DisableColors = true
	}

	if arg.GetB(ArgVer) {
		showAbout()
		return
	}

	if len(args) == 0 || arg.GetB(ArgHelp) {
		showUsage()
		return
	}

	errs = core.Init(arg.GetS(ArgConfig))

	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Println(err.Error())
		}

		os.Exit(1)
	}

	if arg.GetB(ArgDaemon) {
		registerSignalHandlers()
		setupLog()
	}

	execCommand(args)
}

func setupLog() {
	if logFd != nil {
		logFd.Close()
	}

	logFd, err := os.OpenFile(core.Config.GetS(core.ConfLogFile), os.O_WRONLY|os.O_CREATE|os.O_APPEND, core.Config.GetM(core.ConfLogPerms))

	if err != nil {
		fmt.Println("Can't open log file " + core.Config.GetS(core.ConfLogFile) + " for writing")
		os.Exit(1)
	}

	log.SetOutput(logFd)
}

func execCommand(args []string) {
	command := args[0]

	switch command {
	case CommandRun:
		runServer()

	case CommandMake:
		makeMock(args[1:])

	case CommandList:
		listMocks(args[1:])

	default:
		printError(fmt.Sprintf("Unknown command %s", command))
		os.Exit(1)
	}
}

func registerSignalHandlers() {
	c := make(chan os.Signal)

	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for {
			sig := <-c

			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				log.Println("Received TERM or INT signal, shutdown...")
				os.Exit(0)
			case syscall.SIGHUP:
				log.Println("Received HUP signal, log reopened")
				setupLog()
			}
		}
	}()
}

func runServer() {
	observer := rules.NewObserver(core.Config.GetS(core.ConfMainRuleDir))
	observer.Start(core.Config.GetI(core.ConfMainCheckDelay))

	err := server.Start(observer, AppName+"/"+AppVer, arg.GetS(ArgPort))

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

	if len(args) != 0 {
		service = args[0]
	}

	observer := rules.NewObserver(core.Config.GetS(core.ConfMainRuleDir))
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

func showUsage() {
	info := usage.NewInfo("mck")

	info.AddCommand(CommandRun, "Run mockka server")
	info.AddCommand(CommandMake, "Create mock file from template", "name")
	info.AddCommand(CommandList, "Show list of exist rules")

	info.AddOption(ArgConfig, "Path to config file", "file")
	info.AddOption(ArgPort, "Overwrite port", fmt.Sprintf("%d-%d", core.MinPort, core.MaxPort))
	info.AddOption(ArgDaemon, "Run server in daemon mode")
	info.AddOption(ArgNoColor, "Disable colors in output")
	info.AddOption(ArgHelp, "Show this help message")
	info.AddOption(ArgVer, "Show version")

	info.AddExample("run")
	info.AddExample("make service1/test1")
	info.AddExample("list")

	info.Render()
}

func showAbout() {
	about := &usage.About{
		App:     AppName,
		Version: AppVer,
		Desc:    AppDesc,
		Year:    2009,
		Owner:   "ESSENTIALKAOS",
		License: "Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>",
	}

	about.Render()
}
