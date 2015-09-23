#!/bin/bash

# find the directory containing the test cases
TESTS_DIR="$(readlink -f "$(dirname $0)")"

run_testcase() {
    local TEST_NAME=$1
    echo ">> Running testcase $TEST_NAME..."

    # determine testcase location
    local TESTCASE_DIR="$TESTS_DIR/$TEST_NAME"
    if [ ! -d "$TESTCASE_DIR" ]; then
        echo "Cannot run $TEST_NAME: testcase not found" >&2
        return 1
    fi
    # set cwd!
    cd "$TESTCASE_DIR"

    # clean the testcase directory
    git clean -qdXf .

    # setup environment for holo run
    cp -R source/ target/
    mkdir -p target/usr/share/holo/repo
    mkdir -p target/var/lib/holo/backup
    export HOLO_CHROOT_DIR="./target/"
    export HOLO_MOCK=1
    export HOLO_CURRENT_DISTRIBUTION=unittest
    # the test may define a custom environment, mostly for $HOLO_CURRENT_DISTRIBUTION
    [ -f env.sh ] && source ./env.sh

    # when backup files exist, make sure their mtimes are in sync with the
    # targets (or else `holo apply` will refuse to work on them)
    cd "$TESTCASE_DIR/target/var/lib/holo/backup/"
    find -type f -o -type l | while read FILE; do
        REPOFILE="$TESTCASE_DIR/target/usr/share/holo/repo/$FILE"
        [ -f "$REPOFILE" ] && touch -r "$FILE" "$REPOFILE"
    done
    cd "$TESTCASE_DIR/"

    # run holo
    ../../build/holo scan          2>&1 | ../strip-ansi-colors.sh > scan-output
    ../../build/holo apply         2>&1 | ../strip-ansi-colors.sh > apply-output
    # if "holo apply" that certain operations will only be performed with --force, do so now
    grep -q -- --force apply-output && \
    ../../build/holo apply --force 2>&1 | ../strip-ansi-colors.sh > apply-force-output

    # dump the contents of the target directory into a single file for better diff'ing
    # (NOTE: I concede that this is slightly messy.)
    cd "$TESTCASE_DIR/target/"
    find \( -type f -printf '>> %p = regular\n' -exec cat {} \; \) -o \( -type l -printf '>> %p = symlink\n' -exec readlink {} \; \) \
        | perl -E 'local $/; print for sort split /^(?=>>)/m, <>' > "$TESTCASE_DIR/tree"
    cd "$TESTCASE_DIR/"

    local EXIT_CODE=0

    # use diff to check the actual run with our expectations
    for FILE in tree scan-output apply-output apply-force-output; do
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
    echo ">> All tests completed successfully."
else
    echo "!! Some or all tests failed. Please check the output above for more information."
fi
exit $TEST_EXIT_CODE
