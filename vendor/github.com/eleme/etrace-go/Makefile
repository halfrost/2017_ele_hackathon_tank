ENABLE_VENDOR=1
export GO15VENDOREXPERIMENT=${ENABLE_VENDOR}
GO_PKGS=$(shell go list ./... | grep -v '/vendor/')
dep:
	godep save -t -v ${GO_PKGS}
test:
	go test -cover -v ${GO_PKGS} 
lint:
	golint .
