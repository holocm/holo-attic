This test checks the detection and handling of backup files, especially
orphaned backup files.

* `/etc/still-existing.conf` has an intact config file and repo file. The usual
  case, just for comparison.
* `/etc/targetfile-deleted.conf` has no config file and no repo files. So we
  assume that the application package and all holograms using that application
  have been uninstalled, and delete the backup file (which came from the now
  uninstalled application package).
* `/etc/targetfile-deleted-with-pacsave.conf` is the same, but we simulate that
  the package manager saved the config file with a `.pacsave` suffix while
  uninstalling the application package. This file should be cleaned up, too.
* `/etc/repofile-deleted.conf` has no repo files, but the config file is still
  present. We assume that the application package is still installed, but the
  installed target is no longer valid after removing the holograms that act on
  it, so we restore the backup file to the target location.

In this testcase, a `backup` directory is under source control, representing
the state of the backup directory before the `holo apply` run that we're
looking at. (Since we're testing the effects of package removal, we assume that
`holo apply` has been run before.)

The backup for `/etc/still-existing.conf` is also under source control, to
check that the backup file is usually not touched by subsequent `holo apply`
runs.
