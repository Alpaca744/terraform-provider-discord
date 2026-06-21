package discord

import (
	"reflect"
	"testing"
)

func TestPermissionsToBitfield(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    string
		wantErr bool
	}{
		{"empty", nil, "0", false},
		{"single", []string{"VIEW_CHANNEL"}, "1024", false},
		{"combined", []string{"VIEW_CHANNEL", "SEND_MESSAGES"}, "3072", false},
		{"order independent", []string{"SEND_MESSAGES", "VIEW_CHANNEL"}, "3072", false},
		{"duplicate collapses", []string{"VIEW_CHANNEL", "VIEW_CHANNEL"}, "1024", false},
		{"high bit", []string{"MODERATE_MEMBERS"}, "1099511627776", false},
		{"unknown", []string{"NOT_A_PERM"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PermissionsToBitfield(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBitfieldToPermissions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{"empty", "", []string{}, false},
		{"zero", "0", []string{}, false},
		{"single", "1024", []string{"VIEW_CHANNEL"}, false},
		{"combined sorted", "3072", []string{"SEND_MESSAGES", "VIEW_CHANNEL"}, false},
		{"unknown bits ignored", "1099511627778", []string{"KICK_MEMBERS", "MODERATE_MEMBERS"}, false},
		{"invalid", "notanumber", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BitfieldToPermissions(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionRoundTrip(t *testing.T) {
	// Every known permission must survive name -> bitfield -> name.
	for _, name := range PermissionNames() {
		bf, err := PermissionsToBitfield([]string{name})
		if err != nil {
			t.Fatalf("%s: to bitfield: %v", name, err)
		}
		names, err := BitfieldToPermissions(bf)
		if err != nil {
			t.Fatalf("%s: from bitfield: %v", name, err)
		}
		if len(names) != 1 || names[0] != name {
			t.Errorf("%s round-tripped to %v", name, names)
		}
	}
}

func TestIsPermissionName(t *testing.T) {
	if !IsPermissionName("ADMINISTRATOR") {
		t.Error("ADMINISTRATOR should be valid")
	}
	if IsPermissionName("administrator") {
		t.Error("lowercase should be invalid")
	}
}
