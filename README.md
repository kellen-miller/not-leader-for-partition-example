# NOT_LEADER_FOR_PARTITION Sample Code

Code is a sample to reproduce the intermittent produce issues when using franz-go to produce to redpanda from within a
WASM runtime.

## Reproduction Steps

1. Set up a cluster on Redpanda with SASL/SCRAM and SSL
2. Create a user with a SASL Mechanism of SCRAM-SHA-256
3. Create and ACL for that user and click `Allow all operations` so they have full access to the cluster
4. Override the following values in `traefik/config/dynamic.yaml` under the
   `http.middlewares.pub-middleware.plugin.pubPlugin` block:
    1. `username`: username of user
    2. `password`: password of user
    3. `brokers`: list of seed brokers (Bootstrap server URL)
5. Run `task start` or execute the following commands

```bash
traefik_plugin_dir="traefik/plugins-local/src/publisher-plugin"

mkdir -p "$traefik_plugin_dir"
go build -buildmode=c-shared -trimpath -o "$traefik_plugin_dir/plugin.wasm" "cmd/wasm/main.go"
cp ".traefik.yml" "$traefik_plugin_dir/.traefik.yml"
docker compose up -d --force-recreate 
```

Open the logs for the traefik container and wait for it to load the plugin and be available (may take a few seconds).
With the Traefik logs up, make a request to the `whoami` which was also started using:

```bash
curl whoami.docker.localhost
```

to hit the server repeatedly, you can run:

```bash
for i in `seq 1 20`; do curl whoami.docker.localhost; printf "\n"; done
```

A successful call to the `whoami` service will result in a response that looks like:

```text
Hostname: whoami
IP: 127.0.0.1
IP: ::1
IP: 172.18.0.2
RemoteAddr: 172.18.0.3:47326
GET / HTTP/1.1
Host: whoami.docker.localhost
User-Agent: curl/8.7.1
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 192.168.65.1
X-Forwarded-Host: whoami.docker.localhost
X-Forwarded-Port: 80
X-Forwarded-Proto: http
X-Forwarded-Server: traefik
X-Real-Ip: 192.168.65.1
```

but intermittently, the request will fail with a response error:

```text
failed to produce request record: NOT_LEADER_FOR_PARTITION: This server is not the leader for that topic-partition.
```

in the Traefik container logs, there will be repeated retry attempts to produce to the same broker that it states is not
the leader:

```text
2025-03-20 14:49:59 2025-03-20T18:49:59Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"wrote Produce v9","broker":"6","bytes_written":"124","write_wait":"1.390458ms","time_to_write":"342.5µs","err":"" 

2025-03-20 14:49:59 2025-03-20T18:49:59Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"read Produce v9","broker":"6","bytes_read":"58","read_wait":"711µs","time_to_read":"48.219459ms","err":"" 

2025-03-20 14:49:59 2025-03-20T18:49:59Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"retry batches processed","wanted_metadata_update":"true","triggering_metadata_update":"true","should_backoff":"false" 

2025-03-20 14:49:59 2025-03-20T18:49:59Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"produced","broker":"6","to":"request[0{retrying@-1,1(NOT_LEADER_FOR_PARTITION: This server is not the leader for that topic-partition.)}]"
```

If I try to do the same thing with Confluent Cloud, I will see similar behavior, with an event failure loop looking
like:

```text
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"received produce response","err":"NOT_LEADER_FOR_PARTITION: This server is not the leader for that topic-partition.","err_msg":"","err_records":"","recBuf":"", ent
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"recBuf","topicPartitionData":"[leader:11,leaderEpoch:4]"
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"handleRetryBatches called","why":"produce request had retry batches","updateMeta":"true","canFail":"true"
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"retry batches processed","wanted_metadata_update":"true","triggering_metadata_update":"false","should_backoff":"false"
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"produced","broker":"11","to":"130451.v1.observability.wasi.log_http_request_event[0{move:11:4@-1,1}]"

2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"wrote Produce v11","broker":"11","bytes_written":"606","write_wait":"184.575µs","time_to_write":"132.277µs","err":""
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"read Produce v11","broker":"11","bytes_read":"201","read_wait":"247.375µs","time_to_read":"1.064695ms","err":""
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"received produce response","err":"NOT_LEADER_FOR_PARTITION: This server is not the leader for that topic-partition.","err_msg":"","err_records":"","recBuf":""
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"recBuf","topicPartitionData":"[leader:11,leaderEpoch:4]"
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"handleRetryBatches called","why":"produce request had retry batches","updateMeta":"true","canFail":"true"
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"retry batches processed","wanted_metadata_update":"true","triggering_metadata_update":"false","should_backoff":"false"
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"produced","broker":"11","to":"130451.v1.observability.wasi.log_http_request_event[0{move:11:4@-1,1}]"

2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"wrote Produce v11","broker":"11","bytes_written":"606","write_wait":"217.528µs","time_to_write":"179.774µs","err":""
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"read Produce v11","broker":"11","bytes_read":"201","read_wait":"226.954µs","time_to_read":"1.116624ms","err":""
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"received produce response","err":"NOT_LEADER_FOR_PARTITION: This server is not the leader for that topic-partition.","err_msg":"","err_records":"","recBuf":""
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"recBuf","topicPartitionData":"[leader:11,leaderEpoch:4]"
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"handleRetryBatches called","why":"produce request had retry batches","updateMeta":"true","canFail":"true"
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"retry batches processed","wanted_metadata_update":"true","triggering_metadata_update":"false","should_backoff":"false"
2025-03-20T18:54:43Z DBG github.com/traefik/traefik/v3/pkg/logs/wasm.go:31 > 
msg:"produced","broker":"11","to":"130451.v1.observability.wasi.log_http_request_event[0{move:11:4@-1,1}]"
```