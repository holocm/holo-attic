holo - minimal configuration management
=======================================

Today's DevOps is all about configuration management tools like Chef and
Puppet, humongous software suites that intend to manage your system
configuration. Their complex domain model allows you to document and manage the
configuration of thousands of systems at once.

And I'm just sitting here, wanting a slice of the cake for my single notebook
and my single home server. I certainly don't want to bother with all that
complexity in order to achieve a defined system state.

Defined system state... Wasn't that what we invented package management for?
Why slap another 100k lines of Ruby code on the existing package management
solution for my simple use-case?

holo is a radically simple configuration management tool that relies as much as
possible on package management for the whole system setup and maintenance
process. This is achieved by using metapackages to define personal package
selections for all systems or for certain types of systems.

What the package management does not cover
------------------------------------------

Metapackages go only 90% of the way, though. The most important shortcoming is
that metapackages cannot install custom configuration files where the original
packages already installed stock configuration files.

Instead, metapackages designed for holo place their custom configuration files
under the `/holo/repo` directory, e.g. `/holo/repo/etc/foobar.conf`. The
`holo-apply` is then run by the metapackage's post-install and post-update hook
to place the custom configuration file at its designated position
(e.g. `/etc/foobar.conf`), while simultaneously retaining a copy of the stock
configuration file in `/holo/backup` (e.g. `/holo/backup/etc/foobar.conf`) for
reference.

Dependencies
============

holo is written in Go, so it compiles to static binaries without any runtime
dependencies (other than a UNIX kernel). That being said, the current
implementation depends on Arch Linux at some points (the packaging using a
PKGBUILD, and the built-in handling of `pacnew` files). The algorithm itself is
distribution-independent, though.

Installation
------------

On Arch Linux, the preferred installation method is as a package (following the
prime directive of never installing programs in `/usr` manually). To do so,

    make archpackage

and then install the resulting package, either through `pacman -U` or by putting
the package in a [private package repository](https://www.archlinux.org/pacman/repo-add.8.html).

Building holo metapackages
==========================

So system configuration is now expressed as metapackages, which I personally
refer to as **holodecks**. You can also choose any other weird name that you
like. For how to build packages, please refer to the documentation of your
distribution. What will your metapackage need to do?

1. List all the packages in the metapackage's dependency list, to ensure that
   they are installed when you install your metapackage (and removed when you
   remove the metapackage and its dependencies recursively).

2. You will also need to list `holo-tools` as a package dependency to get the
   `holo-apply` command.

3. Install any configuration files that the included software needs to function
   in whatever way you desire. If the base packages for that software install a
   sample configuration file in the same location, install your file in
   `/holo/repo` instead, e.g. `/holo/repo/etc/locale.conf` instead of
   `/etc/locale.conf`.

4. If the package format specifies post-install and post-upgrade script hooks
   (the important ones all do), use these to run `holo-apply`. For example, an
   Arch PKGBUILD would need a `.install` script containing:

```
post_install() {
    holo-apply
}
post_upgrade() {
    holo-apply
}
```

How files from /holo/repo are applied
-------------------------------------

The default strategy is to copy the repo file at e.g.
`/holo/repo/etc/foobar.conf` to its target location at `/etc/foobar.conf`
(minus the `/holo/repo` prefix), while taking a backup of the stock
configuration in `/holo/backup/etc/foobar.conf` (target location plus
`/holo/backup` prefix).

However, if the repo file carries the extra `.holoscript` extension, it will be
executed like this to produce the target configuration file:

    /holo/repo/etc/foobar.conf.holoscript < /holo/backup/etc/foobar.conf > /etc/foobar.conf

So the `.holoscript` program takes the stock configuration file on standard
input and produces the custom configuration file on standard output. This is
especially useful to modify only selected configuration values while otherwise
retaining the default configuration:

    $ cat /holo/repo/etc/pacman.conf.holoscript
    #!/bin/sh
    # enable the "Color" option of pacman
    sed 's/^#\s*Color$/Color/'

TODO
====

* include some example metapackages in the repo
* support for other distributions (I rely on external patches here, I'm on Arch
  only)
