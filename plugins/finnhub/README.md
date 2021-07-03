You need to create a `~/.gowerline/finnub.yaml` file like so

```yaml
token: <finnhub token>
tickers:
- AAPL
- FB
```

Then you can add the config like
```json
{
    "function": "gowerline.gowerline.gwl",
    "priority": 10,
    "args": {
        "function": "time"
    }
}
```

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
},
```