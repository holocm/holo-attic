default: holobinary holo.8
.PHONY: install

# This binary will be installed as "holo", but we have to use a different name
# because there is a directory called "holo".
holobinary: main.go holo/*.go
	go build -o $@ $<

# the manpage is generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
holo.8: holo.pod main.go
	pod2man --name="holo" --section=8 --center="Configuration Management" \
		--release="Holo $(shell grep 'var version string' main.go | cut -d'"' -f2)" \
		$< $@

install: holobinary holo-apply holo.8
	install -d -m 0755 "$(DESTDIR)/holo/backup"
	install -d -m 0755 "$(DESTDIR)/holo/repo"
	install -D -m 0755 holobinary "$(DESTDIR)/usr/bin/holo"
	install -D -m 0755 holo-apply "$(DESTDIR)/usr/bin/holo-apply"
	install -D -m 0755 holo.8     "$(DESTDIR)/usr/share/man/man8/holo.8"
