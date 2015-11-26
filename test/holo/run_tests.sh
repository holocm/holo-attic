#!/bin/bash

# find the directory containing the test cases
TESTS_DIR="$(readlink -f "$(dirname $0)")"

run_testcase() {
    local TEST_NAME=$1
    echo ">> Running testcase holo/$TEST_NAME..."

    # determine testcase location
    local TESTCASE_DIR="$TESTS_DIR/$TEST_NAME"
    if [ ! -d "$TESTCASE_DIR" ]; then
        echo "Cannot run $TEST_NAME: testcase not found" >&2
        return 1
    fi
    # set cwd!
    cd "$TESTCASE_DIR"

    # setup chroot for holo run
    rm -rf -- target/
    cp -R source/ target/
    mkdir -p target/usr/share/holo/repo
    mkdir -p target/usr/share/holo/run-scripts
    mkdir -p target/usr/share/holo/users-groups
    mkdir -p target/var/lib/holo/base
    mkdir -p target/var/lib/holo/provisioned
    [ ! -f target/etc/holorc ] && cp ../holorc target/etc/holorc

    # fix a bug with Travis (Travis has an ancient git which incorrectly prints
    # paths relative to the nearest git root instead of the $PWD when called
    # with the "git diff --no-index -- FILE FILE" syntax, which changes the
    # diff-output of the unit tests; to fix this, we put a temporary git root
    # at the $PWD)
    git init >/dev/null

    # consistent file modes in the target/ directory (for test reproducability)
    find target/ -type f                     -exec chmod 0644 {} +
    find target/ -type f -name \*.sh         -exec chmod 0755 {} +
    find target/ -type f -name \*.holoscript -exec chmod 0755 {} +
    find target/ -type d                     -exec chmod 0755 {} +

    # setup environment for holo run
    export HOLO_CHROOT_DIR="./target/"
    export HOLO_MOCK=1
    export HOLO_CURRENT_DISTRIBUTION=unittest
    # the test may define a custom environment, mostly for $HOLO_CURRENT_DISTRIBUTION
    [ -f env.sh ] && source ./env.sh

    # run holo
    ../../../build/holo scan          2>&1 | ../../strip-ansi-colors.sh > scan-output
    ../../../build/holo diff          2>&1 | ../../strip-ansi-colors.sh > diff-output
    ../../../build/holo apply         2>&1 | ../../strip-ansi-colors.sh > apply-output
    # if "holo apply" that certain operations will only be performed with --force, do so now
    grep -q -- --force apply-output && \
    ../../../build/holo apply --force 2>&1 | ../../strip-ansi-colors.sh > apply-force-output

    # clean up the useless Git repo we created earlier to fix a Travis bug
    rm -rf -- .git

    # dump the contents of the target directory into a single file for better diff'ing
    # (NOTE: I concede that this is slightly messy.)
    cd "$TESTCASE_DIR/target/"
    find \( -type f -printf '>> %p = regular\n' -exec cat {} \; \) -o \( -type l -printf '>> %p = symlink\n' -exec readlink {} \; \) \
        | perl -E 'local $/; print for sort split /^(?=>>)/m, <>' > "$TESTCASE_DIR/tree"
    cd "$TESTCASE_DIR/"

    local EXIT_CODE=0

    # use diff to check the actual run with our expectations
    for FILE in tree scan-output diff-output apply-output apply-force-output; do
        if [ -f $FILE ]; then
            if diff -q expected-$FILE $FILE >/dev/null; then true; else
                echo "!! The $FILE deviates from our expectation. Diff follows:"
                diff -u expected-$FILE $FILE 2>&1 | sed 's/^/    /'
                EXIT_CODE=1
            fi
        fi
    done

    return $EXIT_CODE
}

# this var will be set to 1 when a testcase fails
TEST_EXIT_CODE=0

# inspect arguments
if [ $# -gt 0 ]; then
    # testcase names given - run these testcases
    for TESTCASE in $@; do
        run_testcase $TESTCASE || TEST_EXIT_CODE=1
    done
else
    # no testcases given - run them all in order
    for TESTCASE in $(find "$TESTS_DIR" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | sort); do
        run_testcase $TESTCASE || TEST_EXIT_CODE=1
    done
fi

if [ $TEST_EXIT_CODE = 0 ]; then
    echo ">> All tests for \"holo\" completed successfully."
else
    echo "!! Some or all tests for \"holo\" failed. Please check the output above for more information."
fi
exit $TEST_EXIT_CODE
