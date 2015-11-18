#!/bin/sh
echo "Running provisioning script 02-failing.sh"
echo "This is output on stdout"
echo "This is output on stderr" >&2
echo "Done with 02-failing.sh, exiting with code 1"
exit 1
