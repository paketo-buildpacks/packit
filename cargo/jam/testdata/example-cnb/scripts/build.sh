#!/bin/bash
readonly PROGDIR="$(cd "$(dirname "${0}")" && pwd)"

echo "hello" > "$PROGDIR/../generated-file"
