ENABLE_VENDOR=1
export GO15VENDOREXPERIMENT=${ENABLE_VENDOR}
GO_PKGS=$(shell go list ./... | grep -v '/vendor/')
deps:
	godep save -t -v ${GO_PKGS}
test:
	for pkg in $(GO_PKGS); do \
		coverFile=$${GOPATH}/src/$${pkg}/cover.out;\
		go test  -cover -v -coverprofile=$${coverFile} $$pkg ; \
	done

.PHONY: deps
