$(GOPATH)/bin/kauth: $(wildcard cmd/*.go) $(wildcard *.go)
	go build -o $(GOPATH)/bin/kauth cmd/*.go