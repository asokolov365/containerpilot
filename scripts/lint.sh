#!/bin/bash
result=0
for pkg in $(go list -f '{{ join .Deps  "\n"}}' . | grep asokolov365/containerpilot)
do
    golint -set_exit_status "$pkg" || result=1
    go vet "$pkg" || result=1
done
exit $result
