all: bin server plugins archive upgrade_script

GOLANGCI_LINT_VERSION = v1.45.2

# This part of the makefile is adapted from https://gist.github.com/grihabor/4a750b9d82c9aa55d5276bd5503829be
DESCRIBE           := $(shell git tag | sort -V -r | head -n 1)

ifeq ($(DESCRIBE),)
DESCRIBE = v0.0.0
endif

DESCRIBE_PARTS     := $(subst -, ,$(DESCRIBE))

VERSION_TAG        := $(word 1,$(DESCRIBE_PARTS))
COMMITS_SINCE_TAG  := $(word 2,$(DESCRIBE_PARTS))

VERSION            := $(subst v,,$(VERSION_TAG))
VERSION_PARTS      := $(subst ., ,$(VERSION))

MAJOR              := $(word 1,$(VERSION_PARTS))
MINOR              := $(word 2,$(VERSION_PARTS))
MICRO              := $(word 3,$(VERSION_PARTS))

NEXT_MAJOR         := $(shell echo $$(($(MAJOR)+1)))
NEXT_MINOR         := $(shell echo $$(($(MINOR)+1)))
NEXT_MICRO         := $(shell echo $$(($(MICRO)+1)))

_dirty_files       := $(shell git status --untracked-files=no --porcelain | wc -l)
ifeq ($(_dirty_files),0)
DIRTY := false
else
DIRTY := true
endif

HASH               := $(shell git rev-parse --short HEAD)
COMMITS_SINCE_TAG  := $(shell git log $(shell git describe --tags --abbrev=0)..HEAD --oneline | wc -l)
BUILD_USER         := $(shell whoami)


ifeq ($(BUMP),)
BUMP := micro
endif

ifeq ($(MAJOR),)
MAJOR := 0
endif

ifeq ($(MINOR),)
MINOR := 0
endif

ifeq ($(MICRO),)
MICRO := 0
endif

ifeq ($(BUMP),minor)
BUMPED_VERSION_NO_V := $(MAJOR).$(NEXT_MINOR).0
endif
ifeq ($(BUMP),major)
BUMPED_VERSION_NO_V := $(NEXT_MAJOR).0.0
endif
ifeq ($(BUMP),micro)
BUMPED_VERSION_NO_V := $(MAJOR).$(MINOR).$(NEXT_MICRO)
endif

BUMPED_VERSION := v$(BUMPED_VERSION_NO_V)
VERSION_NO_V := $(MAJOR).$(MINOR).$(MICRO)

ifneq ($(COMMITS_SINCE_TAG),0)
VERSION_NO_V := $(VERSION_NO_V)-$(COMMITS_SINCE_TAG)
endif

ifeq ($(DIRTY),true)
VERSION_NO_V := $(VERSION_NO_V)-$(HASH)-dirty-$(BUILD_USER)
endif

VERSION := v$(VERSION_NO_V)

version:
	@echo "Version           : $(VERSION), no v: $(VERSION_NO_V)"
	@echo "Bumped version    : $(BUMPED_VERSION), no v: $(BUMPED_VERSION_NO_V)"
	@echo "Bump              : $(BUMP)"
	@echo "Commits since tag : $(COMMITS_SINCE_TAG)"
	@echo "SHA1 hash         : $(HASH)"

# End of the semver part

GOENV   :=
GOFLAGS := -ldflags \
	"\
	-X 'github.com/thomas-maurice/gowerline/gowerline-server/version.Version=$(VERSION)' \
	-X 'github.com/thomas-maurice/gowerline/gowerline-server/version.BuildHost=$(shell hostname)' \
	-X 'github.com/thomas-maurice/gowerline/gowerline-server/version.BuildTime=$(shell date)' \
	-X 'github.com/thomas-maurice/gowerline/gowerline-server/version.BuildHash=$(HASH)' \
	-X 'github.com/thomas-maurice/gowerline/gowerline-server/version.OS=$(shell go env GOOS)' \
	-X 'github.com/thomas-maurice/gowerline/gowerline-server/version.Arch=$(shell go env GOARCH)' \
	"

BINARY_SUFFIX := $(VERSION)_$(shell go env GOOS)_$(shell go env GOARCH)

clean:
	rm -rf ./bin __pycache__ gowerline.egg-info build dist

