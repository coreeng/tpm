#!/usr/bin/env bash
# Assertion helpers for the Lab E2E workflow. Sourced by workflow steps.
#
# Reads the validator pod's status conditions directly via kubectl/jq -- the same
# source `tpm lab status` renders -- so assertions don't depend on parsing the
# human-readable status table.
#
# Requires SYSTEM_NS in the environment (set by the "Start lab" step).

# cond <condition-type>: print the condition's status (True/False/Unknown), or "" if absent.
cond() {
  kubectl get pods -n "$SYSTEM_NS" -l app.kubernetes.io/component=validator -o json \
    | jq -r --arg t "$1" '.items[0].status.conditions[]? | select(.type==$t) | .status'
}

# dump_conditions: print all IA* conditions for the run log.
dump_conditions() {
  echo "--- validator conditions ($SYSTEM_NS) ---"
  kubectl get pods -n "$SYSTEM_NS" -l app.kubernetes.io/component=validator -o json \
    | jq -r '.items[0].status.conditions[]? | select(.type|startswith("IA")) | "  \(.type)=\(.status) (\(.reason))"'
}

# assert_eq <condition-type> <expected-status>: fail the step if they differ.
assert_eq() {
  local got
  got="$(cond "$1")"
  if [ "$got" != "$2" ]; then
    echo "::error::expected $1=$2 but got '${got:-<absent>}'"
    dump_conditions
    return 1
  fi
  echo "ok: $1=$2"
}

# poll <condition-type> <want-status> <timeout-seconds>: wait until the condition
# reaches want-status, polling every 10s; fail on timeout.
poll() {
  local t=0
  while [ "$t" -lt "$3" ]; do
    if [ "$(cond "$1")" = "$2" ]; then
      echo "ok: $1 reached $2 after ${t}s"
      return 0
    fi
    sleep 10
    t=$((t + 10))
  done
  echo "::error::$1 did not reach $2 within $3s (last='$(cond "$1")')"
  dump_conditions
  return 1
}
