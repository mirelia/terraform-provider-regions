HOSTNAME=registry.terraform.io
NAMESPACE=mirelia
NAME=regions
BINARY=terraform-provider-${NAME}
VERSION=1.0.3
OS_ARCH=linux_amd64

default: install

build:
	go build -o ${BINARY}

release:
	GOOS=darwin GOARCH=amd64 go build -o ./${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -o ./${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o ./${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o ./${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o ./${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o ./${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o ./${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}