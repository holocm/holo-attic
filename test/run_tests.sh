#!/bin/sh

set -e
cd "$(dirname $0)"
bash ../src/holo-test holo holo/??-*
bash ./holo-build/run_tests.sh
