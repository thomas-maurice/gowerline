# Bash - run commands on a regular basis

This plugin allows you to run bash commands on a schedule and store the result

## How to configure the plugin

You need to add the following config structure in the `plugin[].config` field of the `~/.gowerline/gowerline.yaml` file

```yaml
---
commands:
  date:
    cmd: "date"
    interval: 30
    highlightGroup: "information:regular"
  kubeContext:
    cmd: "kubectl config get-contexts --no-headers | grep '*' | awk '{ print $3 }'"
    interval: 5
    highlightGroup: "information:regular"
```

## Example powerline configuration
This is where you show how a user can use your segment
```json
{
    "function": "gowerline.gowerline.gwl",
    "priority": 60,
    "args": {
        "function": "bash",
        "cmd": "kubeContext"
    }
}
```

## Highlight groups used
Any highlight group you put in the config