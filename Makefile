default: holobinary
.PHONY: install

# This binary will be installed as "holo", but we have to use a different name
# because there is a directory called "holo".
holobinary: main.go holo/*.go
	go build -o $@ $<

install: holobinary holo-apply
	install -d -m 0755 "$(DESTDIR)/holo/backup"
	install -d -m 0755 "$(DESTDIR)/holo/repo"
	install -D -m 0755 holobinary "$(DESTDIR)/usr/bin/holo"
	install -D -m 0755 holo-apply "$(DESTDIR)/usr/bin/holo-apply"
