default: build/holo build/holo-build build/holo.8 build/holo-build.8
.PHONY: install check test

build/holo: src/holo/main.go src/holo/*/*.go src/shared/*.go
	go build -o $@ $<
build/holo-build: src/holo-build/main.go src/holo-build/*/*.go src/shared/*.go
	go build -o $@ $<

# the manpage is generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
build/%.8: doc/manpage-%.pod src/shared/version.go
	pod2man --name="$*" --section=8 --center="Configuration Management" \
		--release="Holo $(shell grep 'var version =' src/shared/version.go | cut -d'"' -f2)" \
		$< $@

# this utility is needed only for testing
build/dump-package: test/holo-build/dump-package.go
	go build -o $@ $<

test: check # just a synonym
check: build/holo build/holo-build build/dump-package
	@bash test/run_tests.sh

install: build/holo build/holo-build build/holo.8 build/holo-build.8 util/completion.bash util/completion.zsh
	install -d -m 0755 "$(DESTDIR)/var/lib/holo"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/base"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/provisioned"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/repo"
	install -D -m 0755 build/holo           "$(DESTDIR)/usr/bin/holo"
	install -D -m 0755 build/holo-build     "$(DESTDIR)/usr/bin/holo-build"
	install -D -m 0644 build/holo.8         "$(DESTDIR)/usr/share/man/man8/holo.8"
	install -D -m 0644 build/holo-build.8   "$(DESTDIR)/usr/share/man/man8/holo-build.8"
	install -D -m 0644 util/completion.bash "$(DESTDIR)/usr/share/bash-completion/completions/holo"
	install -D -m 0644 util/completion.zsh  "$(DESTDIR)/usr/share/zsh/site-functions/_holo"

# the website is generated with pod2html (also from Perl) and a HTML template;
# everything is mushed together using a small helper program
build/holo-makewebsite: doc/makewebsite.go
	go build -o $@ $<

.PHONY: prepare-website-repo
prepare-website-repo:
	@[ -d website/.git ] || git clone https://github.com/holocm/holocm.github.io website/

# the manpages are also used for man-*.html, but the manpage-style all-caps
# headings need to be converted to title case
doc/website-man-%.pod: doc/manpage-%.pod
	perl -pE 's/^=head1\s+([A-Z ]+)/=head1 \u\L\1/' $< > $@
.SECONDARY: doc/website-man-holo.pod doc/website-man-holo-build.pod

website/%.html: doc/website-%.pod doc/template.html build/holo-makewebsite prepare-website-repo
	build/holo-makewebsite $*

.PHONY: website
website: prepare-website-repo website/man-holo.html website/man-holo-build.html $(patsubst doc/website-%.pod,website/%.html,$(wildcard doc/website-*.pod))
