#!/usr/bin/env bash
# Single source of the extension-libraries pin: the OpenSysML-Extensions-Library
# repository, the commit scripts/sync-extension-libraries.sh vendors
# "OpenSysML Libraries/" from, and the ref that commit came from.
#
# Kept in one file so what CI checks and what a maintainer syncs cannot drift.
# The ref names where the pin was taken; the commit is what every fetch
# verifies, because a ref is mutable. Change them together.
EXTENSIONS_REPO="${EXTENSIONS_REPO:-https://github.com/DavidIngraham/OpenSysML-Extensions-Library.git}"
EXTENSIONS_REF="${EXTENSIONS_REF:-review/failure-analysis-20261122}"
EXTENSIONS_COMMIT="${EXTENSIONS_COMMIT:-38097bbc0198767590c58f89796a436678319924}"
