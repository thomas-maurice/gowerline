from __future__ import (unicode_literals, division,
                        absolute_import, print_function)

from powerline.segments import Segment
from powerline.theme import requires_segment_info

from pathlib import Path

import logging
import requests
import traceback
import sys
import json
import yaml
import os

defaultConf = {
    "port": 6666
}

logPath = os.path.join(str(Path.home()), ".gowerline")
cfgPath = os.path.join(str(Path.home()), ".gowerline", "server.yaml")
cfg = {}

if not os.path.isdir(logPath):
    os.mkdir(logPath)


logging.basicConfig(filename=os.path.join(
    logPath, "gowerline.log"), level=logging.DEBUG)

if os.path.isfile(cfgPath):
    with open(cfgPath, "r") as dat:
        cfg = yaml.load(dat, Loader=yaml.FullLoader)
else:
    cfg = {"port": 6666}


@requires_segment_info
class Gowerline(Segment):
    divider_highlight_group = None

    def __call__(self, pl, segment_info, **kwargs):
        try:
            for k, v in segment_info['environ'].items():
                if type(v) is not str:
                    # sanitise the env before passing it in
                    del(segment_info['environ'], k)

            vim = {}
            for k in ["winnr", "bufnr", "tabnr", "mode"]:
                if k in segment_info:
                    vim[k] = segment_info[k]

            payload = {
                "env": segment_info['environ'] if 'environ' in segment_info else None,
                "args": kwargs,
                "function": kwargs["function"],
                "cwd": segment_info["getcwd"](),
                "home": segment_info["home"] if segment_info["home"] is not False else "",
                "vim": vim,
            }

            # TODO: it should be a list
            resp = requests.post(
                "http://127.0.0.1:{}/plugin".format(cfg["port"]),
                json=payload,
            )

            return resp.json()
        except Exception as exce:
            logging.error(
                "failed to run {}: {}".format(kwargs, traceback.format_exc()),

            )

            return [{
                "contents": "error rendering gwl",
                "highlight_groups": ["critical:failure"],
            }]


gwl = Gowerline()
