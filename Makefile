default: holo-apply
.PHONY: install

holo-apply: holo-apply.go holo/*.go
	go build -o $@ $<

install: holo-apply
	install -d -m 0755 "$(DESTDIR)/holo/backup"
	install -d -m 0755 "$(DESTDIR)/holo/repo"
	install -D -m 0755 holo-apply "$(DESTDIR)/usr/bin/holo-apply"
