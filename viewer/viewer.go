package viewer

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

	"pkg.re/essentialkaos/ek.v3/arg"
	"pkg.re/essentialkaos/ek.v3/fmtc"
	"pkg.re/essentialkaos/ek.v3/fsutil"
	"pkg.re/essentialkaos/ek.v3/knf"
	"pkg.re/essentialkaos/ek.v3/path"
	"pkg.re/essentialkaos/ek.v3/sliceutil"
	"pkg.re/essentialkaos/ek.v3/strutil"
	"pkg.re/essentialkaos/ek.v3/usage"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	APP  = "Mockka Log Viewer"
	VER  = "1.0.1"
	DESC = "Utility for reading and highlighting Mockka logs"
)

const (
	ARG_NO_COLOR  = "nc:no-color"
	ARG_DATE_FROM = "f:from"
	ARG_DATE_TO   = "t:to"
	ARG_HELP      = "h:help"
	ARG_VER       = "v:version"
)

const (
	TYPE_EMPTY_LINE = 0
	TYPE_SEPARATOR  = 1
	TYPE_HEADER     = 2
	TYPE_RECORD     = 3
	TYPE_DATA       = 4
)

// ////////////////////////////////////////////////////////////////////////////////// //

// argMap is struct with command-line arguments
var argMap = arg.Map{
	ARG_NO_COLOR:  &arg.V{Type: arg.BOOL},
	ARG_DATE_FROM: &arg.V{Alias: "s:start"},
	ARG_DATE_TO:   &arg.V{Alias: "e:end"},
	ARG_HELP:      &arg.V{Type: arg.BOOL, Alias: "u:usage"},
	ARG_VER:       &arg.V{Type: arg.BOOL, Alias: "ver"},
}

// headers is slice of sections headers
var headers = []string{
	"HEADERS",
	"COOKIES",
	"QUERY",
	"REQUEST BODY",
	"RESPONSE BODY",
	"RESPONSE HEADERS",
}

// confPaths is slice with valid config paths
var confPaths = []string{
	"/etc/mockka.conf",
	"~/mockka.conf",
	"mockka.conf",
}

// separatorTimeLayout is time layout used for parsing date
var separatorTimeLayout = "2006/01/02 15:04:05"

// ////////////////////////////////////////////////////////////////////////////////// //

