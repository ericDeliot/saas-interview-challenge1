all: deps build test docker-build

build:
	go install ./...

test:
	go test -v -failfast ./...

clean:
	go clean
	rm -f $(GOPATH)/bin/workersvc
	rm -f $(GOPATH)/bin/producersvc
	rm -f ./workersvc

deps:
	go get -u github.com/adjust/rmq

docker-build:
	cp $(GOPATH)/bin/workersvc .
	docker build -t worker --build-arg server=./workersvc .
