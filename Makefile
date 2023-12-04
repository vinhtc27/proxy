single-origin:
	go run origin/tcp/tcp.go 127.0.0.1:8080

multi-origin:
	go run origin/tcp/main.go 127.0.0.1:8081 & \
	go run origin/tcp/main.go 127.0.0.1:8082 & \
	go run origin/tcp/main.go 127.0.0.1:8083 & \
	go run origin/tcp/main.go 127.0.0.1:8084 & \
	go run origin/tcp/main.go 127.0.0.1:8085

single-proxy:
	go run proxy/main.go --servers=http://127.0.0.1:8080

multi-proxy:
	go run proxy/main.go --servers=http://127.0.0.1:8081,http://127.0.0.1:8082,http://127.0.0.1:8083,http://127.0.0.1:8084,http://127.0.0.1:8085

bench:
	wrk -c 128 -t 16 -d 32 http://127.0.0.1:9090
port:
	lsof -i -P -n | grep LISTEN

websocket-origin:
	go run origin/websocket/main.go

websocket-proxy:
	go run proxy/main.go --servers=http://127.0.0.1:8080 --websocket=true