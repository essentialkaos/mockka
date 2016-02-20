![Mockka Logo](https://essentialkaos.com/github/mockka-v5.png)

`Mockka` is a utility for mocking and testing HTTP API's.

#### Installation
````
go get github.com/essentialkaos/mockka
````

#### Usage

Show basic info about all mock files:
````
mockka -c /path/to/mockka.conf list
````

Run Mockka server:
````
mockka -c /path/to/mockka.conf run
````

#### Rule examples

##### Example 1
````bash
# This is comment

# Description is optional field
@DESCRIPTION
Example mock file #1

# Basic auth login and password
@AUTH
user:password

@REQUEST
GET /test/login?action=login&type=1

# You can use go template language for response body
@RESPONSE
{{ if .HeaderIs "X-Suppa-Header" "Magic" }}
{
  "status": "ok",
  "result": {
    "username":"{{ .UserName "en" }}",
    "ip": "{{ .IPv4 }}",
    "first_name": "{{ .MaleFirstName "ru" }}",
    "last_name": "{{ .MaleLastName "ru" }}",
    "city": "{{ .City "ru" }}",
    "password": "{{ .SimplePassword }}",
    "action": "{{ .Query "action" }}"
  }
}
{{ else }}
{
  "status": "error",
  "result": {
    "username":"{{ .UserName "en" }}",
    "action": "{{ .Query "action" }}"
  }
}
{{ end }}

# Code is optional field, 200 by default
@CODE
200

@HEADERS
Content-Type:application/json
X-Content-Type-Options:nosniff
X-Seraph-LoginReason:OK

# Delay has float type
@DELAY
2.0

````

##### Example 2

````bash
@DESCRIPTION
Example mock file #2

@REQUEST
GET /test/login?action=login&type=1

# You can define different responses in one mock file
@RESPONSE:1
{
  "status": "ok",
  "result": {
    "username":"{{ .UserName "en" }}"
  }
}

@RESPONSE:2
{
  "status": "error",
  "result": {}
}

# You can define different http codes for each response
@CODE:1
200

@CODE:2
502

# Or define default value for all responses
@HEADERS
Content-Type:application/json
X-Content-Type-Options:nosniff
X-Seraph-LoginReason:OK

@DELAY:1
2.0

@DELAY:2
15

````

##### Example 3

````bash
@DESCRIPTION
Example mock file #3

# If you use hostnames instead ip you can define it here
@HOST
my-suppa-host.domain.com

@REQUEST
GET /test/login?action=login&type=1

# You can store each response in file
@RESPONSE:1 < example-error.json
@RESPONSE:2 < example-ok.json
@RESPONSE:3 < example-unknown-user.json

# Or return body from proxied request
@RESPONSE:4 < https://api.domain.com/someurl

# Also you can overwrite headers and status code from rule by
# values from proxied request response
@RESPONSE:5 << https://api.domain.com/someurl

@CODE
200

@HEADERS
Content-Type:application/json

````

##### Example 4 (wildcard query)

````bash
@DESCRIPTION
Example mock file #4

@REQUEST
GET /test/status?username=*&type=1

# You can define query params in response body
@RESPONSE
{
  "status": "ok",
  "result": {
    "username": "{{ .Query "username" }}",
  }
}

@HEADERS
Content-Type:application/json

````

#### Viewer

For viewing mockka logs we provide simple tool named `mockka-viewer`.

Some features:

* Syntax higlighting
* Filtering records by time range
* Log file search

Usage:

````
Usage: mockka-viewer <options> log-file

Options:

  --no-color, -nc     Disable colors in output
  --from, -f date     Time range start
  --to, -t date       Time range end
  --help, -h          Show this help message
  --version, -v       Show version

Examples:

  mockka-viewer /path/to/file.log
  Read log file

  mockka-viewer file.log
  Try to find file.log in mockka logs directory

  mockka-viewer file
  Try to find file.log in mockka logs directory

  mockka-viewer -f 2016/01/02 -t 2016/01/05 /path/to/file.log
  Read file and show only records between 2016/01/02 and 2016/01/05

  mockka-viewer -f "2016/01/02 12:00" /path/to/file.log
  Read file and show only records between 2016/01/02 12:00 and current moment

````

#### Build Status

| Repository | Status |
|------------|--------|
| Stable | [![Build Status](https://travis-ci.org/essentialkaos/mockka.svg?branch=master)](https://travis-ci.org/essentialkaos/mockka) |
| Unstable | [![Build Status](https://travis-ci.org/essentialkaos/mockka.svg?branch=develop)](https://travis-ci.org/essentialkaos/mockka) |

#### License

[EKOL](https://essentialkaos.com/ekol)
