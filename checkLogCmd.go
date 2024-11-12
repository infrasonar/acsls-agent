package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/infrasonar/go-libagent"
)


func getInt64(line string) (int64, error) {
	matches := reNumber.FindStringSubmatch(line)
	if len(matches) != 1 {
		return 0, errors.New("no match for int")
	}
	return strconv.ParseInt(matches[0], 10, 64)
}

func getSize(line string) (string, int64, error) {
	matches := reSize.FindStringSubmatch(line)
	if len(matches) != 5 {
		return line, 0, errors.New("no match for size")
	}
	line = matches[1]
	f, err := strconv.ParseFloat(matches[3], 64)
	if err != nil {
		return line, 0, err
	}

	switch matches[4] {
	case "T":
		f *= 1024
		fallthrough
	case "G":
		f *= 1024
		fallthrough
	case "M":
		f *= 1024
		fallthrough
	case "K":
		f *= 1024
	}

	return strings.TrimSpace(line), int64(f), err
}

func getPages(line string) (int64, error) {
	line, _, _ = getSize(line) // strip bytes
	matches := rePages.FindStringSubmatch(line)
	if len(matches) != 3 {
		return 0, errors.New("no match for pages")
	}
	return strconv.ParseInt(matches[1], 10, 64)
}

func getBool(line string) (bool, error) {
	if strings.HasSuffix(line, "Yes") || strings.HasSuffix(line, "On") {
		return true, nil
	}
	if strings.HasSuffix(line, "No") || strings.HasSuffix(line, "Off") {
		return false, nil
	}
	return false, errors.New("failed to read boolean Yes/No")
}

func getString(line string) (string, error) {
	matches := reString.FindStringSubmatch(line)
	if len(matches) != 2 {
		return "", errors.New("no match for string")
	}
	return matches[1], nil
}

type params struct {
	params   map[string]any
	replicas []map[string]any
}

