#!/bin/bash

# http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euo pipefail

dir=${0%/*}
cd "$dir"

echo $dir

mkdir tmp -p

PID_TEMPL=""
PID_AIR=""
PID_TW=""

cleanup() {
    echo -e "\nStopping tailwind, air, and templ "
    pkill -9 air
    pkill -9 templ
    pkill -9 tailwindcss
}

log_with_timestamp() {
    local cmd=$1
    local logfile=$2
    local program=$3

    bash -c "$cmd" 2>&1 | while IFS= read -r line; do
        echo "$program = $(date +"%T"): $line" | tee -a $logfile
    done &
    echo -e "$cmd @ PID: $!"
}

log_with_timestamp "templ generate -path src/components --watch" "tmp/templ.log" "templ      "
log_with_timestamp "./node_modules/.bin/tailwindcss -i static/in.css -o static/style.css -w" "tmp/tailwind.log" "tailwindcss"
log_with_timestamp "air" "tmp/air.log" "air        "  


trap 'cleanup' INT # Trap interrupt signal and call cleanup function
( trap exit SIGINT ; read -r -d '' _ </dev/tty ) # wait for Ctrl-C
exit
