#!/bin/sh

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
    mkdir -p backup/
    export HOLO_TARGET_DIR="./target/"
    export HOLO_BACKUP_DIR="./backup/"
    export HOLO_REPO_DIR="./repo/"

    # run holo
    ../../build/holo scan 2>&1  | ../strip-ansi-colors.sh > scan-output
    ../../build/holo apply 2>&1 | ../strip-ansi-colors.sh > apply-output

    # dump the contents of the target/ directory into a single file for better diff'ing
    # (NOTE: I concede that this is slightly messy.)
    cd "$TESTCASE_DIR/target/"
    find \( -type f -printf '>> %p = regular\n' -exec cat {} \; \) -o \( -type l -printf '>> %p = symlink\n' -exec readlink {} \; \) \
        | perl -E 'local $/; print for sort split /^(?=>>)/m, <>' > "$TESTCASE_DIR/target-tree"
    cd "$TESTCASE_DIR/"

    local EXIT_CODE=0

    # use diff to check the actual run with our expectations
    for FILE in target-tree scan-output apply-output; do
        if diff -q expected-$FILE $FILE &>/dev/null; then true; else
            echo "!! The $FILE deviates from our expectation. Diff follows:"
            diff -u expected-$FILE $FILE 2>&1 | sed 's/^/    /'
            EXIT_CODE=1
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

if [ $TEST_EXIT_CODE == 0 ]; then
    echo ">> All tests completed successfully."
else
    echo "!! Some or all tests failed. Please check the output above for more information"
fi
exit $TEST_EXIT_CODE
