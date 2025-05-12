package main

import (
	"strings"
	"testing"
)

func TestAepcli(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
			errMsg:  "requires at least 1 arg(s), only received 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := aepcli(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("aepcli() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("aepcli() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("aepcli() unexpected error = %v", err)
			}

			if code != 0 {
				t.Errorf("aepcli() unexpected exit code = %v, want 0", code)
			}
		})
	}
}
