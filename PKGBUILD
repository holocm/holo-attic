pkgname='holo-tools'
pkgver=0.2.2
pkgrel=1
pkgdesc='holo package management toolchain'
arch=('any')
url=''
license=('GPLv2')
makedepends=('go')
source=('holo-apply')

package() {
    cd "${pkgdir}"
    install -d -m 0755 holo/backup
    install -d -m 0755 holo/repo
    install -D -m 0755 "${srcdir}/holo-apply" usr/bin/holo-apply
}
