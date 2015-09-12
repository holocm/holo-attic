default: build/holo build/holo.8
.PHONY: install check test

build/holo: src/main.go src/*/*.go
	go build -o $@ $<

# the manpage is generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
build/holo.8: doc/manpage.pod src/main.go
	pod2man --name="holo" --section=8 --center="Configuration Management" \
		--release="Holo $(shell grep 'var version string' src/main.go | cut -d'"' -f2)" \
		$< $@

test: check # just a synonym
check: build/holo
	sh test/run_tests.sh

install: build/holo build/holo.8
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/backup"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/repo"
	install -D -m 0755 build/holo   "$(DESTDIR)/usr/bin/holo"
	install -D -m 0644 build/holo.8 "$(DESTDIR)/usr/share/man/man8/holo.8"

# the website is generated with pod2html (also from Perl) and a HTML template;
# everything is mushed together using a small helper program
build/holo-makewebsite: doc/makewebsite.go
	go build -o $@ $<

.PHONY: prepare-website-repo
prepare-website-repo:
	@[ -d website/.git ] || git clone https://github.com/majewsky/majewsky.github.io website/

website/%.html: doc/website-%.pod doc/template.html build/holo-makewebsite prepare-website-repo
	build/holo-makewebsite $*

.PHONY: website
website: prepare-website-repo $(patsubst doc/website-%.pod,website/%.html,$(wildcard doc/website-*.pod))
