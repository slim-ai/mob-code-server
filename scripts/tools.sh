#!/usr/bin/env bash

set -euo pipefail

die () {
    echo >&2 "$@"
    exit 1
}

#################################################################
# Pretty colors for nice messages
RESET="\\033[0m"
RED="\\033[31;1m"
GREEN="\\033[32;1m"

print()
{
  COLOR=${1:-}
  MESSAGE=${2:-}
  [ -z "${SILENT:-}" ] && printf "%b%s%b\\n" "${COLOR}" "${MESSAGE}" "${RESET}"
  return 0
}

print_unsupported_platform()
{
    >&2 print ${RED} "error: Darn it. It looks like your system is not supported for platform development!"
}

at_exit()
{
    if [ "$?" -ne 0 ]; then
        >&2 print ${RED}
        >&2 print ${RED} "Yo... it looks like something might have gone wrong during tools setup."
    fi
}
trap at_exit EXIT

if ! command -v pulumi >/dev/null; then
  case $(uname -s) in
      Linux*)
      (
        curl -fsSL https://get.pulumi.com | sh
      )
      ;;
      Darwin*)
      (
        brew install pulumi
      )
      ;;
      *)
          print_unsupported_platform
          exit 1
          ;;
  esac
fi

print ${GREEN} "tools installed..."



