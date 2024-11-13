package main

import "regexp"

var reSplit = regexp.MustCompile(`\r?\n`)
var reLSM = regexp.MustCompile(`^(\d+[\,\d+]*)\s+(\w+)\s+(\w+)\s+(\w+)\s+(\d+)\s+(\d+)\s+(\d+)$`)
var reCAP = regexp.MustCompile(`^(\d+[\,\d+]*)\s+([A-Za-z]+)\s+([A-Za-z]+)\s+([A-Za-z]+)\s+([A-Za-z]+)\s+(\d+)\s+([A-Za-z]+)$`)
