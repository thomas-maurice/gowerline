# Finnhub - Stock tickers

This plugin enables you tu get live update to your terminal from stock
tickers you are interested in. The refresh is more or less every minute,
you can have as many tickers as you like as long as it does not get rate
limited by Finnhub.io.

## How to configure the plugin
You need to create a `~/.gowerline/finnub.yaml` file like so

```yaml
token: <finnhub token>
tickers:
- AAPL
- FB
```

## Example powerline configuration
Then you can add a config like
```json
{
    "function": "gowerline.gowerline.gwl",
    "priority": 60,
    "args": {
        "function": "ticker",
        "ticker": "CFLT"
    }
}
```

## Highlight groups used
Every highlight group will default to `information:regular` when no other is available.

| Highlight group | Description |
| --- | --- |
| `gwl:ticker_up` | Will be used when the ticker goes up compared to the previous close |
| `gwl:ticker_down` | Will be used when the ticker goes down compared to the previous close |
| `gwl:ticker` | Is a more generic catchall |

## Miscellaneous

You should also add this to your themes config:
```json
"gwl:ticker_up": {
    "fg": "green",
    "bg": "gray0",
    "attrs": [
        "bold"
    ]
},
"gwl:ticker_down": {
    "fg": "red",
    "bg": "gray10",
    "attrs": [
        "bold"
    ]
},
"gwl:ticker": {
    "fg": "yellow",
    "bg": "gray0",
    "attrs": [
        "bold"
    ]
}
```