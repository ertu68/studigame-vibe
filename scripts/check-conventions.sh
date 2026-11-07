#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <branch-name> <base-sha>..<head-sha>" >&2
  exit 2
fi

branch_name=$1
commit_range=$2

if [[ ! $branch_name =~ ^(main|(feat|fix|docs|ci|build|test|refactor|chore|perf|style)/[a-z0-9]+([.-][a-z0-9]+)*)$ ]]; then
  echo "Invalid branch name: '$branch_name'." >&2
  echo "Use main or <type>/<lowercase-description>, for example docs/update-guide." >&2
  exit 1
fi

if ! subjects=$(git log --format=%s "$commit_range"); then
  echo "Could not read commits in range '$commit_range'." >&2
  exit 1
fi

if [[ -z $subjects ]]; then
  echo "No commits found in range '$commit_range'." >&2
  exit 1
fi

failed=0
while IFS= read -r subject; do
  if [[ ! $subject =~ ^(feat|fix|docs|ci|build|test|refactor|chore|perf|style)(\([a-z0-9][a-z0-9._/-]*\))?!?:\ .+ ]]; then
    echo "Invalid commit subject: '$subject'." >&2
    echo "Use <type>(optional-scope): description, for example docs: update guide." >&2
    failed=1
  fi
done <<< "$subjects"

exit "$failed"
