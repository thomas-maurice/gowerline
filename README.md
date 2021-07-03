# gowerline

Because Python is hard and I've always wanted to write my segments in Go.

![Example](https://github.com/thomas-maurice/gowerline/blob/master/_assets/screenshot.png)

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
port: 6666
plugins:
- time
- finnhub
- vault
- colourenv
```

## How do I extend it ?
Go have a look at the [time plugin](https://github.com/thomas-maurice/gowerline/blob/master/plugins/time/main.go). It should
be easy to understand. You essentially have to implement one function, `Init` that builds and returns a `Plugin` object.

The plugins are complied as Go plugins (essentialy `.so` libraries) that are loaded by the main daemon.
