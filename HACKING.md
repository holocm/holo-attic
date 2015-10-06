# Terminology

All the words marked in *italic* font are defined terminology for Holo

Holo can work on various *entities*. These include:

* *Users*: POSIX user accounts (as defined in `/etc/passwd`)
* *Groups*: POSIX user groups (as defined in `/etc/group`)
* *Target files*: configuration files or data files for other programs

Possible actions on entities include:

* *apply*: modify the entity so that it conforms to the configuration as defined in `/usr/share/holo`
* *diff*: display difference between the entity as it will be applied by Holo, and the current state of the entity
* *scan*: display information about the entity and where it is defined in `/usr/share/holo`

Users and groups are configured by *entity definitions* which are found at `/usr/share/holo/*.toml`. Multiple entity definitions can be *stacked*, which means that they all work on the same entity. For user and group entities, stacked entity definitions are not allowed to contradict one another.

For target files, the application algorithm involves several files:

* the target file (hereinafter referred to as `$target`)
* the *target base* (at `/var/lib/holo/base/$target`): the initial configuration file distributed by the application package
* the last provisioned version of the target file (at `/var/lib/holo/provisioned/$target`): a copy of the provisioned target file that is used by `holo diff` (there is no special term for this file)
* *repository entries* (at `/usr/share/holo/*/$target`): files provided by configuration packages that modify or replace the target base
