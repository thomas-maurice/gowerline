# Network plugin

Gives you info about network stuff, such as your public IP address and such.

## How to configure the plugin

Configure it in in `~/.gowerline/gowerline.yaml` in the `plugin[].config` field like so:
```yaml
# ipService needs to just return your ip address with nothing
# else. Otherwise, it would fail.
# Known to work are:
#  * https://checkip.amazonaws.com/
#  * https://ifconfig.me/ip
ipService: https://checkip.amazonaws.com/
```

:warning: :warning: Please put a sample `YOUR_PLUGIN_NAME.yaml` file in this directory, it will get coppied to the user's install in case the plugin has never been installed.

:warning: Your configuration file *should very much* be named `YOUR_PLUGIN_NAME.yaml`.

## Example powerline configuration
This is how you display your public IP
```json
{
    "function": "gowerline.gowerline.gwl",
    "priority": 60,
    "args": {
        "function": "public_ip",
    }
}
```

This is how you get the address of an interface
```json
{
    "function": "gowerline.gowerline.gwl",
    "priority": 60,
    "args": {
        "function": "interface_ip",
        "interface": "eth0"
    }
}
```

## Highlight groups used
Every highlight group should default to `information:regular` when no other is available.

| Highlight group | Description |
| --- | --- |
| `gwl:public_ip` | Your public IP address |
| `gwl:interface_ip` | The ip of a given interface |

## Miscellaneous
None yet