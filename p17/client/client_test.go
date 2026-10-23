package main

import (
	"reflect"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      *command
		wantError bool
	}{
		{
			name:  "create",
			input: "CREATE TASK Learn Go",
			want:  &command{action: "CREATE", arguments: []string{"TASK", "Learn Go"}},
		},
		{
			name:  "get",
			input: "GET TASK-1",
			want:  &command{action: "GET", arguments: []string{"TASK-1"}},
		},
		{
			name:  "list",
			input: "LIST",
			want:  &command{action: "LIST"},
		},
		{
			name:  "complete",
			input: "COMPLETE TASK-1",
			want:  &command{action: "COMPLETE", arguments: []string{"TASK-1"}},
		},
		{
			name:  "delete",
			input: "DELETE TASK-1",
			want:  &command{action: "DELETE", arguments: []string{"TASK-1"}},
		},
		{
			name:  "trims surrounding whitespace",
			input: "  LIST  ",
			want:  &command{action: "LIST"},
		},
		{
			name:      "rejects incomplete command",
			input:     "GET",
			wantError: true,
		},
		{
			name:      "rejects trailing arguments",
			input:     "GET TASK-1 extra",
			wantError: true,
		},
		{
			name:      "rejects unknown command",
			input:     "UNKNOWN",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCommand(tt.input)
			if (err != nil) != tt.wantError {
				t.Fatalf("parseCommand() error = %v, wantError %t", err, tt.wantError)
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseCommand() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
