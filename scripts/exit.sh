#!/bin/bash

set -u

collect_pids() {
	local pattern
	local pid
	local -A seen=()

	for pattern in "$@"; do
		while read -r pid; do
			if [ -n "$pid" ]; then
				seen["$pid"]=1
			fi
		done < <(pgrep -f "$pattern" || true)
	done

	for pid in "${!seen[@]}"; do
		echo "$pid"
	done
}

stop_group() {
	local name="$1"
	shift
	local pid
	local pids=()

	while read -r pid; do
		if [ -n "$pid" ]; then
			pids+=("$pid")
		fi
	done < <(collect_pids "$@")

	if [ "${#pids[@]}" -eq 0 ]; then
		echo "$name is not running."
		return
	fi

	echo "Stopping $name: ${pids[*]}"
	kill "${pids[@]}" 2>/dev/null || true

	sleep 1

	local remaining=()
	for pid in "${pids[@]}"; do
		if kill -0 "$pid" 2>/dev/null; then
			remaining+=("$pid")
		fi
	done

	if [ "${#remaining[@]}" -gt 0 ]; then
		echo "Force stopping $name: ${remaining[*]}"
		kill -9 "${remaining[@]}" 2>/dev/null || true
	fi
}

stop_group "listen_and_send.sh" \
	"bash scripts/listen_and_send.sh" \
	"scripts/listen_and_send.sh"

stop_group "trustdsn-api" \
	"go run ./cmd/trustdsn-api" \
	"/tmp/go-build.*/exe/trustdsn-api" \
	"trustdsn-api"

stop_group "demo-web dev server" \
	"bash -lc cd demo-web && npm run dev" \
	"npm run dev" \
	"node.*vite"

if [ -x "./lotus-miner" ]; then
	echo "Stopping lotus-miner..."
	./lotus-miner stop 2>/dev/null || true
fi

if [ -x "./lotus" ]; then
	echo "Stopping lotus daemon..."
	./lotus daemon stop 2>/dev/null || true
fi

echo "Exit script finished."
