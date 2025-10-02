name = terracurl
organization = devops-rob
version = 2.0.0
arch = darwin_amd64
#arch = linux_amd64

.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --examples-dir=./examples
build:
	go build -o bin/terraform-provider-$(name)_v$(version)

install: build
	mkdir -p ~/.terraform.d/plugins/local/$(organization)/$(name)/$(version)/$(arch)
	mv bin/terraform-provider-$(name)_v$(version) ~/.terraform.d/plugins/local/$(organization)/$(name)/$(version)/$(arch)/
test:
	go test ./internal/provider -v
multi_build:
	@echo ""
	@echo "Compile Provider"

	# Clear the output
	rm -rf ./bin

	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ./bin/linux_arm64/terraform-provider-$(name)_v$(version) ./main.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/linux_amd64/terraform-provider-$(name)_v$(version) ./main.go
	GOOS=darwin GOARCH=arm64 go build -o ./bin/darwin_arm64/terraform-provider-$(name)_v$(version) ./main.go
	GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin_amd64/terraform-provider-$(name)_v$(version) ./main.go
	GOOS=windows GOARCH=amd64 go build -o ./bin/windows_amd64/terraform-provider-$(name)_v$(version).exe ./main.go
	GOOS=windows GOARCH=386 go build -o ./bin/windows_386/terraform-provider-$(name)_v$(version).exe ./main.go
zip:
	pwd
	zip -j ./bin/terraform-provider-$(name)_$(version)_linux_arm64.zip ./bin/linux_arm64/terraform-provider-$(name)_v$(version)
	zip -j ./bin/terraform-provider-$(name)_$(version)_linux_amd64.zip ./bin/linux_amd64/terraform-provider-$(name)_v$(version)
	zip -j ./bin/terraform-provider-$(name)_$(version)_darwin_arm64.zip ./bin/darwin_arm64/terraform-provider-$(name)_v$(version)
	zip -j ./bin/terraform-provider-$(name)_$(version)_darwin_amd64.zip ./bin/darwin_amd64/terraform-provider-$(name)_v$(version)
	zip -j ./bin/terraform-provider-$(name)_$(version)_windows_amd64.zip ./bin/windows_amd64/terraform-provider-$(name)_v$(version).exe
	zip -j ./bin/terraform-provider-$(name)_$(version)_windows_386.zip ./bin/windows_386/terraform-provider-$(name)_v$(version).exe
	ls -lha ./bin
shasum:
	cd bin/; shasum -a 256 *.zip > terraform-provider-$(name)_$(version)_SHA256SUMS
gpg:
	gpg --detach-sign ./bin/terraform-provider-$(name)_$(version)_SHA256SUMS
release_package: docs multi_build zip shasum gpg
