package main

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bufio"
	"os"
	"runtime"
	"strings"
	"time"

	"pkg.re/essentialkaos/ek.v1/arg"
	"pkg.re/essentialkaos/ek.v1/fmtc"
	"pkg.re/essentialkaos/ek.v1/fsutil"
	"pkg.re/essentialkaos/ek.v1/sliceutil"
	"pkg.re/essentialkaos/ek.v1/strutil"
	"pkg.re/essentialkaos/ek.v1/usage"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	APP  = "Mockka Viewer"
	VER  = "1.0.0"
	DESC = "Utility for reading and highlighting Mockka logs"
)

const (
	ARG_NO_COLOR = "nc:no-color"
	ARG_HELP     = "h:help"
	ARG_VER      = "v:version"
)

const (
	TYPE_EMPTY_LINE = 0
	TYPE_SEPARATOR  = 1
	TYPE_HEADER     = 2
	TYPE_RECORD     = 3
	TYPE_DATA       = 4
)

// ////////////////////////////////////////////////////////////////////////////////// //

var argMap = arg.Map{
	ARG_NO_COLOR: &arg.V{Type: arg.BOOL},
	ARG_HELP:     &arg.V{Type: arg.BOOL, Alias: "u:usage"},
	ARG_VER:      &arg.V{Type: arg.BOOL, Alias: "ver"},
}

var headers = []string{
	"HEADERS", "COOKIES", "QUERY", "FORM DATA",
	"REQUEST BODY", "RESPONSE BODY", "RESPONSE HEADERS",
}

// ////////////////////////////////////////////////////////////////////////////////// //

func main() {
	runtime.GOMAXPROCS(1)

	args, errs := arg.Parse(argMap)

	if len(errs) != 0 {
		fmtc.Println("{r}Errors while argument parsing:{!}")

		for _, err := range errs {
			fmtc.Printf("  {r}%v{!}\n", err)
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

	file := args[0]

	checkFile(file)
	readFile(file)
}

func checkFile(file string) {
	if !fsutil.IsExist(file) {
		fmtc.Printf("{r}File %s is not exist{!}", file)
		os.Exit(1)
	}

	if !fsutil.IsReadable(file) {
		fmtc.Printf("{r}File %s is not readable{!}", file)
		os.Exit(1)
	}
}

func readFile(file string) {
	fd, err := os.OpenFile(file, os.O_RDONLY|os.O_APPEND, 0644)

	if err != nil {
		fmtc.Printf("{r}Can't open file %s: %v{!}\n", file, err)
		os.Exit(1)
	}

	defer fd.Close()

	stat, err := fd.Stat()

	if err != nil {
		fmtc.Printf("{r}Can't read file stats %s: %v{!}\n", file, err)
		os.Exit(1)
	}

	if stat.Size() > 2048 {
		fd.Seek(-2048, 2)
	}

	reader := bufio.NewReader(fd)

	nearRecordFound := false

	var currentSection = ""
	var dataSections = []string{"REQUEST BODY", "RESPONSE BODY"}

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			time.Sleep(time.Millisecond * 250)
			continue
		}

		line = strings.TrimRight(line, "\n")
		line = strings.TrimRight(line, "\r")

		if !nearRecordFound {
			if strutil.Head(line, 3) != "-- " {
				continue
			}

			nearRecordFound = true
		}

		rt := getLineType(line)

		if rt == TYPE_HEADER {
			currentSection = extractHeaderName(line)
			renderLine(line, rt)
			continue
		}

		if sliceutil.Contains(dataSections, currentSection) {
			fmtc.Println(line)
			continue
		}

		renderLine(line, rt)
	}
}

func getLineType(line string) int {
	if line == "" {
		return TYPE_EMPTY_LINE
	}

	if strutil.Head(line, 3) == "-- " {
		return TYPE_SEPARATOR
	}

	if strutil.Head(line, 2) == "+ " {
		if sliceutil.Contains(headers, strutil.Substr(line, 2, 99)) {
			return TYPE_HEADER
		}
	}

	if strutil.Head(line, 2) == "  " && strings.Contains(line, ":") {
		return TYPE_RECORD
	}

	return TYPE_DATA
}

func extractHeaderName(line string) string {
	return strutil.Substr(line, 2, 99)
}

func renderLine(line string, dataType int) {
	switch dataType {
	case TYPE_EMPTY_LINE:
		fmtc.NewLine()
	case TYPE_SEPARATOR:
		fmtc.Printf("{s}%s{!} %s {s}%s{!}\n", line[:3], line[3:22], line[22:])
	case TYPE_HEADER:
		fmtc.Printf("{s}%s{!}\n", line)
	case TYPE_RECORD:
		sepIndex := strings.Index(line, ":")
		fmtc.Printf("{*}%s{!} %s\n", line[:sepIndex+1], line[sepIndex+1:])
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

func showUsage() {
	info := usage.NewInfo("", "log-file")

	info.AddOption(ARG_NO_COLOR, "Disable colors in output")
	info.AddOption(ARG_HELP, "Show this help message")
	info.AddOption(ARG_VER, "Show version")

	info.Render()
}

func showAbout() {
	about := &usage.About{
		App:     APP,
		Version: VER,
		Desc:    DESC,
		Year:    2009,
		Owner:   "ESSENTIAL KAOS",
		License: "Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>",
	}

	about.Render()
}
