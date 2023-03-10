# Build

build: build-discovery build-judge build-bridge build-spawn build-utility
    
mkdir-dist:
    mkdir -p dist

build-discovery: mkdir-dist
    go build -o dist/hpc-discovery github.com/lcpu-club/hpcjudge/cmd/hpc-discovery

build-judge: mkdir-dist
    go build -o dist/hpc-judge github.com/lcpu-club/hpcjudge/cmd/hpc-judge

build-bridge: mkdir-dist
    go build -o dist/hpc-bridge github.com/lcpu-club/hpcjudge/cmd/hpc-bridge

build-spawn: mkdir-dist
    go build -o dist/hpc-spawn github.com/lcpu-club/hpcjudge/cmd/hpc-spawn

build-utility: mkdir-dist
    go build -o dist/hpcgame github.com/lcpu-club/hpcjudge/cmd/hpcgame

# End Build

run-dependencies:
    go run github.com/lcpu-club/hpcjudge/cmd/dev-util multi-run "just nsqlookupd" "just minio" "just nsqd" "just nsqadmin" "just hpc-discovery"

gen-uuid-v4:
    go run github.com/lcpu-club/hpcjudge/cmd/dev-util gen-uuid-v4

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

# redis configure
redis_work_dir := "./dev-temp/redis/"
redis_port := "6379"
# end redis configure

redis:
    cd {{ redis_work_dir }}
    redis-server --port {{ redis_port }}

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

# hpc-judge configure
hpc_nsqd_address := "127.0.0.1" + nsqd_tcp
hpc_nsqlookupd_address := nsqlookupd_tcp_address
hpc_minio_address := "127.0.0.1" + minio_addr
hpc_minio_user := minio_user
hpc_minio_pass := minio_pass
hpc_nsqlookupd_topic := "hpc:judge:jobs"
hpc_nsqlookupd_channel := "judgers"
hpc_nsqd_topic := "hpc:judge:report"
hpc_judge_configure := "./configure/hpc-judge.yml"
# end hpc-judge configure

hpc-judge: build-judge
    ./dist/hpc-judge serve -c {{ hpc_judge_configure }}
