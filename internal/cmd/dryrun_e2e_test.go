package cmd

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDryRunE2E_MutatingCommandsSkipAuthAndAPI(t *testing.T) {
	cases := []struct {
		name string
		args []string
		op   string
	}{
		{
			name: "contacts create",
			args: []string{"contacts", "create", "--given", "Smoke", "--email", "smoke@example.com"},
			op:   "contacts.create",
		},
		{
			name: "contacts update",
			args: []string{"contacts", "update", "people/123", "--given", "Smoke"},
			op:   "contacts.update",
		},
		{
			name: "docs insert",
			args: []string{"docs", "insert", "doc123", "hello"},
			op:   "docs.insert",
		},
		{
			name: "drive move",
			args: []string{"drive", "move", "file123", "--parent", "folder123"},
			op:   "drive.move",
		},
		{
			name: "drive rename",
			args: []string{"drive", "rename", "file123", "New"},
			op:   "drive.rename",
		},
		{
			name: "calendar create-calendar",
			args: []string{"calendar", "create-calendar", "SmokeCal", "--timezone", "UTC"},
			op:   "calendar.create-calendar",
		},
		{
			name: "forms create",
			args: []string{"forms", "create", "--title", "SmokeForm"},
			op:   "forms.create",
		},
		{
			name: "forms publish",
			args: []string{"forms", "publish", "form123"},
			op:   "forms.publish",
		},
		{
			name: "forms watch create",
			args: []string{"forms", "watch", "create", "form123", "--topic", "projects/p/topics/t"},
			op:   "forms.watches.create",
		},
		{
			name: "forms watch delete",
			args: []string{"forms", "watch", "delete", "form123", "watch123"},
			op:   "forms.watches.delete",
		},
		{
			name: "forms watch renew",
			args: []string{"forms", "watch", "renew", "form123", "watch123"},
			op:   "forms.watches.renew",
		},
		{
			name: "forms move question",
			args: []string{"forms", "move-question", "form123", "0", "1"},
			op:   "forms.moveQuestion",
		},
		{
			name: "gmail label rename",
			args: []string{"gmail", "labels", "rename", "Label_1", "NewLabel"},
			op:   "gmail.labels.rename",
		},
		{
			name: "gmail label style",
			args: []string{"gmail", "labels", "style", "Label_1", "--background-color", "#ffffff", "--text-color", "#000000"},
			op:   "gmail.labels.style",
		},
		{
			name: "gmail label delete",
			args: []string{"gmail", "labels", "delete", "Label_1"},
			op:   "gmail.labels.delete",
		},
		{
			name: "meet update",
			args: []string{"meet", "update", "abc-defg-hij", "--access", "open"},
			op:   "meet.spaces.patch",
		},
		{
			name: "meet end",
			args: []string{"meet", "end", "abc-defg-hij"},
			op:   "meet.spaces.end_active_conference",
		},
		{
			name: "slides create",
			args: []string{"slides", "create", "SmokeSlides"},
			op:   "slides.create",
		},
		{
			name: "slides create from template",
			args: []string{"slides", "create-from-template", "template123", "SmokeSlides", "--replace", "NAME=World"},
			op:   "slides.create-from-template",
		},
		{
			name: "appscript create",
			args: []string{"appscript", "create", "--title", "SmokeScript"},
			op:   "appscript.create",
		},
		{
			name: "sheets banding clear all",
			args: []string{"sheets", "banding", "clear", "sheet123", "--sheet", "Sheet1", "--all"},
			op:   "sheets.banding.clear",
		},
		{
			name: "sheets conditional clear index",
			args: []string{"sheets", "conditional-format", "clear", "sheet123", "--sheet", "Sheet1", "--index", "0"},
			op:   "sheets.conditional-format.clear",
		},
		{
			name: "sheets conditional clear all",
			args: []string{"sheets", "conditional-format", "clear", "sheet123", "--sheet", "Sheet1", "--all"},
			op:   "sheets.conditional-format.clear",
		},
		{
			name: "sheets table delete",
			args: []string{"sheets", "table", "delete", "sheet123", "Tbl"},
			op:   "sheets.table.delete",
		},
		{
			name: "sheets table append",
			args: []string{"sheets", "table", "append", "sheet123", "Tbl", "a|b"},
			op:   "sheets.table.append",
		},
		{
			name: "sheets table clear",
			args: []string{"sheets", "table", "clear", "sheet123", "Tbl"},
			op:   "sheets.table.clear",
		},
		{
			name: "sheets named ranges add",
			args: []string{"sheets", "named-ranges", "add", "sheet123", "MyRange", "Sheet1!A1:B2"},
			op:   "sheets.named_ranges.add",
		},
		{
			name: "sheets named ranges update",
			args: []string{"sheets", "named-ranges", "update", "sheet123", "range123", "--name", "NewRange"},
			op:   "sheets.named_ranges.update",
		},
		{
			name: "sheets named ranges delete",
			args: []string{"sheets", "named-ranges", "delete", "sheet123", "range123"},
			op:   "sheets.named_ranges.delete",
		},
		{
			name: "forms delete question",
			args: []string{"forms", "delete-question", "form123", "0"},
			op:   "forms.deleteQuestion",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"--json", "--dry-run", "--no-input", "--access-token", "invalid-token"}, tc.args...)
			var stderr string
			out := captureStdout(t, func() {
				stderr = captureStderr(t, func() {
					if err := Execute(args); err != nil && ExitCode(err) != 0 {
						t.Fatalf("Execute: %v", err)
					}
				})
			})
			if stderr != "" {
				t.Fatalf("dry-run touched auth/API stderr: %q", stderr)
			}

			var payload struct {
				DryRun bool   `json:"dry_run"`
				Op     string `json:"op"`
			}
			if err := json.Unmarshal([]byte(out), &payload); err != nil {
				t.Fatalf("decode dry-run output: %v\nout=%q", err, out)
			}
			if !payload.DryRun || payload.Op != tc.op {
				t.Fatalf("unexpected dry-run output: %#v", payload)
			}
		})
	}
}

