#!/bin/bash -e
set -o pipefail

# set -x (bash debug) if log level is trace
# https://github.com/osixia/docker-light-baseimage/blob/stable/image/tool/log-helper
log-helper level eq trace && set -x

log-helper info "Running as [$0][$*]"
log-helper info "Running command '/init.sh start'"
sh /init.sh start

exit $?
