#!/usr/bin/env python3
"""Run a command quietly and print its elapsed wall time in milliseconds."""

import subprocess
import sys
import time


if len(sys.argv) < 2:
    raise SystemExit("usage: time-command.py COMMAND [ARG ...]")

started = time.perf_counter_ns()
completed = subprocess.run(
    sys.argv[1:],
    stdout=subprocess.DEVNULL,
    stderr=subprocess.DEVNULL,
    check=False,
)
elapsed_ms = round((time.perf_counter_ns() - started) / 1_000_000)

if completed.returncode != 0:
    raise SystemExit(completed.returncode)

print(elapsed_ms)
