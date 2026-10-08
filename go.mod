module not-leader-for-partition-example

go 1.26.0

require (
	github.com/http-wasm/http-wasm-guest-tinygo v0.4.0
	github.com/spf13/cast v1.10.0
	github.com/stealthrocket/net v0.2.1
	github.com/twmb/franz-go v1.22.1
	github.com/twmb/franz-go/pkg/kadm v1.19.0
)

require (
	github.com/klauspost/compress v1.20.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.30 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.14.0 // indirect
	golang.org/x/net v0.21.0 // indirect
)

// necessary for this library to compile using go 1.24, referencing this traefik maintainers PR: https://github.com/http-wasm/http-wasm-guest-tinygo/pull/34
replace github.com/http-wasm/http-wasm-guest-tinygo => github.com/traefik/http-wasm-guest-tinygo v0.0.0-20231031142937-e57ec90ac6a1
