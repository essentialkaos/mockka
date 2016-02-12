package urlutil

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"net/url"
	"sort"
	"strings"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Match return true if url match given pattern
func Match(pattern, url string) bool {
	if pattern == "" {
		return false
	}

	if pattern == url {
		return true
	}

	var (
		patternIndex  = 0
		urlIndex      = 0
		patternSymbol = ""
		urlSymbol     = ""
		patternLength = len(pattern)
		urlLength     = len(url)
		ignoreQuery   = false
		queryPart     = false
	)

	if patternLength > urlLength {
		return false
	}

	for ; urlIndex < urlLength; urlIndex++ {

		patternSymbol = pattern[patternIndex : patternIndex+1]
		urlSymbol = url[urlIndex : urlIndex+1]

		if patternSymbol == "?" && patternSymbol == urlSymbol {
			queryPart = true
			patternIndex++
			continue
		}

		if patternSymbol == "*" {
			if !queryPart {
				ignoreQuery = true
			}

			if queryPart && !ignoreQuery {
				if patternIndex+1 < patternLength {
					if pattern[patternIndex+1:patternIndex+2] == url[urlIndex+1:urlIndex+2] || urlSymbol == "&" {
						patternIndex += 1
					}
				} else {
					if urlSymbol == "&" {
						return false
					}
				}
			} else {
				if patternIndex+1 == patternLength {
					return true
				}

				if pattern[patternIndex+1:patternIndex+2] == url[urlIndex+1:urlIndex+2] {
					patternIndex += 1
				}
			}
		} else {
			if patternSymbol != urlSymbol {
				return false
			} else {
				patternIndex++

				if patternIndex == patternLength {
					return urlIndex+1 == urlLength
				}
			}
		}
	}

	return patternIndex+1 == patternLength
}

// EqualPatterns compare two patterns
func EqualPatterns(pattern1, pattern2 string) bool {
	if Match(pattern1, pattern2) || Match(pattern2, pattern1) {
		return true
	}

	return false
}

// SortURLParams return url with sorted get parameters
func SortURLParams(u *url.URL) string {
	query := u.Query()

	if len(query) == 0 {
		return u.RequestURI()
	}

	result := u.Path + "?"

	var sortedQuery []string

	for qp := range query {
		sortedQuery = append(sortedQuery, qp)
	}

	sort.Strings(sortedQuery)

	for _, qp := range sortedQuery {
		value := strings.Join(query[qp], "")

		if value == "" {
			result += qp + "&"
		} else {
			result += qp + "=" + value + "&"
		}
	}

	result = result[0 : len(result)-1]

	if u.Fragment != "" {
		result += "#" + u.Fragment
	}

	return result
}

// SortParams return url with sorted get parameters
func SortParams(path string) string {
	if !strings.Contains(path, "?") {
		return path
	}

	u, err := url.Parse(path)

	if err != nil {
		return path
	}

	return SortURLParams(u)
}
