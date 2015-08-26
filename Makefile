default: build/holo build/holo.8
.PHONY: install

build/holo: src/main.go src/holo/*.go
	go build -o $@ $<

# the manpage is generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
build/holo.8: holo.pod src/main.go
	pod2man --name="holo" --section=8 --center="Configuration Management" \
		--release="Holo $(shell grep 'var version string' src/main.go | cut -d'"' -f2)" \
		$< $@

install: build/holo src/holo-apply build/holo.8
	install -d -m 0755 "$(DESTDIR)/holo/backup"
	install -d -m 0755 "$(DESTDIR)/holo/repo"
	install -D -m 0755 build/holo     "$(DESTDIR)/usr/bin/holo"
	install -D -m 0755 src/holo-apply "$(DESTDIR)/usr/bin/holo-apply"
	install -D -m 0755 build/holo.8   "$(DESTDIR)/usr/share/man/man8/holo.8"
