# Vault - Displays information about your current token

This plugins displays [Hashicorp Vault](https://vaultproject.io) related info to your powerline bars.

It reads the `VAULT_ADDR` from the env, and the `VAULT_TOKEN`
either from the env or from `~/.vault-token` file if present. Then on regular basis it will perform the equivalent of a `vault token lookup self` call to the configured vault server to extract a whole lot of information, such as the display name of the token, the issuance time, the expiry and so on and so forth. The output is somewhat templatable using the standard Go templating syntax, and the following struct fields:
```go
type VaultState struct {
	Accessor       string // Vault accessor, basically auth backend id
	CreationTime   int64 // Token creation time
	DisplayName    string // Display name of your token
	EntityID       string // Your Vault entity id
	RenderedExpiry string // Rendered expiry time, like `expired` or `69h42m00s`
	CreationTTL    int64 // Token creation time
	ExpiryTime     int64 // Token expiration time
}
```
## Example powerline configuration
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

The `expired_theme` configuration option will render a different highlight group based on wether the token is expired or not.


## Highlight groups used
Every highlight group will default to `information:regular` when no other is available.

| Highlight group | Description |
| --- | --- |
| `gwl:vault` | Used for normal operations |
| `gwl:vault_expired` | Used when the token is expired |

## Miscellaneous
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