This testcase checks the basic behavior with single, plain repo files and
target files of all kinds (regular or symlink).

    /etc/plain-over-plain.conf          # stock config is plain file, repo has plain file
    /etc/link-over-plain.conf           # stock config is plain file, repo has link file
    /etc/plain-over-link.conf           # stock config is link file, repo has plain file
    /etc/link-over-link.conf            # stock config is link file, repo has link file

Also, some error cases are tested:

* `/etc/no-stock-file.conf` has a repo file, but not a stock config file.

