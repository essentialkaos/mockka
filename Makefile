########################################################################################

DESTDIR?=
PREFIX?=/usr

########################################################################################

.PHONY = all clean install uninstall deps test

########################################################################################

all: mockka mockka-viewer

deps:
	go get -v pkg.re/check.v1
	go get -v pkg.re/essentialkaos/ek.v3
	go get -v github.com/icrowley/fake
	go get -v golang.org/x/tools/cmd/cover

mockka:
	go build mockka.go

mockka-viewer:
	go build mockka-viewer.go

test:
	go test ./rules ./urlutil

install:
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	cp mockka $(DESTDIR)$(PREFIX)/bin/
	cp mockka-viewer $(DESTDIR)$(PREFIX)/bin/

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/mockka
	rm -f $(DESTDIR)$(PREFIX)/bin/mockka-viewer

clean:
	rm -f mockka
	rm -f mockka-viewer
