all: binary

build/stackdriver-webhook-bridge:
	mkdir -p build && CGO_ENABLED=0 go build -o build/stackdriver-webhook-bridge .

binary: build/stackdriver-webhook-bridge

image: clean
	docker build -t sysdiglabs/stackdriver-webhook-bridge -f Dockerfile .

clean:
	rm -rf build/

test:
	go test -count=1 `go list ./... | grep -v pkg`

