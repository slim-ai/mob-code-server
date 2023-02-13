package crypto

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
)

var (
	ErrEndOfRecord      = errors.New("end of record")
	ErrInvalidPortValue = errors.New("invalid port value")
)

type ConfigEntry struct {
	Host   string
	Fields map[string]string
}

func (e1 ConfigEntry) Compare(e2 ConfigEntry) bool {
	if e1.Host != e2.Host {
		return false
	}
	if _, ok := e2.Fields["HostName"]; ok && e1.Fields["HostName"] != e2.Fields["HostName"] {
		return false
	}
	if _, ok := e2.Fields["User"]; ok && e1.Fields["User"] != e2.Fields["User"] {
		return false
	}
	if _, ok := e2.Fields["Port"]; ok && e1.Fields["Port"] != e2.Fields["Port"] {
		return false
	}
	if _, ok := e2.Fields["IdentityFile"]; ok && e1.Fields["IdentityFile"] != e2.Fields["IdentityFile"] {
		return false
	}
	return true
}

type Field struct {
	Key   string
	Value string
}

// TryWriteSshConfigFile will try to create, or append to .ssh/config
// if the entry exist - no update is performed
func TryWriteSshConfigFile(username string, sshDirectory string, certFileName string) error {
	configFile := path.Join(sshDirectory, "config")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fi, err := os.Create(configFile)
		if err != nil {
			panic(err)
		}
		fi.Close()
	}
	records, err := parseConfigFile(configFile)
	if err != nil {
		return err
	}
	// add the new entry and rule to allow create/destroy without
	// nonsense around IP address changing...
	records["*"] = ConfigEntry{
		Host: "*",
		Fields: map[string]string{
			"StrictHostKeyChecking": "no",
		},
	}
	records[certFileName] = ConfigEntry{
		Host: certFileName,
		Fields: map[string]string{
			"HostName":     certFileName,
			"User":         "ubuntu",
			"Port":         "22",
			"IdentityFile": path.Join(sshDirectory, certFileName),
		},
	}
	writeConfigFile(configFile, records)
	return nil
}

func parseConfigEntry(line string, entry *ConfigEntry) error {
	if strings.TrimSpace(line) == "" {
		return ErrEndOfRecord
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return fmt.Errorf("invalid config entry: %s", line)
	}

	field := fields[0]
	value := strings.TrimSpace(fields[1])
	switch field {
	case "Host":
		entry.Host = value
	case "User", "IdentityFile", "ForwardAgent", "ForwardX11",
		"StrictHostKeyChecking", "IdentityAgent", "TCPKeepAlive",
		"HostName", "ProxyCommand", "LocalForward", "RemoteForward",
		"LogLevel", "PasswordAuthentication", "PubkeyAuthentication",
		"Compression", "ControlMaster", "ControlPath", "DynamicForward",
		"ServerAliveInterval", "Port", "ControlPersist":
		entry.Fields[field] = value
	}
	return nil
}

func parseConfigFile(filePath string) (map[string]ConfigEntry, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	entries := make(map[string]ConfigEntry)
	entry := &ConfigEntry{Fields: map[string]string{}}
	var parseErr error
	building := false
	for scanner.Scan() {
		line := scanner.Text()
		err := parseConfigEntry(line, entry)
		if err != nil && errors.Is(err, ErrEndOfRecord) {
			host := strings.TrimSpace(entry.Host)
			if host != "" {
				entries[entry.Host] = *entry
				entry = &ConfigEntry{Fields: map[string]string{}}
				building = false
			}
			continue
		} else if err != nil {
			parseErr = err
		} else {
			building = true
		}
	}
	if parseErr == nil && building {
		entries[entry.Host] = *entry
	}
	delete(entries, "")
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, parseErr
}

func writeConfigFile(filePath string, entries map[string]ConfigEntry) error {
	f, err := os.OpenFile(filePath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	w := bufio.NewWriter(f)
	for _, key := range keys {
		fmt.Fprintf(w, "Host %s\n", entries[key].Host)
		for key, val := range entries[key].Fields {
			fmt.Fprintf(w, "  %s %s\n", key, val)
		}
		fmt.Fprintf(w, "\n") // Always give a new line for next entry
	}
	w.Flush()
	return nil
}
