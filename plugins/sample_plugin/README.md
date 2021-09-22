# Sample plugin, fill in the blanks :)

A general description of what your thing does goes there

## How to configure the plugin

Configure it in in `~/.gowerline/YOUR_PLUGIN_NAME.yaml` like so:
```yaml
# some yaml file
```

:warning: :warning: Please put a sample `YOUR_PLUGIN_NAME.yaml` file in this directory, it will get coppied to the user's install in case the plugin has never been installed.

:warning: Your configuration file *should very much* be named `YOUR_PLUGIN_NAME.yaml`.

## Example powerline configuration
This is where you show how a user can use your segment
```json
{
    "function": "gowerline.gowerline.gwl",
    "priority": 60,
    "args": {
        "function": "someplugin",
        "somename": "somevalue"
    }
}
```

## Highlight groups used
Every highlight group should default to `information:regular` when no other is available.

| Highlight group | Description |
| --- | --- |
| `gwl:some_group` | Some description |
| `gwl:some_other_group` | Some other description |

## Miscellaneous
Some other info, like theme suggestions and such :)