.PHONY:run imports lint clean

all: clean authorized-keys

authorized-keys:
	go build -ldflags "-w -s -X main.builddate `date -u +%Y%m%d.%H%M%S`" authorized-keys.go

clean:
	rm -f authorized-keys

run: authorized-keys
	./authorized-keys user_x --force-server=a.server --test

imports:
	goimports -w  authorized-keys.go

lint:
	golint authorized-keys.go


