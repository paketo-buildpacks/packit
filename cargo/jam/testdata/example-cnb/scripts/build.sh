#!/bin/bash
readonly PROGDIR="$(cd "$(dirname "${0}")" && pwd)"

echo "hello from the pre-packaging script"

echo "hello" > "$PROGDIR/../generated-file"
