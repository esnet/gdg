#!/usr/bin/env bash
# Wraps `hugo server` with an auto-restart loop.
#
# Why: Hugo has a known upstream race condition (gohugoio/hugo#13492) that
# can intermittently panic while serving multiple languages together on
# sites using templates.Defer-based shortcodes (Relearn's notice/expand).
# It's a bug in Hugo itself, not this site, and restarting the server
# recovers in about a second - so instead of leaving a dead server until
# someone notices and reruns it by hand, we just do that automatically.
#
# We also clear Hugo's on-disk resource cache (resources/, .hugo_build.lock)
# before every (re)start. If a panic happens mid-build, it can leave that
# cache half-written; reusing it on the next attempt can result in a server
# that comes back up but is silently missing/stale on some pages, showing
# up as real 404s rather than a visible crash. Clearing it forces every
# restart to do a full, consistent rebuild.
set -uo pipefail

cleanup() {
  exit 0
}
trap cleanup INT TERM

while true; do
  rm -rf resources .hugo_build.lock
  hugo server "$@"
  status=$?
  if [ "$status" -eq 0 ] || [ "$status" -eq 130 ]; then
    exit 0
  fi
  echo "" >&2
  echo "hugo server exited unexpectedly (status $status) - most likely the known Hugo templates.Defer race (gohugoio/hugo#13492), not a real error. Clearing the build cache and restarting..." >&2
  sleep 1
done
