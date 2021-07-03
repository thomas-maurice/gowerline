all: bin server plugins

.PHONY: bin
bin:
	if ! [ -d bin ]; then mkdir bin; fi
	if ! [ -d bin/plugins ]; then mkdir bin/plugins; fi

.PHONY: server
server: bin
	go build -o bin/gowerline-server ./gowerline-server

.PHONY: plugins
plugins:
	for plg in $(shell ls plugins); do go build -o bin/plugins/$${plg} -buildmode=plugin ./plugins/$${plg}; done

.PHONY: run
run: server plugins
	if ! [ -d ~/.gowerline ]; then mkdir ~/.gowerline; fi;
	if ! [ -d ~/.gowerline/plugins ]; then mkdir ~/.gowerline/plugins; fi;
	cp -v bin/gowerline-server ~/.gowerline
	cp -v bin/plugins/* ~/.gowerline/plugins
	cp -v systemd/gowerline.service  ~/.config/systemd/user
	pip3 install --editable $(shell pwd)
	~/.gowerline/gowerline-server

.PHONY: install
install: bin server plugins
	if ! [ -d ~/.gowerline ]; then mkdir ~/.gowerline; fi;
	if ! [ -d ~/.gowerline/plugins ]; then mkdir ~/.gowerline/plugins; fi;
	cp -v bin/gowerline-server ~/.gowerline
	cp -v bin/plugins/* ~/.gowerline/plugins
	if ! [ -f ~/.gowerline/server.yaml ]; then cp -v server.yaml ~/.gowerline; fi;
	cp -v systemd/gowerline.service  ~/.config/systemd/user
	pip3 install --editable $(shell pwd)

.PHONY: install-full
install-full: install install-systemd

.PHONY: install-systemd
install-systemd:
	systemctl --user daemon-reload
	systemctl --user restart gowerline
	systemctl --user enable gowerline

.PHONY: uninstall
uninstall:
	systemctl --user stop gowerline
	systemctl --user disable gowerline
	rm -r ~/.gowerline/gowerline-server
	rm -r ~/.gowerline/plugins