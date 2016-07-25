<p align="center"><a href="#installation">Installation</a> • <a href="#first-steps">First steps</a> • <a href="#rule-examples">Rule examples</a> • <a href="#viewer">Viewer</a> • <a href="#usage">Usage</a> • <a href="#build-status">Build Status</a> • <a href="#license">License</a></p>

<p align="center">
<img width="300" height="150" src="https://essentialkaos.com/github/mockka.png"/>
</p>

`Mockka` is a utility for mocking and testing HTTP API's.

## Installation
To build the Mockka from scratch, make sure you have a working Go 1.5+ workspace ([instructions](https://golang.org/doc/install)), then:

```
go get github.com/essentialkaos/mockka
```

If you want update Mockka to latest stable release, do:

```
go get -u github.com/essentialkaos/mockka
```

## First steps

Show basic info about all mock files:
````
mockka -c /path/to/mockka.conf list
````

Run Mockka server:
````
mockka -c /path/to/mockka.conf run
````

By default Mockka try to find configuration file in next locations:

* `/etc/mockka.conf`
* `~/.mockka.conf` (`$HOME` directory)
* `./mockka.conf` (current directory)

## Rule examples

Latest version of Mockka have full [Fake](https://github.com/icrowley/fake) package support. List of all supported functions can be found [here](https://github.com/essentialkaos/mockka/wiki/Supported-Stubber-Methods).

#### Example 1
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

#### Example 2

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

#### Example 3

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

#### Example 4 (wildcard query)

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

## Viewer

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

## Usage

```
Usage: mockka <command> <options>

Commands:

  run                  Run mockka server
  check mock-file      Check rule for problems
  make mock-name       Create mock file from template
  list service-name    Show list of exist rules

Options:

  --config, -c file        Path to config file
  --port, -p 1024-65535    Overwrite port
  --daemon, -d             Run server in daemon mode
  --no-color, -nc          Disable colors in output
  --help, -h               Show this help message
  --version, -v            Show version

Examples:

  mockka -c /path/to/mockka.conf run
  Run mockka server and use config /path/to/mockka.conf

  mockka make service1/test1
  Create file test1.mock for service service1.

  mockka check service1/test1
  Check rule file test1.mock for service service1

  mockka check service1/test1
  Check all rules of service service1

  mockka list
  List all rules

  mockka list service1
  List service1 rules

```

## Build Status

| Branch | Status |
|------------|--------|
| `master` | [![Build Status](https://travis-ci.org/essentialkaos/mockka.svg?branch=master)](https://travis-ci.org/essentialkaos/mockka) |
| `develop` | [![Build Status](https://travis-ci.org/essentialkaos/mockka.svg?branch=develop)](https://travis-ci.org/essentialkaos/mockka) |

## License

[EKOL](https://essentialkaos.com/ekol)
