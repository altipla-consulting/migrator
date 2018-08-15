#!/bin/bash

set -eu

echo ' ===== Lint files'
revive -formatter friendly

echo
echo ' ===== Install commands'
actools go install ./cmd/init-migrator
actools go install ./cmd/migrator

echo
echo ' ===== Remove old database'
actools rm database

echo
echo ' ===== Start new database and wait to it'
actools start database
bash -c "until actools mysql -h database -u dev-user -pdev-password -e ';' 2> /dev/null ; do sleep 1; done"

echo
echo ' ===== Initialize migrator'
actools run go init-migrator -user root -password dev-root -address database

echo
echo ' ===== Run migrations'
actools run go migrator -user root -password dev-root -address database -directory testdata

echo
echo ' ===== Tear down test database'
actools stop database
