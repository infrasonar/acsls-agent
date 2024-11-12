package main

import "regexp"

var reSplit = regexp.MustCompile(`\r?\n`)
var reNumber = regexp.MustCompile(`\d+$`)
var rePages = regexp.MustCompile(`(\d+)(\s*pages)?$`)
var reSize = regexp.MustCompile(`(.*)(\s|\:)([\d\.]+)\s*([TGMK])iB$`)
var reString = regexp.MustCompile(`\:\s*(\S.*)$`)
var reReplica = regexp.MustCompile(`^Replica\s(\d+)\s*\:(.*)$`)
