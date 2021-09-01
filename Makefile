all: bin server plugins archive

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
	cd bin && tar zcvf plugins-$(shell git tag| head -n 1)-$(shell go env GOOS)-$(shell go env GOARCH).tar.gz plugins

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
	go build -o bin/gowerline-server-$(shell git tag| head -n 1)-$(shell go env GOOS)-$(shell go env GOARCH) ./gowerline-server

.PHONY: plugins
plugins:
	for plg in $(shell ls plugins); do go build -o bin/plugins/$${plg} -buildmode=plugin ./plugins/$${plg}; done

.PHONY: run
run: install-extension install-server install-plugins
	~/.gowerline/gowerline-server

.PHONY: install-extension
install-extension:
	pip3 install --editable $(shell pwd)

.PHONY: install-server
install-server: server stop
	if ! [ -d ~/.gowerline ]; then mkdir ~/.gowerline; fi;
	if ! [ -d ~/.gowerline/plugins ]; then mkdir ~/.gowerline/plugins; fi;
	cp -v bin/gowerline-server-$(shell git tag| head -n 1)-$(shell go env GOOS)-$(shell go env GOARCH) ~/.gowerline
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
	if ! [ -d ~/.gowerline ]; then mkdir ~/.gowerline; fi;
	if ! [ -d ~/.gowerline/plugins ]; then mkdir ~/.gowerline/plugins; fi;
	cp -v bin/gowerline-server ~/.gowerline
	cp -v bin/plugins/* ~/.gowerline/plugins
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