.PHONY: bump-version
bump-version:
	if [ git branch | grep \* | cut -f 2 -d \ != "master" ]; then echo "This must be ran from master"; exit 1; fi;
	echo "$(BUMPED_VERSION)" > VERSION
	git add VERSION
	git commit -m "bump version $(VERSION) -> $(BUMPED_VERSION)"
	git tag $(BUMPED_VERSION) -m "bump version $(VERSION) -> $(BUMPED_VERSION)"
	@echo "Don't forget to git push --tags, run make push_tags"

test:
	( cd gowerline-server; go test ./... -race -cover )
	for plg in $(shell ls plugins); do \
		( cd ./plugins/$${plg} ; $(GOENV) go test ./... -race -cover ) \
	done;

lint:
	if ! which golangci-lint 2>/dev/null; then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); fi;
	( cd gowerline-server; golangci-lint run )
	for plg in $(shell ls plugins); do \
		( cd ./plugins/$${plg} ; golangci-lint run ) \
	done;

push_tags:
	git push
	git push --tags

autobump: bump-version push_tags

.PHONY: upgrade_script
upgrade_script: bin
	cp -v install.sh bin/upgrade-gowerline
	chmod +x bin/upgrade-gowerline

.PHONY: install-upgrade-script
install-upgrade-script: upgrade_script
	cp -v bin/upgrade-gowerline ~/.gowerline/bin
	chmod +x ~/.gowerline/bin/upgrade-gowerline

.PHONY: upload-pypi
upload-pypi:
	if [ -d dist ]; then rm -fr dist; fi
	python3 -m build -s
	python3 -m twine upload --repository testpypi dist/*

.PHONY: bin
bin:
	if ! [ -d bin ]; then mkdir bin; fi
	if ! [ -d bin/plugins ]; then mkdir bin/plugins; fi

.PHONY: archive
archive: plugins
	cd bin && tar zcvf plugins-$(BINARY_SUFFIX).tar.gz plugins

.PHONY: start
start:
	systemctl --user enable gowerline
	systemctl --user start gowerline

.PHONY: restart-powerline
restart-powerline:
	if pgrep -f powerline-daemon; then powerline-daemon --replace; fi;

.PHONY: stop
stop:
	systemctl --user stop gowerline || true

.PHONY: restart
restart:
	systemctl --user restart gowerline

.PHONY: server
server: bin
	( cd gowerline-server ; $(GOENV) go build -o ../bin/gowerline-$(BINARY_SUFFIX) $(GOFLAGS) . )

.PHONY: plugins
plugins:
	for plg in $(shell ls plugins); do \
		( cd ./plugins/$${plg} ; $(GOENV) go build -o ../../bin/plugins/$${plg} -buildmode=plugin $(GOFLAGS) . ) \
	done;

.PHONY: run
run: install-extension install-server install-plugins
	~/.gowerline/bin/gowerline server run

.PHONY: install-extension
install-extension:
	pip3 install --editable $(shell pwd)
	if pgrep -f powerline-daemon >/dev/null; then powerline-daemon --replace; fi;


.PHONY: install-server
install-server: server stop
	echo "Installing the server"
	if ! [ -d ~/.gowerline ]; then mkdir ~/.gowerline; fi;
	if ! [ -d ~/.gowerline/plugins ]; then mkdir ~/.gowerline/plugins; fi;
	if ! [ -d ~/.gowerline/bin ]; then mkdir ~/.gowerline/bin; fi;
	cp -v bin/gowerline-$(BINARY_SUFFIX) ~/.gowerline/bin/gowerline
	make start

.PHONY: install-plugins
install-plugins: plugins
	if ! [ -d ~/.gowerline ]; then mkdir ~/.gowerline; fi;
	if ! [ -d ~/.gowerline/plugins ]; then mkdir ~/.gowerline/plugins; fi;
	cp -v bin/plugins/* ~/.gowerline/plugins

.PHONY: copy-config
copy-config:
	if ! [ -f ~/.gowerline/gowerline.yaml ]; then cp -v gowerline.yaml ~/.gowerline; fi;

.PHONY: install
install: install-extension install-server install-plugins copy-config install-upgrade-script restart

.PHONY: install-full
install-full: install-systemd install

.PHONY: install-systemd
install-systemd:
	if ! [ -d ~/.config/systemd/user ]; then mkdir -p ~/.config/systemd/user; fi
	cp -v systemd/gowerline.service  ~/.config/systemd/user
	systemctl --user daemon-reload
	systemctl --user enable gowerline

.PHONY: uninstall
uninstall:
	systemctl --user stop gowerline
	systemctl --user disable gowerline
	rm -r ~/.gowerline
	rm -f ~/.config/systemd/user/gowerline.service
	systemctl --user daemon-reload
