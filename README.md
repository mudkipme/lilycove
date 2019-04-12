lilycove
========

A MediaWiki reverse-proxy cache purger for [52Poké Wiki](https://wiki.52poke.com/). 

## Requirements

* MediaWiki
* Apache Kafka

## Configuration

Place a `config.toml` in current directory. The `uris` should match the 

```toml
[http]
port = 8080

[queue]
type = "kafka"
group_id = "lilycove"
broker = "localhost:9092"
topic = "lilycove"
rate_limit = 5
rate_interval = 10000

[purge]
expiry = 86400000

[[purge.entries]]
host = "localhost"
method = "PURGE"
variants = ["zh", "zh-hans", "zh-hant"]
uris = [
    "http://localhost#url##variants#",
    "http://localhost#url##variants#mobile"
]

[[purge.entries]]
host = "media_proxy"
method = "PURGE"
uris = [
    "http://media_proxy#url#",
]

```

Also add this server to `$wgSquidServers` to allow MediaWiki produce `PURGE` requests.

If you use nginx as a reverse proxy server, you'll need `libnginx-mod-http-cache-purge` module installed.