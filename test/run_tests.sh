#!/bin/sh

set -e
bash "$(dirname $0)/holo/run_tests.sh"
bash "$(dirname $0)/holo-build/run_tests.sh"
