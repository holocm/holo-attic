This test checks the platform integration for RPM-based distributions.

* `/etc/targetfile-with-rpmnew.conf` has a config file and repo file with an
  existing backup, and there is also a `.rpmnew` file that the package manager
  has placed next to the config file as part of an update of the application
  package. We should recognize this file and move it into the backup location.

[Reference](https://ask.fedoraproject.org/en/question/25722/what-are-rpmnew-files/)
