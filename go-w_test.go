package main

import (
	"encoding/binary"
	"os"
	"testing"
	"time"
)

// TestFormatDuration tests the formatDuration function.
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{time.Hour + 23*time.Minute, "1:23:00"},
		{2*time.Hour + 5*time.Minute, "2:05:00"},
		{59 * time.Second, "0:59"},
		{0, "0:00"},
	}

	for _, test := range tests {
		result := formatDuration(test.duration)
		if result != test.expected {
			t.Errorf("formatDuration(%v) = %v; expected %v", test.duration, result, test.expected)
		}
	}
}

// TestFormatTime tests the formatTime function.
func TestFormatTime(t *testing.T) {
	// Define the Unix timestamps and their expected UTC times
	tests := []struct {
		sec      int64
		expected string
	}{
		{1672502400, "16:00"}, // 2023-01-01 00:00:00 UTC -> 16:00 UTC (your local timezone offset)
		{1672545600, "04:00"}, // 2023-01-01 12:00:00 UTC -> 04:00 UTC (your local timezone offset)
	}

	for _, test := range tests {
		// Print local time and UTC time for debugging
		localTime := time.Unix(test.sec, 0).Local().Format("15:04")
		utcTime := time.Unix(test.sec, 0).UTC().Format("15:04")
		t.Logf("Timestamp: %v, Local Time: %v, UTC Time: %v", test.sec, localTime, utcTime)

		// Test the formatTime function
		result := formatTime(test.sec)
		if result != test.expected {
			t.Errorf("formatTime(%v) = %v; expected %v", test.sec, result, test.expected)
		}
	}
}

// TestParseUtmp tests the parseUtmp function with a mock utmp file.
func TestParseUtmp(t *testing.T) {
	// Create a mock utmp file
	mockUtmpData := make([]byte, binary.Size(utmp{})) // Create a byte slice of the correct size

	// Fill in the fields
	binary.LittleEndian.PutUint16(mockUtmpData[0:2], 7)                      // Type = 7 (USER_PROCESS)
	binary.LittleEndian.PutUint32(mockUtmpData[4:8], 123)                    // Pid = 123
	copy(mockUtmpData[8:40], []byte("tty1\x00"))                             // Line = "tty1"
	copy(mockUtmpData[40:44], []byte("id1\x00"))                             // ID = "id1"
	copy(mockUtmpData[44:76], []byte("user1\x00"))                           // User = "user1"
	copy(mockUtmpData[76:332], []byte("host1\x00"))                          // Host = "host1"
	binary.LittleEndian.PutUint64(mockUtmpData[332:340], uint64(1672502400)) // Time = 2023-01-01 00:00:00 UTC

	// Write mock data to a temporary file
	tmpFile, err := os.CreateTemp("", "utmp")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(mockUtmpData); err != nil {
		t.Fatalf("Failed to write mock data: %v", err)
	}
	tmpFile.Close()

	// Override the utmp path for testing
	oldUtmpPath := utmpPath
	utmpPath = tmpFile.Name()
	defer func() {
		utmpPath = oldUtmpPath
	}()

	// Parse the mock utmp file
	sessions, method, err := parseUtmp()
	if err != nil {
		t.Fatalf("parseUtmp failed: %v", err)
	}

	// Verify the parsed data
	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	session := sessions[0]
	if session.User != "user1" {
		t.Errorf("Expected user 'user1', got '%s'", session.User)
	}
	if session.TTY != "tty1" {
		t.Errorf("Expected TTY 'tty1', got '%s'", session.TTY)
	}
	if session.From != "host1" {
		t.Errorf("Expected host 'host1', got '%s'", session.From)
	}
	if session.LoginAt != "00:00" {
		t.Errorf("Expected login time '00:00', got '%s'", session.LoginAt)
	}
	if method != "using /var/run/utmp" {
		t.Errorf("Expected method 'using /var/run/utmp', got '%s'", method)
	}
}

// TestGetSystemInfo tests the getSystemInfo function with mocked file reads.
func TestGetSystemInfo(t *testing.T) {
	// Mock /proc/uptime
	uptimeData := "12345.67 23456.78\n"
	uptimeFile, err := os.CreateTemp("", "uptime")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(uptimeFile.Name())

	if _, err := uptimeFile.WriteString(uptimeData); err != nil {
		t.Fatalf("Failed to write mock uptime data: %v", err)
	}
	uptimeFile.Close()

	// Mock /proc/loadavg
	loadAvgData := "0.15 0.10 0.05 1/100 12345\n"
	loadAvgFile, err := os.CreateTemp("", "loadavg")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(loadAvgFile.Name())

	if _, err := loadAvgFile.WriteString(loadAvgData); err != nil {
		t.Fatalf("Failed to write mock loadavg data: %v", err)
	}
	loadAvgFile.Close()

	// Override the file paths for testing
	oldUptimePath := uptimePath
	oldLoadAvgPath := loadAvgPath
	uptimePath = uptimeFile.Name()
	loadAvgPath = loadAvgFile.Name()
	defer func() {
		uptimePath = oldUptimePath
		loadAvgPath = oldLoadAvgPath
	}()

	// Call getSystemInfo
	info, err := getSystemInfo()
	if err != nil {
		t.Fatalf("getSystemInfo failed: %v", err)
	}

	// Verify the results
	expectedUptime := "3:25:45"
	if info.Uptime != expectedUptime {
		t.Errorf("Expected uptime '%s', got '%s'", expectedUptime, info.Uptime)
	}

	expectedLoadAvg := "0.15 0.10 0.05"
	if info.LoadAvg != expectedLoadAvg {
		t.Errorf("Expected load average '%s', got '%s'", expectedLoadAvg, info.LoadAvg)
	}
}
