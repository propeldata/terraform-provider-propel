.PHONY: build lint release install_macos uninstall_macos test testacc

GO_FILES=$(wildcard */*.go)
BINARY=terraform-provider-propel
VERSION=0.1.5

build: terraform-provider-propel

terraform-provider-propel: $(GO_FILES)
	echo $(GO_FILES)
	go build -o ${BINARY}

lint: $(GO_FILES)
	goimports -l -w .
	go mod tidy

release:
	goreleaser release --rm-dist --snapshot --skip-publish  --skip-sign

install_macos: $(BINARY)
	mkdir -p ~/Library/Application\ Support/io.terraform/plugins/registry.terraform.io/propeldata/propel/${VERSION}/darwin_arm64
	cp ${BINARY}  ~/Library/Application\ Support/io.terraform/plugins/registry.terraform.io/propeldata/propel/${VERSION}/darwin_arm64

install_linux_amd64:
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/propeldata/propel/${VERSION}/linux_amd64
	cp ${BINARY} ~/.terraform.d/plugins/registry.terraform.io/propeldata/propel/${VERSION}/linux_amd64

uninstall_linux_amd64:
	rm -rf ~/.terraform.d/plugins/registry.terraform.io/propeldata/propel/${VERSION}/linux_amd64

uninstall_macos:
	rm -r ~/Library/Application\ Support/io.terraform/plugins/registry.terraform.io/propeldata

test: $(GO_FILES)
	go test ./... || exit 1

testacc: $(GO_FILES)
	TF_ACC=1 go test ./... -timeout 120m