func TestDryRunE2E_ValidatesFormsAndSheetsLocalInputs(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{
			name: "forms add choice requires options before auth",
			args: []string{"forms", "add-question", "form123", "--title", "Q", "--type", "radio"},
		},
		{
			name: "forms add scale rejects inverted range",
			args: []string{"forms", "add-question", "form123", "--title", "Q", "--type", "scale", "--scale-low", "5", "--scale-high", "1"},
		},
		{
			name: "forms update requires a field before auth",
			args: []string{"forms", "update", "form123"},
		},
		{
			name: "forms update validates quiz before dry-run",
			args: []string{"forms", "update", "form123", "--quiz", "maybe"},
		},
		{
			name: "sheets conditional clear validates index before auth",
			args: []string{"sheets", "conditional-format", "clear", "sheet123", "--sheet", "Sheet1", "--index", "-1"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"--json", "--dry-run", "--no-input", "--access-token", "invalid-token"}, tc.args...)
			_ = captureStdout(t, func() {
				_ = captureStderr(t, func() {
					if err := Execute(args); ExitCode(err) == 0 {
						t.Fatalf("expected validation failure")
					}
				})
			})
		})
	}
}

func TestDryRunE2E_ContactsUpdateValidatesLocalInputs(t *testing.T) {
	tempDir := t.TempDir()
	malformed := tempDir + "/malformed.json"
	unsupported := tempDir + "/unsupported.json"
	mismatch := tempDir + "/mismatch.json"
	valid := tempDir + "/valid.json"
	for path, body := range map[string]string{
		malformed:   "{",
		unsupported: `{"notAContactField":true}`,
		mismatch:    `{"resourceName":"people/other","names":[{"givenName":"Dry"}]}`,
		valid:       `{"resourceName":"people/123","names":[{"givenName":"Dry"}]}`,
	} {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatalf("write fixture %s: %v", path, err)
		}
	}

	invalidCases := []struct {
		name string
		args []string
	}{
		{
			name: "bad birthday",
			args: []string{"contacts", "update", "people/123", "--birthday", "nope"},
		},
		{
			name: "bad custom",
			args: []string{"contacts", "update", "people/123", "--custom", "bad"},
		},
		{
			name: "bad relation",
			args: []string{"contacts", "update", "people/123", "--relation", "bad"},
		},
		{
			name: "malformed from-file",
			args: []string{"contacts", "update", "people/123", "--from-file", malformed},
		},
		{
			name: "unsupported from-file key",
			args: []string{"contacts", "update", "people/123", "--from-file", unsupported},
		},
		{
			name: "resource mismatch from-file",
			args: []string{"contacts", "update", "people/123", "--from-file", mismatch},
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"--json", "--dry-run", "--no-input", "--access-token", "invalid-token"}, tc.args...)
			_ = captureStdout(t, func() {
				_ = captureStderr(t, func() {
					if err := Execute(args); ExitCode(err) == 0 {
						t.Fatalf("expected validation failure")
					}
				})
			})
		})
	}

	t.Run("valid from-file skips auth and API", func(t *testing.T) {
		args := []string{"--json", "--dry-run", "--no-input", "--access-token", "invalid-token", "contacts", "update", "people/123", "--from-file", valid}
		out := captureStdout(t, func() {
			_ = captureStderr(t, func() {
				if err := Execute(args); err != nil && ExitCode(err) != 0 {
					t.Fatalf("Execute: %v", err)
				}
			})
		})

		var payload struct {
			DryRun bool   `json:"dry_run"`
			Op     string `json:"op"`
		}
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatalf("decode dry-run output: %v\nout=%q", err, out)
		}
		if !payload.DryRun || payload.Op != "contacts.update" {
			t.Fatalf("unexpected dry-run output: %#v", payload)
		}
	})
}
