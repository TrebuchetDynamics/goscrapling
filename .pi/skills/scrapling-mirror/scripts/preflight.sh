#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$root"

echo "## goscrapling mirror preflight"
echo "root: $root"
echo

echo "## worktree"
git status --short --branch --untracked-files=all

echo
if [ -d references/Scrapling/.git ]; then
  echo "## upstream Scrapling"
  if git -C references/Scrapling fetch --tags origin >/tmp/goscrapling-scrapling-fetch.log 2>&1; then
    echo "fetch: ok"
  else
    echo "fetch: failed (continuing with local checkout)"
    sed -n '1,40p' /tmp/goscrapling-scrapling-fetch.log
  fi
  echo "head: $(git -C references/Scrapling rev-parse HEAD)"
  if git -C references/Scrapling rev-parse origin/main >/dev/null 2>&1; then
    echo "origin/main: $(git -C references/Scrapling rev-parse origin/main)"
    echo "delta HEAD..origin/main:"
    git -C references/Scrapling diff --name-status HEAD..origin/main | sed -n '1,120p'
  else
    echo "origin/main: unavailable"
  fi
else
  echo "## upstream Scrapling"
  echo "missing references/Scrapling checkout"
fi

echo
if [ -f codemap.md ]; then
  echo "codemap: present"
else
  echo "codemap: absent"
fi

echo
printf 'map-validate: '
go run ./cmd/progress map-validate
printf 'progress validate: '
go run ./cmd/progress validate
printf 'progress json: '
jq empty docs/content/building-goscrapling/architecture_plan/progress.json && echo ok
printf 'app-map json: '
jq empty docs/content/building-goscrapling/architecture_plan/upstream-app-map.json && echo ok
printf 'diff check: '
git diff --check && echo ok

echo
if [ -f docs/content/building-goscrapling/builder-loop/next-slices.md ]; then
  echo "## next slices"
  sed -n '1,80p' docs/content/building-goscrapling/builder-loop/next-slices.md
fi
