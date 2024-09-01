#!/bin/bash
cd ../sso/tests/unit_tests && go test *.go -v
cd ../../../loyalty/tests/unit_tests && go test *.go -v