This test checks the platform integration for dpkg-based distributions.

* `/etc/targetfile-with-dpkg-dist.conf` has a config file and repo file with an
  existing backup, and there is also a `.dpkg-dist` file that the package manager
  has placed next to the config file as part of an update of the application
  package. We should recognize this file and move it into the backup location.
* `/etc/targetfile-with-dpkg-old.conf` is the same basic situation, but instead
  of saving the new default config in `$TARGET_PATH.dpkg-dist`, dpkg decided to
  overwrite the configuration file directly, and save a backup of the previous
  configuration at `$TARGET_PATH.dpkg-old`. (It does that sometimes, apparently.)

[Reference](https://raphaelhertzog.com/2010/09/21/debian-conffile-configuration-file-managed-by-dpkg/)
