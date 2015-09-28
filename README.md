# Holo - minimalistic config management

[![Build Status](https://travis-ci.org/holocm/holo.svg?branch=master)](https://travis-ci.org/holocm/holo)

Today's DevOps is all about configuration management tools like Chef and
Puppet, humongous software suites that intend to manage your system
configuration. Their sophisticated domain model allows you to document and
manage the configuration of thousands of systems at once.

And I'm just sitting here, wanting a slice of the cake for my handful of
private Linux systems. I certainly don't want to bother with all that
complexity in order to achieve a defined system state.

Defined system state... Wasn't that what we invented package management for?
Why slap another 100k lines of Ruby code on the existing package management
solution for my simple use-case?

holo is a radically simple configuration management tool that relies as much as
possible on package management for the whole system setup and maintenance
process. This is achieved by using metapackages to define personal package
selections for all systems or for certain types of systems.

## Installation

It is recommended to install to Holo as a package.
The [website](http://holocm.org) lists distributions that have a Holo
package available.

Holo depends on the following other packages:

* [Go](https://golang.org) is needed to compile Holo.
* [Perl](https://perl.org) is used for the unit tests.
* [shadow](https://pkg-shadow.alioth.debian.org/) is used to create and modify
  user accounts and groups, and is only needed at runtime.

All dependencies are available as packages for any major Linux distribution.
Once you're all set, the build is done with

```
git submodule update --init --recursive
make
make check
sudo make install
```

## Documentation

User documentation is now available at [holocm.org](http://holocm.org).
