########################################################################################

DEST_DIR ?= /usr/bin

########################################################################################

.PHONY = all clean install uninstall deps

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

install:
	mkdir -p $(DEST_DIR)
	cp mockka $(DEST_DIR)/
	cp mockka-viewer $(DEST_DIR)/

uninstall:
	rm -f $(DEST_DIR)/mockka
	rm -f $(DEST_DIR)/mockka-viewer

clean:
	rm -f mockka
	rm -f mockka-viewer
