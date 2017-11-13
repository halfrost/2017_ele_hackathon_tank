NOW=$(shell date '+%Y-%m-%d_%H:%M:%S')
REV=$(shell echo "$${REV:-$$(git rev-parse HEAD)}")
LDFLAGS=-ldflags '-X github.com/eleme/nex.Build=${NOW}@${REV}'
GO_PKGS=$(shell go list ./... | egrep -v '/example|/vendor/|/tutorial')

build: nex

nex: cmd/nex
	mkdir -p bin
	go build -o 'bin/nex' ${LDFLAGS} ./cmd/nex/

install:
	go install ${LDFLAGS} ./cmd/nex/

test:
	go test -v ${GO_PKGS}
	for pkg in $(GO_PKGS); do \
		coverFile=$${pkg#github.com\/eleme\/nex}/cover.out ; \
		coverFile=./$${coverFile#\/} ; \
		go test -race -cover -v -coverprofile=$${coverFile} $$pkg ; \
	done

dep:
	godep get -t ${GO_PKGS}
	godep save -t ${GO_PKGS}

git-hooks:
	ln -sf `pwd`/tools/git-hooks/* .git/hooks/


clean:
	find . -name "cover.out" -delete

.PHONY: build nex test dep git-hooks
