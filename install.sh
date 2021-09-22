#!/bin/bash

# this script will install gowerline on your system

set -euo pipefail

TEMPDIR=$(mktemp -d)

echo " - Creating directories"
for directory in ~/.gowerline ~/.gowerline/bin ~/.gowerline/plugins ~/.config/systemd/user; do 
    mkdir -p "${directory}"
done;

echo " - Retrieving latest release number"
release=$(curl -s https://api.github.com/repos/thomas-maurice/gowerline/releases|jq .[].id | sort -n | tail -n1)
releaseInfo=$(curl -s https://api.github.com/repos/thomas-maurice/gowerline/releases/${release} | jq -c .)
tagName=$(echo "${releaseInfo}"| jq -r .name)

echo " - Downloading the last release tarball"
wget -O "${TEMPDIR}/release.tgz" https://api.github.com/repos/thomas-maurice/gowerline/tarball/v0.0.5 > /dev/null 2>&1
(cd "${TEMPDIR}" ; tar zxf release.tgz)
releaseDirName=$(ls "${TEMPDIR}" | grep thomas-maurice)

echo " - Will install gowerline ${tagName}"
echo " - Stopping gowerline if it is running"
systemctl stop --user gowerline || true

echo " - Installing gowerline binary"
wget -O ~/.gowerline/bin/gowerline "https://github.com/thomas-maurice/gowerline/releases/download/${tagName}/gowerline-${tagName}_linux_amd64" > /dev/null 2>&1

echo " - Installing plugins"
wget -O ~/.gowerline/plugins.tgz "https://github.com/thomas-maurice/gowerline/releases/download/${tagName}/plugins-${tagName}_linux_amd64.tar.gz" > /dev/null 2>&1
(cd ~/.gowerline/ ; tar zxf plugins.tgz)

echo " - Installing systemd unit file"
cp "${TEMPDIR}/${releaseDirName}/systemd/gowerline.service" ~/.config/systemd/user/gowerline.service

echo " - Installing config file if not present"
if ! [ -f ~/.gowerline/gowerline.yaml ]; then cp -v "${TEMPDIR}/${releaseDirName}/systemd/gowerline.service" ~/.config/systemd/user/gowerline.service; fi;

echo " - Refreshing systemd and restart gowerline"
systemctl --user daemon-reload
systemctl start --user gowerline
systemctl enable --user gowerline

echo " - Cleaning up"
rm -r "${TEMPDIR}"
rm ~/.gowerline/plugins.tgz

cat <<EOF

Remember to add ~/.gowerline/bin to your path so you can use the cli !

 $ export PATH=\${PATH}:\${HOME}/.gowerline/bin
 
EOF