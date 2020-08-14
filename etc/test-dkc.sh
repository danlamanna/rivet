#!/bin/bash
set -exu
set -o pipefail

readonly RIVET="$1"
shift

TESTDATA_DIR=$(mktemp -d)
TESTDATA_DIR2=$(mktemp -d)

test_checksums() {
    DIR_CHECKSUM=$(find "$1" -type d -printf '%P\n' | sort | md5sum | awk '{ print $1; }')
    FILE_CHECKSUM=$(find "$1" -type f | sort | xargs cat | md5sum | awk '{ print $1; }')
    if [ "$DIR_CHECKSUM" != "26101ec47f01a3abe91385e4ba8dcb60" ]; then
        echo "Dir checksum didn't match"
        exit 1;
    fi
    if [ "$FILE_CHECKSUM" != "2f97b47fb942df9bfb5bf5464924b2e3" ]; then
        echo "File checksum didn't match"
        exit 1;
    fi
}

# don't pass auth, this is a public collection
$RIVET sync --no-config --url data.kitware.com girder://5d77d580d35580e6dc005775 "$TESTDATA_DIR"

test_checksums "$TESTDATA_DIR"

# configure auth
expect -f ./test-dkc-configure.expect "$RIVET" "$DKC_API_KEY" > /dev/null

# make folder for testing syncing
TEST_DIR="$(date +%s)"
TEST_FOLDER=$($RIVET api-create-folder 5d77d6e1d35580e6dc0076d9 "$TEST_DIR" | xargs echo -n)

$RIVET sync "$TESTDATA_DIR" "girder://$TEST_FOLDER"

$RIVET sync "girder://$TEST_FOLDER" "$TESTDATA_DIR2"

test_checksums "$TESTDATA_DIR2"

# TODO: consider using traps for these
rm -rf "$TESTDATA_DIR" "$TESTDATA_DIR2"

