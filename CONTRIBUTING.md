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
`stable` branch is the current stable release, and development for the next
stable release happens on the `master` branch. Therefore, users can always
compile the `stable` branch to get the latest bugfixes, without fear of
unexpected instability.

Bugfixes should be developed on the `stable` branch, since forward-merging to
the development branch is always easier than cherry-picking back into the
stable branch.

Documentation
-------------

Documentation is written in POD (Perl's documentation format), since that
format has converters to manpage and HTML readily available on all
distributions (through the pod2man and pod2html executables included with
Perl). Also I know a lot of Perl and thus was already familiar with POD.

The manpage lives at `doc/manpage.pod` and is built with:

    make build/holo.8

The website contents are at `doc/website-*.pod` and are compiled with:

    make website

This will clone the website repository into the `website` subdirectory of this
repo and place the generated HTML files in there at the right places. When you
change the website contents by editing the POD files, you only need to submit
a pull request to this repo (where the doc lives); I will take care of
publishing the changes to the website after merging your pull request here.

Of course, when you change something else about the website, e.g. the
stylesheets or the images, you need to submit a pull request to the website
repo.
