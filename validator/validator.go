package validator

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"errors"
	"pkg.re/essentialkaos/ek.v1/sliceutil"
	"strings"

	"pkg.re/essentialkaos/ek.v1/fmtc"
	"pkg.re/essentialkaos/ek.v1/fmtutil"
	"pkg.re/essentialkaos/ek.v1/fsutil"
	"pkg.re/essentialkaos/ek.v1/httputil"
	"pkg.re/essentialkaos/ek.v1/knf"
	"pkg.re/essentialkaos/ek.v1/mathutil"
	"pkg.re/essentialkaos/ek.v1/path"
	"pkg.re/essentialkaos/ek.v1/rand"

	"github.com/essentialkaos/mockka/rules"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	PROBLEM_NONE uint8 = 0
	PROBLEM_WARN       = 1
	PROBLEM_ERR        = 2
)

const (
	DATA_RULE_DIR = "data:rule-dir"
)

// ////////////////////////////////////////////////////////////////////////////////// //

type RuleInfo struct {
	Service string
	Name    string
	Dir     string
}

type Problem struct {
	Type uint8  // Problem type (none/warn/error)
	Info string // Info (short info)
	Desc string // Description (long info)
}

type Validator func(r *rules.Rule) []*Problem

// ////////////////////////////////////////////////////////////////////////////////// //

// errorMessages is slice with error messages
var errorMessages = []string{
	"Errors... It's so sad :(",
	"OMG errors? Dude, I know, you can fix it!",
	"Errors? Again?! If you don't fix it, I will no longer check you rules!",
	"Red messages it's mean errors? I hate errors. Please fix it. PLEEEAAASE.",
	"Errors? It's unacceptable. (c) Lemongrab",
}

// warnMessages is slice with warning messages
var warnMessages = []string{
	"I found warnings. If you fix that I would be so happy...",
	"Warnings? Yellow messages sucks, isn't it?",
	"I hate warnings... Maybe you can fix it?",
	"Nobody likes warnings like this. I think cool guys like you can fix it easy.",
}

