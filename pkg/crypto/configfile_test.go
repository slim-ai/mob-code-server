package crypto

import (
	"os"
	"testing"
)

func TestParseConfigFile(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    map[string]ConfigEntry
		expectedErr error
	}{
		{
			name:        "empty config file",
			input:       "",
			expected:    map[string]ConfigEntry{},
			expectedErr: nil,
		},
		{
			name:  "valid config file",
			input: simpleEntries,
			expected: map[string]ConfigEntry{
				"test-host": {
					Host: "test-host",
					Fields: map[string]string{
						"HostName":     "test-host",
						"User":         "test-user",
						"Port":         "22",
						"IdentityFile": "/path/to/id_rsa",
					},
				},
				"test-host-2": {
					Host: "test-host-2",
					Fields: map[string]string{
						"HostName":     "test-host-2",
						"User":         "test-user-2",
						"Port":         "22",
						"IdentityFile": "/path/to/id_rsa_2",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:  "valid config file with duplicates and no newline between some records",
			input: bigFileWithDupsAndMissingNewlinesBetweenRecords,
			expected: map[string]ConfigEntry{
				"*": {
					Host: "*",
					Fields: map[string]string{
						"StrictHostKeyChecking": "no",
					},
				},
				"bigid-sandbox.dev.saas.getslim.ai": {
					Host: "bigid-sandbox.dev.saas.getslim.ai",
					Fields: map[string]string{
						"HostName":     "bigid-sandbox.dev.saas.getslim.ai",
						"User":         "arch",
						"IdentityFile": "~/.ssh/kyles-bigid-key-pair.pem",
					},
				},
				"code.dev-jb.saas.getslim.ai": {
					Host: "code.dev-jb.saas.getslim.ai",
					Fields: map[string]string{
						"HostName":     "code.dev-jb.saas.getslim.ai",
						"User":         "ubuntu",
						"Port":         "22",
						"IdentityFile": "/Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai",
					},
				},
				"nikita.dev-jb.saas.getslim.ai": {
					Host: "nikita.dev-jb.saas.getslim.ai",
					Fields: map[string]string{
						"HostName":     "nikita.dev-jb.saas.getslim.ai",
						"User":         "ubuntu",
						"Port":         "22",
						"IdentityFile": "/Users/josephbarnett/.ssh/nikita.dev-jb.saas.getslim.ai",
					},
				},
				"money.dev-jb.saas.getslim.ai": {
					Host: "money.dev-jb.saas.getslim.ai",
					Fields: map[string]string{
						"HostName":     "money.dev-jb.saas.getslim.ai",
						"User":         "ubuntu",
						"Port":         "22",
						"IdentityFile": "/Users/josephbarnett/.ssh/money.dev-jb.saas.getslim.ai",
					},
				},
				"candidate.dev-jb.saas.getslim.ai": {
					Host: "candidate.dev-jb.saas.getslim.ai",
					Fields: map[string]string{
						"HostName":     "candidate.dev-jb.saas.getslim.ai",
						"User":         "ubuntu",
						"Port":         "22",
						"IdentityFile": "/Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai",
					},
				},
				"codi.dev-jb.saas.getslim.ai": {
					Host: "codi.dev-jb.saas.getslim.ai",
					Fields: map[string]string{
						"HostName":     "codi.dev-jb.saas.getslim.ai",
						"User":         "ubuntu",
						"Port":         "22",
						"IdentityFile": "/Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai",
					},
				},
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "config")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())
			if _, err := tmpfile.Write([]byte(tc.input)); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			entries, err := parseConfigFile(tmpfile.Name())
			if err != tc.expectedErr {
				t.Errorf("Expected error to be %v, but got %v", tc.expectedErr, err)
			}
			if len(entries) != len(tc.expected) {
				t.Errorf("Expected %d entries,but got %d", len(tc.expected), len(entries))
			}
			for host, entry := range tc.expected {
				if e, ok := entries[host]; !ok {
					t.Errorf("Expected entry for host %s not found", host)
				} else if !entry.Compare(e) {
					t.Errorf("Expected entry for host %s to be %+v, but got %+v", host, entry, e)
				}
			}
		})
	}
}

const simpleEntries = `
Host test-host
    HostName test-host
    User test-user
    Port 22
    IdentityFile /path/to/id_rsa

Host test-host-2
    HostName test-host-2
    User test-user-2
    Port 22
    IdentityFile /path/to/id_rsa_2
`

const bigFileWithDupsAndMissingNewlinesBetweenRecords = `
Host *
        StrictHostKeyChecking no

Host bigid-sandbox.dev.saas.getslim.ai
        HostName bigid-sandbox.dev.saas.getslim.ai
        IdentityFile ~/.ssh/kyles-bigid-key-pair.pem
        User arch

Host code.dev-jb.saas.getslim.ai
    HostName code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai


Host nikita.dev-jb.saas.getslim.ai
    Hostname nikita.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/nikita.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai
Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai
Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai
Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host money.dev-jb.saas.getslim.ai
    Hostname money.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/money.dev-jb.saas.getslim.ai

Host money.dev-jb.saas.getslim.ai
    Hostname money.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/money.dev-jb.saas.getslim.ai

Host money.dev-jb.saas.getslim.ai
    Hostname money.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/money.dev-jb.saas.getslim.ai

Host money.dev-jb.saas.getslim.ai
    Hostname money.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/money.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai
`

const bigFileWithDupsMissingNewline = `
Host *
        StrictHostKeyChecking no

Host bigid-sandbox.dev.saas.getslim.ai
        HostName bigid-sandbox.dev.saas.getslim.ai
        IdentityFile ~/.ssh/kyles-bigid-key-pair.pem
        User arch

Host code.dev-jb.saas.getslim.ai
    HostName code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai


Host nikita.dev-jb.saas.getslim.ai
    Hostname nikita.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/nikita.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host code.dev-jb.saas.getslim.ai
    Hostname code.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/code.dev-jb.saas.getslim.ai

Host money.dev-jb.saas.getslim.ai
    Hostname money.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/money.dev-jb.saas.getslim.ai
Host money.dev-jb.saas.getslim.ai
    Hostname money.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/money.dev-jb.saas.getslim.ai

Host money.dev-jb.saas.getslim.ai
    Hostname money.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/money.dev-jb.saas.getslim.ai

Host money.dev-jb.saas.getslim.ai
    Hostname money.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/money.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host candidate.dev-jb.saas.getslim.ai
    Hostname candidate.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/candidate.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai

Host codi.dev-jb.saas.getslim.ai
    Hostname codi.dev-jb.saas.getslim.ai
    User ubuntu
    Port 22
    IdentityFile /Users/josephbarnett/.ssh/codi.dev-jb.saas.getslim.ai
`
