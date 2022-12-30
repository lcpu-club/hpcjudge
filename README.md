# hpcjudge

HPCGame Judger.

## Development

Please install `just` (command utility runner), `go` (compiler), `nsq` (message queue), `minio` (object storage) in your system `PATH`.

- hpc-discovery: depends on nothing
- hpc-authd: depends on nothing
- hpc-bridge: depends on `minio` and an HPC cluster
- hpc-judge: depends on `minio`, `nsqlookupd`, `nsqd`, `hpc-discovery`, `hpc-bridge`
- hpc-agent: depends on `hpc-bridge`, `hpc-judge` and an HPC cluster

All these components should be available in the same network.

Run `just build` to build all the components
