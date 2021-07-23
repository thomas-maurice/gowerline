# Colourenv - Context dependent evironment variable printing

Returns an different colour segment depending on the value of an env variable and a regex

## How to configure the plugin

Configure it in in `~/.gowerline/colourenv.yaml` like so:
```yaml
---
variables:
  ENV:
    - regex: stag
      highlightGroup: "information:priority"
    - regex: devel
      highlightGroup: "information:regular"
    - regex: prod
      highlightGroup: "warning:regular"
```

And this will print the value of the variable using the specified highlight groups.

## Example powerline configuration
Then you can add a config like
```json
{
    "function": "gowerline.gowerline.gwl",
    "priority": 60,
    "args": {
        "function": "colourenv",
        "variable": "ENV"
    }
}
```

## Highlight groups used
Litterally anything that you specify. It will eventually default to `information:regular`

## Miscellaneous
Nope