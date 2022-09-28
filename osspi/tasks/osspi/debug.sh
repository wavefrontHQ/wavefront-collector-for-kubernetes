# shellcheck shell=bash

function is_debug_mode {
  if [ "${DEBUG+defined}" = defined ] && [ "$DEBUG" = 'on' ]; then
    return 0
  else
    return 1
  fi
}
