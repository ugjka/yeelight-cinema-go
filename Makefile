prefix=/usr/local
PWD := $(shell pwd)
GOPATH :=$(PWD)/deps
appname = yeelight-cinema-go

all:
	GOPATH=$(GOPATH) go get -d github.com/ugjka/$(appname)
	GOPATH=$(GOPATH) go build -v
install:
	install -Dm755 $(appname) $(prefix)/bin/$(appname)

uninstall:
	rm "$(prefix)/bin/$(appname)"

clean:
	rm -rf $(GOPATH)
	rm $(appname)