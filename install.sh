#!/bin/bash

# this script will install gowerline on your system

set -euo pipefail

TEMPDIR=$(mktemp -d)

echo " - Creating directories"
for directory in ~/.gowerline ~/.gowerline/bin ~/.gowerline/plugins ~/.config/systemd/user; do
    mkdir -p "${directory}"
done;

echo " - Retrieving latest release number"
release=$(curl -s https://api.github.com/repos/thomas-maurice/gowerline/releases|jq -r '.[] | "\(.created_at) \(.id)"' | sort | cut -d\  -f 2 | tail -n1)
releaseInfo=$(curl -s https://api.github.com/repos/thomas-maurice/gowerline/releases/${release} | jq -c .)
tagName=$(echo "${releaseInfo}"| jq -r .name)

echo " - Downloading the last release tarball"
wget -O "${TEMPDIR}/release.tgz" https://api.github.com/repos/thomas-maurice/gowerline/tarball/v0.0.5 > /dev/null 2>&1
(cd "${TEMPDIR}" ; tar zxf release.tgz)
releaseDirName=$(ls "${TEMPDIR}" | grep thomas-maurice)

echo " - Will install gowerline ${tagName}"
echo " - Stopping gowerline if it is running"
systemctl stop --user gowerline || true

echo " - Installing the python extension"
pip install -U gowerline > /dev/null 2>&1

echo " - Restarting powerline if it is running"
if pgrep -f powerline-daemon >/dev/null; then powerline-daemon --replace; fi;

echo " - Installing gowerline binary"
wget -O ~/.gowerline/bin/gowerline "https://github.com/thomas-maurice/gowerline/releases/download/${tagName}/gowerline-${tagName}_linux_amd64" > /dev/null 2>&1
chmod +x ~/.gowerline/bin/gowerline

echo " - Installing plugins"
wget -O ~/.gowerline/plugins.tgz "https://github.com/thomas-maurice/gowerline/releases/download/${tagName}/plugins-${tagName}_linux_amd64.tar.gz" > /dev/null 2>&1
(cd ~/.gowerline/ ; tar zxf plugins.tgz)

echo " - Installing systemd unit file"
cp "${TEMPDIR}/${releaseDirName}/systemd/gowerline.service" ~/.config/systemd/user/gowerline.service

echo " - Installing upgrade script"
cp "${TEMPDIR}/${releaseDirName}/install.sh" ~/.gowerline/bin/upgrade-gowerline
chmod +x ~/.gowerline/bin/upgrade-gowerline

echo " - Installing config file if not present"
if ! [ -f ~/.gowerline/gowerline.yaml ]; then cp -v "${TEMPDIR}/${releaseDirName}/gowerline.yaml" ~/.gowerline/gowerline.yaml; fi;

echo " - Cleaning up"
rm -r "${TEMPDIR}"
rm ~/.gowerline/plugins.tgz

echo " - Refreshing systemd and restart gowerline"
systemctl --user daemon-reload
systemctl enable --user gowerline
systemctl start --user gowerline

sleep 5

~/.gowerline/bin/gowerline version
~/.gowerline/bin/gowerline plugin list

cat <<EOF

Remember to add ~/.gowerline/bin to your path so you can use the cli !

 $ export PATH=\${PATH}:\${HOME}/.gowerline/bin

EOF
