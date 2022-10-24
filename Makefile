TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=registry.terraform.io
NAMESPACE=propeldata
NAME=propel
BINARY=terraform-provider-${NAME}
VERSION=0.0.3
OS_ARCH=darwin_arm64

default: install

build:
	go build -o ${BINARY}

release:
	goreleaser release --rm-dist --snapshot --skip-publish  --skip-sign

install_macos: build
	mkdir -p ~/Library/Application\ Support/io.terraform/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY}  ~/Library/Application\ Support/io.terraform/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

uninstall_macos:
	rm -r ~/Library/Application\ Support/io.terraform/plugins/registry.terraform.io/propeldata

test:
	go test $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m
