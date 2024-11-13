package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/infrasonar/go-libagent"
)

func CheckLibCmd(_ *libagent.Check) (map[string][]map[string]any, error) {
	state := map[string][]map[string]any{}
	libCmdPhysicalExec := os.Getenv("LIB_CMD_PHYSICAL_EXEC")
	if libCmdPhysicalExec == "" {
		libCmdPhysicalExec = "lib_cmd display library physical all"
	}
	out, err := exec.Command("bash", "-c", libCmdPhysicalExec).Output()

	var matches []string

	lsms := []map[string]any{}
	caps := []map[string]any{}

	if err == nil {
		lines := reSplit.Split(string(out), -1)

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			matches = reLSM.FindStringSubmatch(line)
			if matches != nil {
				driveCount, err := strconv.ParseInt(matches[5], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse LSM Drive Count: %v", err)
				}
				volCount, err := strconv.ParseInt(matches[6], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse LSM Vol Count: %v", err)
				}
				freeCellCount, err := strconv.ParseInt(matches[7], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse LSM Free Cell Count: %v", err)
				}

				lsms = append(lsms, map[string]any{
					"name":          matches[1],
					"type":          matches[2],
					"status":        matches[3],
					"state":         matches[4],
					"driveCount":    driveCount,
					"volCount":      volCount,
					"freeCellCount": freeCellCount,
				})
			}

			matches = reCAP.FindStringSubmatch(line)
			if matches != nil {
				size, err := strconv.ParseInt(matches[6], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse CAP Size: %v", err)
				}

				caps = append(caps, map[string]any{
					"name":         matches[1],
					"mode":         matches[2],
					"state":        matches[3],
					"status":       matches[4],
					"condition":    matches[5],
					"size":         size,
					"availability": matches[7],
				})
			}
		}
	} else {
		log.Printf("Failed to execute: bash -c \"%v\" (%v)\n", libCmdPhysicalExec, err)
	}

	// Add the LSM type
	state["LSM"] = lsms

	// Add the CAP type
	state["CAP"] = caps

	// Print debug dump
	b, _ := json.MarshalIndent(state, "", "    ")
	log.Fatal(string(b))

	return state, nil
}