func Init() {
	runtime.GOMAXPROCS(1)

	args, errs := arg.Parse(argMap)

	if len(errs) != 0 {
		printError("Errors while argument parsing:")

		for _, err := range errs {
			printError("  %v", err)
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

	file := findFile(args[0])

	for {
		if fsutil.CheckPerms("FRS", file) {
			break
		}

		time.Sleep(time.Millisecond * 500)
	}

	if arg.Has(ARG_DATE_FROM) || arg.Has(ARG_DATE_TO) {
		filterFile(file)
	} else {
		readFile(file)
	}
}

// readFile starts file reading loop
func readFile(file string) {
	fd, err := os.OpenFile(file, os.O_RDONLY, 0644)

	if err != nil {
		printError("Can't open file %s: %v", file, err)
		os.Exit(1)
	}

	defer fd.Close()

	stat, err := fd.Stat()

	if err != nil {
		printError("Can't read file stats %s: %v", file, err)
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

// filterFile read file and show records only between given time range
func filterFile(file string) {
	fd, err := os.OpenFile(file, os.O_RDONLY, 0644)

	if err != nil {
		printError("Can't open file %s: %v", file, err)
		os.Exit(1)
	}

	defer fd.Close()

	reader := bufio.NewReader(fd)

	var showSection = false
	var currentSection = ""
	var dataSections = []string{"REQUEST BODY", "RESPONSE BODY"}

	fromDate := time.Date(2016, 1, 1, 0, 0, 0, 0, time.Local)
	toDate := time.Now()

	if arg.Has(ARG_DATE_FROM) {
		fromDate, err = parseRangeDate(arg.GetS(ARG_DATE_FROM))

		if err != nil {
			printError("Can't parse range start: %v", err)
			os.Exit(1)
		}
	}

	if arg.Has(ARG_DATE_TO) {
		toDate, err = parseRangeDate(arg.GetS(ARG_DATE_TO))

		if err != nil {
			printError("Can't parse range end: %v", err)
			os.Exit(1)
		}
	}

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			break
		}

		line = strings.TrimRight(line, "\n")
		line = strings.TrimRight(line, "\r")

		rt := getLineType(line)

		if rt == TYPE_SEPARATOR {
			recDateStr := extractTimeFromSeparator(line)
			recDate, _ := time.Parse(separatorTimeLayout, recDateStr)

			showSection = recDate.Unix() >= fromDate.Unix() && recDate.Unix() <= toDate.Unix()
		}

		if !showSection {
			continue
		}

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

// getLineType return data source type
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

// extractHeaderName return header name from data source
func extractHeaderName(line string) string {
	return strutil.Substr(line, 2, 99)
}

// renderLine render different type of source line
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

// findFile try to find log file
func findFile(file string) string {
	if fsutil.IsExist(file) {
		return file
	}

	configPath := fsutil.ProperPath("FRS", confPaths)

	if configPath == "" {
		return file
	}

	config, err := knf.Read(configPath)
	logDir := config.GetS("data:log-dir")

	if err != nil || logDir == "" {
		return file
	}

	if !strings.Contains(file, ".log") {
		file += ".log"
	}

	if fsutil.CheckPerms("FRS", path.Join(logDir, file)) {
		return path.Join(logDir, file)
	}

	return file
}

// extractTimeFromSeparator return string with date from separator
func extractTimeFromSeparator(line string) string {
	if line == "" || len(line) != 88 {
		return ""
	}

	return line[3:22]
}

// parseRangeDate parse different formats of time strings
// and return time struct
func parseRangeDate(date string) (time.Time, error) {
	var (
		dateStr string
		timeStr string
		layout  string
	)

	now := time.Now()

	var (
		year  = now.Year()
		month = now.Month()
		day   = now.Day()
	)

	if strings.Contains(date, " ") {
		dateSlice := strings.Split(date, " ")
		dateStr, timeStr = dateSlice[0], dateSlice[1]
	} else {
		if strings.Contains(date, "/") {
			dateStr = date
		} else {
			timeStr = date
		}
	}

	if dateStr != "" {
		switch strings.Count(dateStr, "/") {
		case 2:
			layout += "2006/01/02"
		case 1:
			layout += "2006/01"
		}
	} else {
		date = fmtc.Sprintf("%d/%02d/%d %s", year, month, day, date)
		layout = "2006/01/02"
	}

	if timeStr != "" {
		layout += " "

		switch strings.Count(timeStr, ":") {
		case 2:
			layout += "15:04:05"
		case 1:
			layout += "15:04"
		case 0:
			layout += "15"
		}
	}

	return time.Parse(layout, date)
}

// printError print error message
func printError(f string, a ...interface{}) {
	fmtc.Printf("{r}"+f+"{!}\n", a...)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func showUsage() {
	info := usage.NewInfo("", "log-file")

	info.AddOption(ARG_NO_COLOR, "Disable colors in output")
	info.AddOption(ARG_DATE_FROM, "Time range start", "date")
	info.AddOption(ARG_DATE_TO, "Time range end", "date")
	info.AddOption(ARG_HELP, "Show this help message")
	info.AddOption(ARG_VER, "Show version")

	info.AddExample("/path/to/file.log", "Read log file")
	info.AddExample("file.log", "Try to find file.log in mockka logs directory")
	info.AddExample("file", "Try to find file.log in mockka logs directory")

	info.AddExample(
		"-f 2016/01/02 -t 2016/01/05 /path/to/file.log",
		"Read file and show only records between 2016/01/02 and 2016/01/05",
	)

	info.AddExample(
		"-f \"2016/01/02 12:00\" /path/to/file.log",
		"Read file and show only records between 2016/01/02 12:00 and current moment",
	)

	info.Render()
}

func showAbout() {
	about := &usage.About{
		App:     APP,
		Version: VER,
		Desc:    DESC,
		Year:    2009,
		Owner:   "ESSENTIAL KAOS",
		License: "Essential Kaos Open Source License <https://essentialkaos.com/ekol?en>",
	}

	about.Render()
}
