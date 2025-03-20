module not-leader-for-partition-example

go 1.24

require (
	github.com/http-wasm/http-wasm-guest-tinygo v0.4.0
	github.com/spf13/cast v1.7.1
	github.com/stealthrocket/net v0.2.1
	github.com/twmb/franz-go v1.18.1
	github.com/twmb/franz-go/pkg/kadm v1.15.0
)

require (
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.9.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
)

// necessary for this library to compile using go 1.24, referencing this traefik maintainers PR: https://github.com/http-wasm/http-wasm-guest-tinygo/pull/34
replace github.com/http-wasm/http-wasm-guest-tinygo => github.com/traefik/http-wasm-guest-tinygo v0.0.0-20240913140402-af96219ffea5
