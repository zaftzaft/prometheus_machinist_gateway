prometheus_macinist_gateway
===========================
Prometheus のデータを IIJ Machinist(https://machinist.iij.jp/) へ送信します。


# Install
1. clone this repogitory
2. `cd misc/`
3. `sh install.sh`

# Configuration
```yaml
prometheus_url: "http://127.0.0.1:9090"

machinist_token: "MACHINIST_TOKEN"

agent_configs:
  - agent_name: 'prometheus'
    query: 'up{instance="127.0.0.1"}'
    tag_includes: [instance]
    meta_includes: [job]
    tag:
      test: test
    meta:
      test: test
```

## agent_configs
### agent_name
`必須`
machinist の agent 名です

### query
`必須`
prometheus の query です

### namespace
machinist の namespace を設定できます

### tag_includes
machinist の tag に付ける prometheus metrics の label 名を設定できます

### meta_includes
machinist の meta に付ける prometheus metrics の label 名を設定できます

### tag
machinist の tag を設定できます

### meta
machinist の meta を設定できます
