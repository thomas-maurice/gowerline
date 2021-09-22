from __future__ import (unicode_literals, division,
                        absolute_import, print_function)

from powerline.segments import Segment
from powerline.theme import requires_segment_info

from pathlib import Path

import logging
import requests
import requests_unixsocket
import urllib.parse
import traceback
import yaml
import os
import os.path

defaultConf = {
    "listen": {
        "port": 6666,
    }
}

logPath = os.path.join(str(Path.home()), ".gowerline")
cfgPath = os.path.join(str(Path.home()), ".gowerline", "gowerline.yaml")
cfg = {}

if not os.path.isdir(logPath):
    os.mkdir(logPath)

serverURL = ""

if os.path.isfile(cfgPath):
    with open(cfgPath, "r") as dat:
        cfg = yaml.load(dat, Loader=yaml.FullLoader)
        if "listen" in cfg and "unix" in cfg["listen"]:
            requests_unixsocket.monkeypatch()
            serverURL = "http+unix://{}".format(
                urllib.parse.quote_plus(os.path.expanduser(cfg["listen"]["unix"])))
        else:
            serverURL = "http://127.0.0.1:{}".format(cfg["port"])
else:
    cfg = {"port": 6666, "debug": False}
    serverURL = "http://127.0.0.1:{}".format(cfg["port"])

if "debug" in cfg and cfg["debug"]:
    logging.basicConfig(filename=os.path.join(
        logPath, "gowerline.log"), level=logging.DEBUG)
else:
    logging.basicConfig(filename=os.path.join(
        logPath, "gowerline.log"), level=logging.INFO)


@requires_segment_info
class Gowerline(Segment):
    def __call__(self, pl, segment_info, **kwargs):
        returnedSegment = [{
            "contents": "None rendering gwl",
            "highlight_groups": ["critical:failure"],
        }]
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
                "{}/plugin".format(serverURL),
                json=payload,
            )

            respJson = resp.json()
            logging.debug("returned segment: {}".format(respJson))

            returnedSegments = []

            if respJson is None:
                return [{
                    "contents": "None rendering gwl",
                    "highlight_groups": ["critical:failure"],
                }]

            for segment in respJson:
                if not "contents" in segment or segment["contents"] == "":
                    continue

                if not "highlight_groups" in segment or segment["highlight_groups"] is None:
                    segment["highlight_groups"] = [
                        "gwl",
                        "information:regular",
                    ]
                else:
                    segment["highlight_groups"].append("gwl")
                    segment["highlight_groups"].append("information:regular")

                returnedSegments.append(segment)

            returnedSegment = returnedSegments
        except Exception as exce:
            logging.error(
                "failed to run {}: {}".format(
                    kwargs, traceback.format_exc()),

            )

            returnedSegment = [{
                "contents": "error rendering gwl",
                "highlight_groups": ["critical:failure"],
            }]
        return returnedSegment


gwl = Gowerline()
