#!/bin/bash

set -eu

echo ' ===== Lint files'
revive -formatter friendly

echo
echo ' ===== Install commands'
go install ./cmd/init-migrator
go install ./cmd/migrator

echo
echo ' ===== Initialize migrator'
init-migrator -user root -address localhost

echo
echo ' ===== Run migrations'
migrator -user root -address localhost -directory testdata
