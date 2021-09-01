# gowerline

Because Python is hard and I've always wanted to write my segments in Go.

![Example](https://github.com/thomas-maurice/gowerline/blob/master/_assets/demo.gif)

## What is this ?
This is a deamon that generates and returns Powerline segments as [described in the docs](https://powerline.readthedocs.io/en/master/develop/segments.html).
This project has 2 parts:
* The daemon that runs your go code and exposes an HTTP API
* The powerline compatible code that glues it with the `powerline-daemon`

## How does it work ?
This is a pluggable segment generating system. Essentially a very simple powerline "segment" in the sense
of the python class, will call a Go server that will be in charge of generating the actual segments. It allows you
to generate data that would be too long to generate if it had to be called every time you pop a shell,
for instance do API calls to your favourite stock ticker or for example check the validity of an auth token
every other minute.

## Available plugins

Every plugin has a `README.md` file at the root of their directory detailing what they do and how they work

| Plugin name | Plugin description |
|-------|-------|
| [Bash](https://github.com/thomas-maurice/gowerline/blob/master/plugins/bash/README.md) | Renders segments that are the result of bash commands ran on a schedule |
| [Finnhub](https://github.com/thomas-maurice/gowerline/blob/master/plugins/finnhub/README.md) | Displays financial infos about a stock ticker (or many!) that you are interested in |
| [Vault](https://github.com/thomas-maurice/gowerline/blob/master/plugins/vault/README.md) | Gives you information about your current Hashicorp Vault token (display name, validity TTL & co) |
| [Colourenv](https://github.com/thomas-maurice/gowerline/blob/master/plugins/colourenv/README.md) | Renders environment variables in your terminal with different colourschemes depending on values (useful to not wreck production by mistake) |

## How does it work (on my system) ?
You have two parts to it:
* The powerline extension, that bridges between poweline and the Go server
* The Go server that will bridge between the python extension and the plugins.

Essentially, everytime you open a prompt, Powerline will call the various extensions of the server to
fetch segments to render, it's as simple as this.

## How do I install it ?
```bash
$ make install # Will install the python extension and the go binary
$ make install-systemd # Will install the userland systemd service to start the server
$ make install-full # will do both
# this will remove it
$ make uninstall
```

You might also need to `pip install -r requirements.txt`

## How do I use it ?
### Add the segment to powerline
This will use the `time` plugin. In your powerline theme, add the following:
```json
{
    "function": "gowerline.gowerline.gwl",
    "priority": 10,
    "args": {
        "function": "time"
    }
}
```

Every plugin exposes one or more `function` that you have to reference in your powerline config. This will effectively
be passed down to the Go code, as long as every other variable you add in this JSON.

The Gowerline config itself lives in `~/.gowerline/server.yaml`
```yaml
---
listen:
    # port: 6666
    unix: ~/.gowerline/server.sock
plugins:
- time
- finnhub
- vault
- colourenv
```

## How do I extend it ?
Go have a look at the [example plugin](https://github.com/thomas-maurice/gowerline/blob/master/_sample_plugin/README.md). It should
be easy to understand. Feel free to copy it in the `plugins/` directory and fill in the blanks.

The `Makefile` is designed so that if you run `make plugins` your new source will be picked up and compiled to `bin/plugins/<plugin>`

The plugins are complied as Go plugins (essentialy `.so` libraries) that are loaded by the main daemon.