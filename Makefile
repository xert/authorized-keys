.PHONY:run imports lint clean

all: clean authorized-keys

authorized-keys:
	go build -ldflags "-w -s -X main.builddate=`date -u +%Y%m%d.%H%M%S` -X main.buildversion=`git describe`" authorized-keys.go

clean:
	rm -f authorized-keys

run: authorized-keys
	./authorized-keys user_x --force-server=a.server --test

imports:
	goimports -w  authorized-keys.go

lint:
	golint authorized-keys.go

freebsd:
	GOOS=freebsd GOARCH=amd64 go build -ldflags "-w -s -X=main.builddate `date -u +%Y%m%d.%H%M%S` -X=main.buildversion `git describe`" authorized-keys.go

mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-w -s -X=main.builddate `date -u +%Y%m%d.%H%M%S` -X=main.buildversion `git describe`" authorized-keys.go
