#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

go run $DIR/main.go removecrdvalidation ../config/crd/bases/etl.dataworkz.nl_workflows.yaml