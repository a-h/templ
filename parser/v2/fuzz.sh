#!/bin/bash
echo Element
go test -fuzz=FuzzElement -fuzztime=120s
echo Script
go test -fuzz=FuzzScript -fuzztime=120s
