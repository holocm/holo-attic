This testcase checks the basic behavior with single, plain repo files and
target files of all kinds (regular or symlink).

    /etc/plain-over-plain.conf          # stock config is plain file, repo has plain file
    /etc/link-over-plain.conf           # stock config is plain file, repo has link file
    /etc/plain-over-link.conf           # stock config is link file, repo has plain file
    /etc/link-over-link.conf            # stock config is link file, repo has link file

Also, some error cases are tested:

* `/etc/stock-file-missing.conf` has a repo file, but not a stock config file.
* `/etc/stock-file-is-directory.conf` has a repo file, but the target is a
  directory (and thus not a manageable file).
