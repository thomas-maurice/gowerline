Returns an different colour segment depending on the value of an env variable

Configure in `~/.gowerline/colourenv.yaml` like so:
```yaml
variables:
  ENV:
    prod: "information:priority"
    devel: "information:regular"
    stag: "warning:regular"
```

And this will print the value of the variable using the specified highlight groups