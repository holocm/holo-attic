default: holo-apply

holo-apply: holo-apply.go holo/*.go
	go build -o $@ $<

archpackage: holo-apply
	makepkg -s -f --skipchecksums
