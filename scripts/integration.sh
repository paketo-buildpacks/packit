#!/usr/bin/env bash

set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly BUILDPACKDIR="$(cd "${PROGDIR}/.." && pwd)"

# shellcheck source=SCRIPTDIR/.util/tools.sh
source "${PROGDIR}/.util/tools.sh"

# shellcheck source=SCRIPTDIR/.util/print.sh
source "${PROGDIR}/.util/print.sh"

# shellcheck source=SCRIPTDIR/.util/git.sh
source "${PROGDIR}/.util/git.sh"

function main() {
  while [[ "${#}" != 0 ]]; do
    case "${1}" in
      --use-token|-t)
        shift 1
        token::fetch
        ;;

      --help|-h)
        shift 1
        usage
        exit 0
        ;;

      "")
        # skip if the argument is empty
        shift 1
        ;;

      *)
        util::print::error "unknown argument \"${1}\""
    esac
  done

  if [[ ! -d "${BUILDPACKDIR}/integration" ]]; then
      util::print::warn "** WARNING  No Integration tests **"
  fi

  tests::run
}

function usage() {
  cat <<-USAGE
integration.sh [OPTIONS]

Runs the integration test suite.

OPTIONS
  --help       -h  prints the command usage
  --use-token  -t  use GIT_TOKEN from lastpass
USAGE
}

function token::fetch() {
  GIT_TOKEN="$(util::git::token::fetch)"
  export GIT_TOKEN
}

function tests::run() {
  util::print::title "Run Library Integration Tests"

  testout=$(mktemp)
  pushd "${BUILDPACKDIR}" > /dev/null
    if GOMAXPROCS="${GOMAXPROCS:-4}" go test -count=1 -timeout 0 ./integration/... -v -run Integration | tee "${testout}"; then
      util::tools::tests::checkfocus "${testout}"
      util::print::success "** GO Test Succeeded **"
    else
      util::print::error "** GO Test Failed **"
    fi
  popd > /dev/null
}

main "${@:-}"
