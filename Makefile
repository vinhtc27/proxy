forward:
	go run cache/*.go

reverse:
	go run origin/*.go &
	go run proxy/*.go &

port:
	lsof -i -P -n | grep LISTEN

stop:
	kill $(lsof -i tcp:8080)
	kill $(lsof -i tcp:9090)