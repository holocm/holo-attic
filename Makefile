default: build/holo build/holo-build build/holo-users-groups build/man/holo.8 build/man/holo-build.8 build/man/holo-plugin-interface.7
.PHONY: install check test

build/holo: src/holo/main.go src/holo/*/*.go src/shared/*.go
	go build -o $@ $<
build/holo-build: src/holo-build/main.go src/holo-build/*/*.go src/shared/*.go
	go build -o $@ $<
build/holo-users-groups: src/holo-users-groups/*.go
	go build -o $@ ./src/holo-users-groups

# manpages are generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
build/man/%: doc/man/%.pod src/shared/version.go
	pod2man --name="$(shell echo $* | cut -d. -f1)" --section=$(shell echo $* | cut -d. -f2) --center="Configuration Management" \
		--release="Holo $(shell grep 'var version =' src/shared/version.go | cut -d'"' -f2)" \
		$< $@

# this utility is needed only for testing
build/dump-package: src/dump-package/main.go src/dump-package/*/*.go
	go build -o $@ $<

test: check # just a synonym
check: default build/dump-package
	@bash test/run_tests.sh

install: default src/holo/holorc src/holo-run-scripts util/completions/holo.bash util/completions/holo-build.bash util/completions/holo.zsh util/completions/holo-build.zsh
	install -d -m 0755 "$(DESTDIR)/var/lib/holo"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/base"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/provisioned"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/run-scripts"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/repo"
	install -D -m 0644 src/holo/holorc         "$(DESTDIR)/etc/holo/holorc"
	install -D -m 0755 build/holo              "$(DESTDIR)/usr/bin/holo"
	install -D -m 0755 build/holo-build        "$(DESTDIR)/usr/bin/holo-build"
	install -D -m 0755 src/holo-run-scripts    "$(DESTDIR)/usr/lib/holo/holo-run-scripts"
	install -D -m 0755 build/holo-users-groups "$(DESTDIR)/usr/lib/holo/holo-users-groups"
	install -D -m 0644 build/man/holo.8        "$(DESTDIR)/usr/share/man/man8/holo.8"
	install -D -m 0644 build/man/holo-build.8  "$(DESTDIR)/usr/share/man/man8/holo-build.8"
	install -D -m 0644 build/man/holo-plugin-interface.7 "$(DESTDIR)/usr/share/man/man7/holo-plugin-interface.7"
	install -D -m 0644 util/completions/holo.bash        "$(DESTDIR)/usr/share/bash-completion/completions/holo"
	install -D -m 0644 util/completions/holo-build.bash  "$(DESTDIR)/usr/share/bash-completion/completions/holo-build"
	install -D -m 0644 util/completions/holo.zsh         "$(DESTDIR)/usr/share/zsh/site-functions/_holo"
	install -D -m 0644 util/completions/holo-build.zsh   "$(DESTDIR)/usr/share/zsh/site-functions/_holo-build"

# the website is generated with pod2html (also from Perl) and a HTML template;
# everything is mushed together using a small helper program
build/holo-makewebsite: doc/makewebsite.go
	go build -o $@ $<

.PHONY: prepare-website-repo
prepare-website-repo:
	@[ -d website/.git ] || git clone https://github.com/holocm/holocm.github.io website/

# the manpages are also used for man-*.html, but the manpage-style all-caps
# headings need to be converted to title case
doc/website-man-%.pod: doc/man/%.pod
	perl -pE 's/^=head1\s+([A-Z ]+)/=head1 \u\L\1/' $< > $@
.SECONDARY: doc/website-man-holo.pod doc/website-man-holo-build.pod

website/%.html: doc/website-%.pod doc/template.html build/holo-makewebsite prepare-website-repo
	build/holo-makewebsite $*

.PHONY: website
website: prepare-website-repo $(patsubst doc/man/%.pod,website/man-%.html,$(wildcard doc/man/*.pod)) $(patsubst doc/website-%.pod,website/%.html,$(wildcard doc/website-*.pod))
