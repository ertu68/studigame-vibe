#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <branch-name> <pull-request-title>" >&2
  exit 2
fi

branch_name=$1
pull_request_title=$2

if [[ ! $branch_name =~ ^(main|(feat|fix|docs|ci|build|test|refactor|chore|perf|style)/[a-z0-9]+([.-][a-z0-9]+)*)$ ]]; then
  echo "Invalid branch name: '$branch_name'." >&2
  echo "Use main or <type>/<lowercase-description>, for example docs/update-guide." >&2
  exit 1
fi

if [[ ! $pull_request_title =~ ^(feat|fix|docs|ci|build|test|refactor|chore|perf|style)(\([a-z0-9][a-z0-9._/-]*\))?!?:\ .+ ]]; then
  echo "Invalid pull request title: '$pull_request_title'." >&2
  echo "Use <type>(optional-scope): description, for example docs: update guide." >&2
  exit 1
fi