func readFilesystem(filesystem string) (*params, error) {
	// out, err := exec.Command("bash", "-c", "cat output.example.txt").Output()
	out, err := exec.Command("bash", "-c", fmt.Sprintf("mmparam %s", filesystem)).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute `mmparam` (%s)", err)
	}

	lines := reSplit.Split(string(out), -1)

	item := map[string]any{
		"name": filesystem,
	}

	var pageSize int64 = 0
	var numReplicas int64 = 0

	for _, line := range lines {
		// page_size
		if strings.HasPrefix(line, "Page size:") {
			_, i, err := getSize(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `page_size` (%s)", err)
			}
			item["page_size"] = i
			pageSize = i
		}

		// replicas
		if strings.HasPrefix(line, "Replicas:") {
			i, err := getInt64(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `replicas` (%s)", err)
			}
			item["replicas"] = i
			numReplicas = i
		}
	}

	if pageSize == 0 {
		return nil, errors.New("missing required page size")
	}

	replicas := make([]map[string]any, numReplicas)
	for i := range numReplicas {
		replicas[i] = map[string]any{
			"name":       fmt.Sprintf("%s-replica-%d", filesystem, i),
			"replica":    fmt.Sprintf("Replica %d", i),
			"filesystem": filesystem,
		}
	}

	var replica *map[string]any = nil
	var present_pages int64 = -1
	var max_number_of_pages int64 = -1

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// keep_in_cache
		if strings.HasPrefix(line, "Keep in cache:") {
			i, err := getInt64(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `keep_in_cache` (%s)", err)
			}
			item["keep_in_cache"] = i
		}

		// archived_since_mount
		if strings.HasPrefix(line, "Archived since mount:") {
			_, i, err := getSize(line)
			if err == nil {
				item["archived_since_mount"] = i
			}
			continue
		}

		// replicated_since_mount
		if strings.HasPrefix(line, "Replicated since mount:") {
			_, i, err := getSize(line)
			if err == nil {
				item["replicated_since_mount"] = i
			}
			continue
		}

		// files_in_cache
		if strings.HasPrefix(line, "Files in cache:") {
			i, err := getInt64(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `files_in_cache` (%s)", err)
			}
			item["files_in_cache"] = i
			continue
		}

		// directories
		if strings.HasPrefix(line, "Directories:") {
			i, err := getInt64(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `directories` (%s)", err)
			}
			item["directories"] = i
			continue
		}

		// streams
		if strings.HasPrefix(line, "Streams:") {
			i, err := getInt64(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `streams` (%s)", err)
			}
			item["streams"] = i
			continue
		}

		// number_of_delayed_events
		if strings.HasPrefix(line, "Number of delayed events:") {
			i, err := getInt64(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `number_of_delayed_events` (%s)", err)
			}
			item["number_of_delayed_events"] = i
			continue
		}

		// read_write_access
		if strings.HasPrefix(line, "Read/write access:") {
			s, err := getString(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `read_write_access` (%s)", err)
			}
			item["read_write_access"] = s
			continue
		}

		// archiving
		if strings.HasPrefix(line, "Archiving:") {
			s, err := getString(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `archiving` (%s)", err)
			}
			item["archiving"] = s
			continue
		}

		//
		// Replica metrics
		//
		rMatches := reReplica.FindStringSubmatch(line)
		if len(rMatches) == 3 {
			i, err := strconv.Atoi(rMatches[1])
			if err != nil {
				return nil, errors.New("failed to read replica number")
			}
			if i < 0 || i >= len(replicas) {
				return nil, errors.New("replica out of range")
			}
			replica = &replicas[i]

			rest := strings.Split(strings.TrimSpace(rMatches[2]), ", ")
			if len(rest) >= 1 {
				key := rest[0]
				(*replica)["key"] = key
				fields := strings.Split(key, "-")
				if len(fields) == 3 {
					(*replica)["location"] = fields[0]
					(*replica)["share"] = fields[1]
					(*replica)["local_intergral_volume"] = fields[2]
				} else if len(fields) == 2 {
					(*replica)["location"] = fields[0]
					(*replica)["local_intergral_volume"] = fields[1]
				}
			}
			online := false
			inSync := false
			isRead := false
			for _, f := range rest {
				if f == "online" {
					online = true
				} else if f == "read" {
					isRead = true
				} else if f == "in sync" {
					inSync = true
				}
			}
			(*replica)["online"] = online
			(*replica)["in_sync"] = inSync
			(*replica)["read"] = isRead
		}

		if replica == nil {
			continue // al metrics below are for replicas
		}

		// migrator
		if strings.HasPrefix(line, "Migrator:") {
			s, err := getString(line)
			if err != nil {
				return nil, errors.New("failed to read `migrator`")
			}
			(*replica)["migrator"] = s
			continue
		}

		// medium_drive_type
		if strings.HasPrefix(line, "Medium drive type:") {
			s, err := getString(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `medium_drive_type` (%s)", err)
			}
			(*replica)["medium_drive_type"] = s
			continue
		}

		// extent_size
		if strings.HasPrefix(line, "Extent size:") {
			_, i, err := getSize(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `extent_size` (%s)", err)
			}
			(*replica)["extent_size"] = i
			continue
		}

		// write_pool_count
		if strings.HasPrefix(line, "Write pool count:") {
			i, err := getInt64(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `write_pool_count` (%s)", err)
			}
			(*replica)["write_pool_count"] = i
			continue
		}

		// last_write_on
		if strings.HasPrefix(line, "Last write on:") {
			s, err := getString(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `last_write_on` (%s)", err)
			}
			(*replica)["last_write_on"] = s
			continue
		}

		// free_space_on_current_partition
		if strings.HasPrefix(line, "Free space on current partition:") {
			_, i, err := getSize(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `free_space_on_current_partition` (%s)", err)
			}
			(*replica)["free_space_on_current_partition"] = i
			continue
		}

		// compression
		if strings.HasPrefix(line, "Compression:") {
			b, err := getBool(line)
			if err != nil {
				return nil, fmt.Errorf("failed to read `compression` (%s)", err)
			}
			(*replica)["compression"] = b
			continue
		}
	}

	// calculate free bytes
	if max_number_of_pages >= 0 && present_pages >= 0 {
		free_pages := max_number_of_pages - present_pages
		if free_pages >= 0 {
			item["free_pages"] = free_pages
			item["free_bytes"] = free_pages * pageSize
		}
	}

	return &params{
		params:   item,
		replicas: replicas,
	}, nil
}

func CheckLogCmd(_ *libagent.Check) (map[string][]map[string]any, error) {
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