// okMessages is slice with ok messages
var okMessages = []string{
	"WOW! All rules is fine. And this is awesome!",
	"Zero warnings?! It's amazing!",
	"You write rules like a pro!",
	"All ok. I am a program, but I really appreciate this.",
	"Hey, your rules it's much better than other. But let it be our little secret.",
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Check check rule or all rules for some service
func Check(target string) error {
	if target == "" {
		return errors.New("You must difine mock file or service name")
	}

	targetService, targetMock, targetDir := rules.ParsePath(target)

	serviceDir := path.Join(knf.GetS(DATA_RULE_DIR), targetService)

	if !fsutil.IsExist(serviceDir) {
		return fmtc.Errorf("Service %s is not exist", targetService)
	}

	var ruleInfoSlice []*RuleInfo

	if targetMock != "" {
		ruleInfoSlice = append(ruleInfoSlice, &RuleInfo{targetService, targetMock, targetDir})
	} else {
		mockFiles := fsutil.ListAllFiles(serviceDir, true,
			&fsutil.ListingFilter{MatchPatterns: []string{"*.mock"}},
		)

		for _, mockFile := range mockFiles {
			mockPath := path.Join(targetService, strings.TrimRight(mockFile, ".mock"))

			_, targetMock, targetDir := rules.ParsePath(mockPath)

			ruleInfoSlice = append(ruleInfoSlice, &RuleInfo{targetService, targetMock, targetDir})
		}
	}

	if len(ruleInfoSlice) == 0 {
		fmtc.Println("\n{y}No mock's were found{!}\n")
		return nil
	}

	var maxProblemType = PROBLEM_NONE

	for _, rule := range ruleInfoSlice {
		maxProblemType = mathutil.MaxU8(checkRule(rule), maxProblemType)
	}

	if maxProblemType > PROBLEM_NONE {
		fmtutil.Separator(false)
	}

	switch maxProblemType {
	case PROBLEM_NONE:
		fmtc.Printf("\n{g}%s{!}\n\n", okMessages[rand.Int(len(okMessages))])
	case PROBLEM_WARN:
		fmtc.Printf("{y}%s{!}\n\n", warnMessages[rand.Int(len(warnMessages))])
	case PROBLEM_ERR:
		fmtc.Printf("{r}%s{!}\n\n", errorMessages[rand.Int(len(errorMessages))])
	}

	return nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// checkRule parse rule, run validators and print check results
func checkRule(ruleInfo *RuleInfo) uint8 {
	var problems []*Problem

	rule, err := rules.Parse(knf.GetS(DATA_RULE_DIR), ruleInfo.Service, ruleInfo.Dir, ruleInfo.Name)

	if err != nil {
		problems = append(problems,
			&Problem{
				Type: PROBLEM_ERR,
				Info: "Parsing error",
				Desc: err.Error(),
			},
		)
	} else {

		validators := []Validator{
			checkDescription,
			checkWildcard,
			checkMethod,
			checkStatusCode,
			checkContent,
		}

		problems = append(problems, execValidators(validators, rule)...)
	}

	if len(problems) == 0 {
		return PROBLEM_NONE
	}

	fmtutil.Separator(false, path.Join(ruleInfo.Service, ruleInfo.Dir, ruleInfo.Name))
	renderProblems(problems)

	var maxProblemType = PROBLEM_NONE

	for _, problem := range problems {
		maxProblemType = mathutil.MaxU8(problem.Type, maxProblemType)
	}

	return maxProblemType
}

// execValidators run all validators
func execValidators(validators []Validator, rule *rules.Rule) []*Problem {
	var result []*Problem

	for _, validator := range validators {
		result = append(result, validator(rule)...)
	}

	return result
}

// renderProblems print error and warn messages
func renderProblems(problems []*Problem) {
	for _, problem := range problems {
		switch problem.Type {
		case PROBLEM_WARN:
			fmtc.Printf("{y}WARNING →{!} {*}%s{!}\n\n", problem.Info)
		case PROBLEM_ERR:
			fmtc.Printf("{r}ERROR →{!} {*}%s{!}\n\n", problem.Info)
		}

		fmtc.Println(fmtutil.Wrap(problem.Desc, "  ", 86))
		fmtc.NewLine()
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// checkDescription check rule description for problems
func checkDescription(r *rules.Rule) []*Problem {
	var result []*Problem

	if r.Desc == "" {
		result = append(result,
			&Problem{
				Type: PROBLEM_WARN,
				Info: "Description is empty",
				Desc: "Description is important part of rule. Please provide short (30-140 chars) info about what this rule do.",
			},
		)
	} else {
		if len(r.Desc) < 16 {
			result = append(result,
				&Problem{
					Type: PROBLEM_WARN,
					Info: "Description is too short",
					Desc: "Description is important part of rule. Please provide short (30-140 chars) info about what this rule do.",
				},
			)
		}

		if len(r.Desc) > 240 {
			result = append(result,
				&Problem{
					Type: PROBLEM_WARN,
					Info: "Description is too long",
					Desc: "Description is important part of rule. Please provide short (30-140 chars) info about what this rule do.",
				},
			)
		}
	}

	return result
}

// checkWildcard check rule URL with wildcard URL for problems
func checkWildcard(r *rules.Rule) []*Problem {
	var result []*Problem

	if r.Request.URL == "/*" {
		result = append(result,
			&Problem{
				Type: PROBLEM_ERR,
				Info: "Global wildcard",
				Desc: "You define global wildcard. It's means what this rule handle ALL request. Avoid to use global wildcard and try define more detailed URL.",
			},
		)
	}

	return result
}

// checkMethod check rule request method name for problems
func checkMethod(r *rules.Rule) []*Problem {
	var result []*Problem

	if !sliceutil.Contains([]string{"OPTIONS", "GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "CONNECT", "PATCH"}, r.Request.Method) {
		result = append(result,
			&Problem{
				Type: PROBLEM_ERR,
				Info: "Unknown HTTP method",
				Desc: fmtc.Sprintf("You define unsupported HTTP method \"%s\". Valid methods is OPTIONS, GET, HEAD, POST, PUT, DELETE, TRACE, CONNECT and PATCH.", r.Request.Method),
			},
		)
	}

	return result
}

// checkStatusCode check rule responses status code for problems
func checkStatusCode(r *rules.Rule) []*Problem {
	var result []*Problem

	for respId, resp := range r.Responses {
		sectionId := "@CODE"

		if respId != rules.DEFAULT {
			sectionId += ":" + respId
		}

		if resp.Code != 0 && httputil.GetDescByCode(resp.Code) == "" {
			if respId == rules.DEFAULT {
				result = append(result,
					&Problem{
						Type: PROBLEM_WARN,
						Info: "Unknown status code",
						Desc: fmtc.Sprintf("You define unknown status code %d in %s section. Please check list of valid status codes https://en.wikipedia.org/wiki/List_of_HTTP_status_codes", resp.Code, sectionId),
					},
				)
			}
		}
	}

	return result
}

// checkResponseDelay check rule responses delay for problems
func checkResponseDelay(r *rules.Rule) []*Problem {
	var result []*Problem

	for respId, resp := range r.Responses {
		sectionId := "@DELAY"

		if respId != rules.DEFAULT {
			sectionId += ":" + respId
		}

		if resp.Delay > 60 {
			result = append(result,
				&Problem{
					Type: PROBLEM_WARN,
					Info: "Response delay is too big",
					Desc: fmtc.Sprintf("Response delay is greater than maximum delay (60 seconds).", sectionId),
				},
			)
		}
	}

	return result
}

// checkContent check rule content for problems
func checkContent(r *rules.Rule) []*Problem {
	var result []*Problem

	for respId, resp := range r.Responses {

		sectionId := "@RESPONSE"

		if respId != rules.DEFAULT {
			sectionId += ":" + respId
		}

		if resp.File != "" && resp.Content != "" {
			result = append(result,
				&Problem{
					Type: PROBLEM_ERR,
					Info: "Response body have two sources",
					Desc: fmtc.Sprintf("You define two different sources for response body (file and content in response section) in section %s. Please use only one source (file OR content in response section).", sectionId),
				},
			)
		}

		if resp.URL != "" && resp.Content != "" {
			result = append(result,
				&Problem{
					Type: PROBLEM_ERR,
					Info: "Response body have two sources",
					Desc: fmtc.Sprintf("You define two different sources for response body (url and content in response section) in section %s. Please use only one source (url OR content in response section).", sectionId),
				},
			)
		}

		if resp.File == "" && resp.Content == "" && resp.URL != "" {
			if resp.Code == 200 || resp.Code == 0 {
				result = append(result,
					&Problem{
						Type: PROBLEM_ERR,
						Info: "Response body is empty",
						Desc: fmtc.Sprintf("Section %s doesn't contains any response data.", sectionId),
					},
				)
			}
		}

		if resp.File != "" && !fsutil.IsExist(resp.File) {
			result = append(result,
				&Problem{
					Type: PROBLEM_ERR,
					Info: "File with response is not exist",
					Desc: fmtc.Sprintf("File %s defined in %s section is not exist.", resp.File, sectionId),
				},
			)
		} else {
			if resp.File != "" && !fsutil.IsReadable(resp.File) {
				result = append(result,
					&Problem{
						Type: PROBLEM_ERR,
						Info: "File with response is not readable",
						Desc: fmtc.Sprintf("File %s defined in %s section is not readable.", resp.File, sectionId),
					},
				)
			}
		}
	}

	return result
}
