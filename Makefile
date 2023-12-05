single-proxy:
	go run proxy/proxy.go

single-tcp:
	go run origin/tcp/tcp.go 127.0.0.1:8080

multi-tcp:
	go run origin/tcp/tcp.go 127.0.0.1:8081 & \
	go run origin/tcp/tcp.go 127.0.0.1:8082 & \
	go run origin/tcp/tcp.go 127.0.0.1:8083 & \
	go run origin/tcp/tcp.go 127.0.0.1:8084 & \
	go run origin/tcp/tcp.go 127.0.0.1:8085

single-rest:
	go run origin/rest/rest.go 127.0.0.1:8080

single-websocket:
	go run origin/websocket/websocket.go 127.0.0.1:8080

single-grpc:
	go run origin/grpc/server/grpc.go 127.0.0.1:8080

single-quic:
	go run origin/quic/server/quic.go 127.0.0.1:8080