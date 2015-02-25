holo - minimal configuration management
=======================================

Today's DevOps is all about configuration management tools like Chef and
Puppet, humongous software suites that intend to manage your system
configuration. Wait, wasn't that what we invented package management for? Why
slap another 100k lines of Ruby code on the existing package management
solution?

holo is a radically different configuration management tool that relies as much
as possible on package management for the whole system setup and maintenance
process. This is achieved by using metapackages to define personal package
selections for all systems or for certain types of systems.

What the package management doesn't do
--------------------------------------

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

TODO
====

* include some example metapackages in the repo
