PROJECT="k8s-webshell"
BINARY="k8s-webshell"
VERSION=1.3
BUILD=`data +%F%T%z`

PACKAGES=`go list ./... |grep -v /vendor/`
VETPACKAGES=`go list ./... |grep -v /vendor/ |grep -v /examples/`
GOFILES=`find . -name "*.go" -type -f -not -path "./vendor/*"`

default:
	@cd src && go build -o ../bin/${BINARY} webshell/main.go


build-linux:
	@cd src \
	&& GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -gcflags \
	all=-trimpath=${GOPATH} -asmflags all=-trimpath=${GOPATH} \
	-ldflags '-w -s' -o ../bin/${BINARY} webshell/main.go

list:
	@echo ${PACKAGES}
	@echo ${VETPACKAGES}
	@echo ${GOFILES}
fmt:
	@gofmt -s -w ${GOFILES}

fmt-check:
	@diff=$$(gofmt -s -d $(GOFILES))
	if [-n "$$diff"]; then \
		 echo "Please run 'make fmt' and commit the result:"; \
		 echo "$${diff}";\
		 exit 1;\
	fi;

install:
	@govendor sync -v

test:
	@go test -cpu=1,2,4 -v -tags   integration ./...

docker:
	@docker build -t k8s-webshell:v1.3 .

clean:
	@if [ -f ${BINARY} ] ; then rm  ${BINARY} ; fi



.PHONY: default fmt fmt-check install test vet docker clean