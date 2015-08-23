default: main
.PHONY: install

main: main.go holo/*.go
	go build -o $@ $<

install: main holo-apply
	install -d -m 0755 "$(DESTDIR)/holo/backup"
	install -d -m 0755 "$(DESTDIR)/holo/repo"
	install -D -m 0755 main       "$(DESTDIR)/usr/bin/holo"
	install -D -m 0755 holo-apply "$(DESTDIR)/usr/bin/holo-apply"
