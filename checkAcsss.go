package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/infrasonar/go-libagent"
)

func CheckAcsss(_ *libagent.Check) (map[string][]map[string]any, error) {
	state := map[string][]map[string]any{}
	acsssStatusExec := os.Getenv("ACSSS_STATUS_EXEC")
	if acsssStatusExec == "" {
		acsssStatusExec = "acsss status"
	}
	out, err := exec.Command("bash", "-c", acsssStatusExec).Output()

	status := map[string]any{}

	if err == nil {
		lines := reSplit.Split(string(out), -1)

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) != 2 {
				continue
			}

			switch fields[0] {
			case "weblogic:":
				status["weblogic"] = fields[1]
			case "surrogate:":
				status["surrogate"] = fields[1]
			case "rmi-registry:":
				status["rmi-registry"] = fields[1]
			case "acsdb:":
				status["acsdb"] = fields[1]
			case "smce:":
				status["smce"] = fields[1]
			case "stmf:":
				status["stmf"] = fields[1]
			case "acsls:":
				status["acsls"] = fields[1]
			}
		}
	} else {
		log.Printf("Failed to execute: bash -c \"%v\" (%v)\n", acsssStatusExec, err)
	}

	// Add the status type
	state["status"] = []map[string]any{status}

	// Add the agent version
	state["agent"] = []map[string]any{{
		"name":    "acsls",
		"version": version,
	}}

	// Print debug dump
	b, _ := json.MarshalIndent(state, "", "    ")
	log.Fatal(string(b))

	return state, nil
}
