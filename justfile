# Build

build: build-discovery
    
mkdir-dist:
    mkdir -p dist

build-discovery: mkdir-dist
    go build -o dist/hpc-discovery github.com/lcpu-club/hpcjudge/cmd/hpc-discovery

# End Build

# MinIO Configure
minio_user := "hpc"
minio_pass := "hpc@devel"
minio_path := "./dev-temp/minio"
minio_addr := ":9000"
minio_console := ":9090"
# End MinIO Configure

minio:
    MINIO_ROOT_USER={{ minio_user }} MINIO_ROOT_PASSWORD={{ minio_pass }} minio server {{ minio_path }} --address={{ minio_addr }} --console-address={{ minio_console }}

# nsq configure
nsqlookupd_tcp := ":4160"
nsqlookupd_http := ":4161"
nsqlookupd_tcp_address := "127.0.0.1:4160"
nsqlookupd_http_address := "127.0.0.1:4161"
nsqd_tcp := ":4150"
nsqd_http := ":4151"
nsqd_max_msg_size := "1048576"
nsqd_mem_queue_size := "100"
nsqd_broadcast := "127.0.0.1"
nsqd_path := "./dev-temp/nsq/"
nsqadmin_http := ":4171"
# end nsq configure

nsqlookupd:
    nsqlookupd -tcp-address={{ nsqlookupd_tcp }} -http-address={{ nsqlookupd_http }}

nsqd:
    nsqd -http-address={{ nsqd_http }} -tcp-address={{ nsqd_tcp }} -data-path={{ nsqd_path }} -max-msg-size={{ nsqd_max_msg_size }} -mem-queue-size={{ nsqd_mem_queue_size }} -lookupd-tcp-address={{ nsqlookupd_tcp_address }} -broadcast-address={{ nsqd_broadcast }}

nsqadmin:
    nsqadmin -http-address={{ nsqadmin_http }} -lookupd-http-address={{ nsqlookupd_http_address }}

# hpc-discovery configure
hpc_discovery_access_key := "e10adc3949ba59abbe"
hpc_discovery_listen := ":20751"
hpc_discovery_external_address := "http://localhost:20752"
hpc_discovery_peer_timeout := "5s"
hpc_discovery_peers := "-p http://localhost:20751"
hpc_discovery_data := "./dev-temp/discovery/hpc-discovery.dat"
# end hpc-discovery configure

hpc-discovery: build-discovery
    ./dist/hpc-discovery serve -k {{ hpc_discovery_access_key }} -l {{ hpc_discovery_listen }} -e {{ hpc_discovery_external_address }} -t {{ hpc_discovery_peer_timeout }} {{ hpc_discovery_peers }}
