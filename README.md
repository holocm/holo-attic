# Holo - minimalistic config management

## Manifesto

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

## Documentation

User documentation, including installation instructions, is now available at
[holotools.org](http://holotools.org).

Tip: If you open that site and cannot find any documentation, try scrolling down. ;)
