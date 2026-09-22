package bridge

import "testing"

func TestParseWinRMPort(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		want    int
		wantErr bool
	}{
		{name: "default HTTP port", port: "5985", want: 5985},
		{name: "default HTTPS port", port: "5986", want: 5986},
		{name: "not numeric", port: "invalid", wantErr: true},
		{name: "zero", port: "0", wantErr: true},
		{name: "negative", port: "-1", wantErr: true},
		{name: "above TCP range", port: "65536", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseWinRMPort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseWinRMPort(%q) error = %v, wantErr %v", tt.port, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("parseWinRMPort(%q) = %d, want %d", tt.port, got, tt.want)
			}
		})
	}
}

func TestInitWinRmRejectsInvalidPort(t *testing.T) {
	client, err := InitWinRm("user", "password", "127.0.0.1", "invalid", false)
	if err == nil {
		t.Fatal("InitWinRm() error = nil, want invalid port error")
	}
	if client != nil {
		t.Fatal("InitWinRm() client is non-nil for an invalid port")
	}
}

func TestInitWinRmAcceptsValidPort(t *testing.T) {
	client, err := InitWinRm("user", "password", "127.0.0.1", "5985", false)
	if err != nil {
		t.Fatalf("InitWinRm() error = %v", err)
	}
	if client == nil {
		t.Fatal("InitWinRm() client is nil for a valid port")
	}
}

func TestVerifyWinRmRejectsInvalidPort(t *testing.T) {
	if VerifyWinRm("user", "password", "127.0.0.1", "invalid", false) {
		t.Fatal("VerifyWinRm() = true for an invalid port")
	}
}
