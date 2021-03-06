BASEDIR=$(CURDIR)
TOOLDIR=$(BASEDIR)/script

BINARY=evoimage
SOURCES := $(shell find $(BASEDIR) -name '*.go')
TESTED=.tested
IMGDIR=./imgs

build: $(BINARY)
$(BINARY): $(SOURCES) $(TESTED)
	go build

install: build
	go install

dist: build
	$(TOOLDIR)/dist

clean:
	rm -f $(BINARY) debug debug.test cover.out $(TESTED)
	go clean

format:
	go fmt *.go

lint: format
	go vet

test: $(TESTED) $(TESTRESOURCES)
$(TESTED): $(SOURCES)
	$(TOOLDIR)/test

testv: clean $(VERSIONOUT)
	$(TOOLDIR)/test -v

cover: $(SOURCES) $(VERSIONOUT)
	$(TOOLDIR)/cover

update: clean
	$(TOOLDIR)/update

.PHONY: clean test testv cover build run update install format dist lint
