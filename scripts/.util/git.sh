#!/usr/bin/env bash

set -eu
set -o pipefail

# shellcheck source=SCRIPTDIR/print.sh
source "$(dirname "${BASH_SOURCE[0]}")/print.sh"

function util::git::token::fetch() {
  if [[ -z "${GIT_TOKEN:-""}" ]]; then
    util::print::title "Fetching GIT_TOKEN"

    GIT_TOKEN="$(
      lpass show Shared-CF\ Buildpacks/concourse-private.yml \
        | grep buildpacks-github-token \
        | cut -d ' ' -f 2
    )"
  fi

  printf "%s" "${GIT_TOKEN}"
}
