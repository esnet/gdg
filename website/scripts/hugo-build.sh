#!/usr/bin/env bash
# Builds the site in two passes - one per doc version ("current"/"0.9.6") -
# and merges the output into a single public/ dir.
#
# Why: Hugo has a known upstream race condition (gohugoio/hugo#13492) that
# can intermittently panic ("deferred execution with id ... not found")
# when building multiple languages together in one process, on sites that
# use templates.Defer-based shortcodes (Relearn's notice/expand). Each
# version alone builds cleanly every time, so we build each one as its own
# Hugo "segment" (see [segments] in config/_default/hugo.toml) instead of
# building both languages in a single invocation.
set -uo pipefail

HUGO_ARGS=("$@")
SEGMENTS=(current v096)

rm -rf public resources .hugo_build.lock

for seg in "${SEGMENTS[@]}"; do
  attempt=1
  max_attempts=3
  while true; do
    hugo --renderSegments "$seg" "${HUGO_ARGS[@]}"
    status=$?
    if [ "$status" -eq 0 ]; then
      break
    fi
    if [ "$attempt" -ge "$max_attempts" ]; then
      echo "hugo build (segment: $seg) failed after $max_attempts attempts (exit $status)" >&2
      exit "$status"
    fi
    echo "hugo build (segment: $seg) hit the known Hugo race-condition panic (exit $status) - retrying ($attempt/$max_attempts)..." >&2
    attempt=$((attempt + 1))
    sleep 1
  done
done

echo "Build complete: public/ contains both versions."
