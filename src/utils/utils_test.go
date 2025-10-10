package utils

import "testing"

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		name string
		b    int64
		want string
	}{
		{"zero", 0, "0B"},
		{"bytes", 512, "512B"},
		{"just under 1KB", 1023, "1023B"},
		{"1KB", 1024, "1KB"},
		{"1.5KB", 1536, "1.5KB"},
		{"1MB", 1024 * 1024, "1MB"},
		{"20MB", 20 * 1024 * 1024, "20MB"},
		{"100MB", 100 * 1024 * 1024, "100MB"},
		{"1GB", 1024 * 1024 * 1024, "1GB"},
		{"1.5GB", int64(1.5 * 1024 * 1024 * 1024), "1.5GB"},
		{"negative", -1024, "-1KB"},
		{"large", 5 * 1024 * 1024 * 1024, "5GB"},
		{"20", 20, "20B"}, // 你关心的case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HumanBytes(tt.b)
			if got != tt.want {
				t.Errorf("HumanBytes(%d) = %v, want %v", tt.b, got, tt.want)
			}
		})
	}
}
