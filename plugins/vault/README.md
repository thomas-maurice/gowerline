This plugins displays vault related info to your powerline bars

It reads the `VAULT_ADDR` from the env, and the `VAULT_TOKEN`
either from the env or from `~/.vault-token` file if present

Then you can add the config like
```json
{
    "function": "gowerline.gowerline.gwl",
    "priority": 60,
    "args": {
        "function": "vault",
        "template": "{{ .DisplayName }}: {{ .RenderedExpiry }}",
        "expired_theme": true
    }
}
```

You should also add this to your themes config:
```json
"gwl:vault": {
    "fg": "green",
    "bg": "gray0",
    "attrs": [
        "bold"
    ]
},
"gwl:vault_expired": {
    "fg": "red",
    "bg": "gray10",
    "attrs": [
        "bold"
    ]
},
```