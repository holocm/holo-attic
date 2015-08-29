How to contribute
=================

Holo uses GitHub's tools (the issue tracker, pull requests and release
management) for its development. So if you want to contribute, have a look at
the open issues, fork the repo, start hacking and submit pull requests.

If you have any questions concerning the code structure or internals, ask your
question as an issue and I'll do my best to explain everything to you.

Branches
--------

Starting with version 0.3, Holo adopts a branching model within which the
`master` branch is the current stable release, and development for the next
stable release happens on the `develop` branch. Therefore, users can always
compile the `master` branch to get the latest bugfixes, without fear of
unexpected instability.

For developers, this means:

* Bugfixes go on the `master` branch. I will take care of forward-merging them
  into `develop` afterwards.
* Features go on the `develop` branch.
