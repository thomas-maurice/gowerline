all: bin server plugins archive

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
	rm -rf ./bin

.PHONY: bump-version
bump-version:
	git tag $(BUMPED_VERSION) -m "bump version $(VERSION) -> $(BUMPED_VERSION)"
	@echo "Don't forget to git push --tags, run make push_tags"

push_tags:
	git push
	git push --tags

autobump: bump-version push_tags

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
	systemctl --user start gowerline

.PHONY: stop
stop:
	systemctl --user stop gowerline

.PHONY: restart
restart:
	systemctl --user restart gowerline

.PHONY: server
server: bin
	$(GOENV) go build -o bin/gowerline-server-$(BINARY_SUFFIX) $(GOFLAGS) ./gowerline-server

.PHONY: plugins
plugins:
	for plg in $(shell ls plugins); do \
		$(GOENV) go build -o bin/plugins/$${plg} -buildmode=plugin $(GOFLAGS) ./plugins/$${plg}; \
	done;

.PHONY: run
run: install-extension install-server install-plugins
	~/.gowerline/gowerline-server

.PHONY: install-extension
install-extension:
	pip3 install --editable $(shell pwd)

.PHONY: install-server
install-server: server stop
	echo "Installing the server"
	if ! [ -d ~/.gowerline ]; then mkdir ~/.gowerline; fi;
	if ! [ -d ~/.gowerline/plugins ]; then mkdir ~/.gowerline/plugins; fi;
	cp -v bin/gowerline-server-$(BINARY_SUFFIX) ~/.gowerline/gowerline-server
	make start

.PHONY: install-plugins
install-plugins: plugins
	if ! [ -d ~/.gowerline ]; then mkdir ~/.gowerline; fi;
	if ! [ -d ~/.gowerline/plugins ]; then mkdir ~/.gowerline/plugins; fi;
	for plg in $(shell ls plugins); do \
		if ! [ -f ~/.gowerline/$${plg}.yaml ]; then \
			if [ -f plugins/$${plg}/$${plg}.yaml ]; then \
				cp -v plugins/$${plg}/$${plg}.yaml ~/.gowerline/$${plg}.yaml; \
			fi; \
		fi; \
	done;
	cp -v bin/plugins/* ~/.gowerline/plugins

.PHONY: install
install: install-extension install-server install-plugins
	if ! [ -f ~/.gowerline/server.yaml ]; then cp -v server.yaml ~/.gowerline; fi;
	if ! [ -d ~/.config/systemd/user ]; then mkdir ~/.config/systemd/user; fi
	cp -v systemd/gowerline.service  ~/.config/systemd/user
	pip3 install --editable $(shell pwd)

.PHONY: install-full
install-full: install install-systemd

.PHONY: install-systemd
install-systemd:
	if ! [ -d ~/.config/systemd/user ]; then mkdir ~/.config/systemd/user; fi
	cp -v systemd/gowerline.service  ~/.config/systemd/user
	systemctl --user daemon-reload
	systemctl --user restart gowerline
	systemctl --user enable gowerline

.PHONY: uninstall
uninstall:
	systemctl --user stop gowerline
	systemctl --user disable gowerline
	rm -r ~/.gowerline
	rm -f ~/.config/systemd/user/gowerline.service
