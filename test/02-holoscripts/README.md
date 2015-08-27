This testcase checks how single `holoscript` repo files are applied to
manageable files (both regular files and symlinks). It ensures that:

1. Symlink buffers are correctly converted into content buffers before applying
   a `holoscript` to them, and the result is always a regular file (not a symlink).
2. A `holoscript` can also be a symlink to an executable, rather than the
   executable itself.

    /etc/plain-through-plain.conf       # stock config is plain file, repo has plain script
    /etc/plain-through-link.conf        # stock config is plain file, repo has symlink to script
    /etc/link-through-plain.conf        # stock config is symlink, repo has plain script
    /etc/link-through-link.conf         # stock config is symlink, repo has symlink to script
