########################################################################################

DESTDIR ?= /usr/bin

########################################################################################

.PHONY = all clean install uninstall deps test

########################################################################################

all: mockka mockka-viewer

deps:
	go get -v pkg.re/check.v1
	go get -v pkg.re/essentialkaos/ek.v1
	go get -v github.com/icrowley/fake
	go get -v golang.org/x/tools/cmd/cover

mockka:
	go build mockka.go

mockka-viewer:
	go build mockka-viewer.go

test:
	go test ./rules 
	go test ./urlutil

install:
	mkdir -p $(DESTDIR)
	cp mockka $(DESTDIR)/
	cp mockka-viewer $(DESTDIR)/

uninstall:
	rm -f $(DESTDIR)/mockka
	rm -f $(DESTDIR)/mockka-viewer

clean:
	rm -f mockka
	rm -f mockka-viewer
