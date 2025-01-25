package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// utmp represents the structure of an entry in the utmp file.
type utmp struct {
	Type int16     // Type of login
	_    [2]byte   // Padding
	Pid  int32     // Process ID
	Line [32]byte  // Device name (tty)
	ID   [4]byte   // Terminal name suffix or ID
	User [32]byte  // Username
	Host [256]byte // Hostname for remote login
	Exit struct {  // Exit status
		Termination int16
		Exit        int16
	}
	Session int32    // Session ID
	Time    int64    // Time entry was made
	Addr    [4]int32 // Internet address of remote host
	Unused  [20]byte // Reserved for future use
}

// SystemInfo holds system-related information.
type SystemInfo struct {
	CurrentTime string
	Uptime      string
	LoadAvg     string
}

// UserSession holds information about a logged-in user session.
type UserSession struct {
	User    string
	TTY     string
	From    string
	LoginAt string
	Idle    string
	JCPU    string
	PCPU    string
	What    string
}

// File paths for system information
var (
	utmpPath    = "/var/run/utmp"
	uptimePath  = "/proc/uptime"
	loadAvgPath = "/proc/loadavg"
)

// getSystemInfo retrieves system information (uptime, load averages, etc.).
func getSystemInfo() (SystemInfo, error) {
	uptime, err := readUptime()
	if err != nil {
		return SystemInfo{}, fmt.Errorf("failed to read uptime: %w", err)
	}

	loadAvg, err := readLoadAverage()
	if err != nil {
		return SystemInfo{}, fmt.Errorf("failed to read load average: %w", err)
	}

	return SystemInfo{
		CurrentTime: time.Now().Format("15:04:05"),
		Uptime:      formatDuration(uptime),
		LoadAvg:     loadAvg,
	}, nil
}

// readUptime reads the system uptime from /proc/uptime.
func readUptime() (time.Duration, error) {
	data, err := os.ReadFile(uptimePath)
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(data))
	uptimeSeconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}

	return time.Duration(uptimeSeconds * float64(time.Second)), nil
}

// readLoadAverage reads the system load averages from /proc/loadavg.
func readLoadAverage() (string, error) {
	data, err := os.ReadFile(loadAvgPath)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		return strings.Join(fields[:3], " "), nil
	}
	return "", fmt.Errorf("invalid loadavg format")
}

// parseUtmp reads and parses the utmp file to extract user sessions.
func parseUtmp() ([]UserSession, string, error) {
	// Check if /var/run/utmp exists
	if _, err := os.Stat(utmpPath); err == nil {
		sessions, err := parseUtmpFile(utmpPath)
		return sessions, "using /var/run/utmp", err
	}

	// Fall back to using /proc
	sessions, err := parseProc()
	return sessions, "using /proc", err
}

// parseUtmpFile reads and parses the utmp file.
func parseUtmpFile(filePath string) ([]UserSession, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open utmp file: %w", err)
	}
	defer file.Close()

	var sessions []UserSession
	for {
		var entry utmp
		if err := binary.Read(file, binary.LittleEndian, &entry); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to read utmp entry: %w", err)
		}

		if entry.Type == 7 { // USER_PROCESS
			sessions = append(sessions, UserSession{
				User:    strings.TrimRight(string(entry.User[:]), "\x00"),
				TTY:     strings.TrimRight(string(entry.Line[:]), "\x00"),
				From:    strings.TrimRight(string(entry.Host[:]), "\x00"),
				LoginAt: formatTime(entry.Time),
				Idle:    ".",
				JCPU:    "0.00s",
				PCPU:    "0.00s",
				What:    "-",
			})
		}
	}

	return sessions, nil
}

// parseProc retrieves logged-in users using /proc.
func parseProc() ([]UserSession, error) {
	var sessions []UserSession

	// Iterate over all processes in /proc
	procDir := "/proc"
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue // Skip non-PID directories
		}

		// Get the username for the process
		user, err := getUserFromPID(pid)
		if err != nil {
			continue
		}

		// Get the terminal (TTY) for the process
		tty, err := getTTYFromPID(pid)
		if err != nil {
			continue
		}

		// Add the session to the list
		sessions = append(sessions, UserSession{
			User:    user,
			TTY:     tty,
			From:    "?", // Remote host not available in /proc
			LoginAt: "?", // Login time not available in /proc
			Idle:    ".",
			JCPU:    "0.00s",
			PCPU:    "0.00s",
			What:    "-",
		})
	}

	return sessions, nil
}

// getUserFromPID retrieves the username for a given process ID.
func getUserFromPID(pid int) (string, error) {
	statusFile := fmt.Sprintf("/proc/%d/status", pid)
	data, err := os.ReadFile(statusFile)
	if err != nil {
		return "", fmt.Errorf("failed to read status file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				uid, err := strconv.Atoi(fields[1])
				if err != nil {
					return "", fmt.Errorf("failed to parse UID: %w", err)
				}
				user, err := getUserByUID(uid)
				if err != nil {
					return "", fmt.Errorf("failed to get user by UID: %w", err)
				}
				return user.Username, nil
			}
		}
	}
	return "", fmt.Errorf("UID not found in status file")
}

// getUserByUID retrieves the username for a given UID.
func getUserByUID(uid int) (*user.User, error) {
	return user.LookupId(strconv.Itoa(uid))
}

// getTTYFromPID retrieves the terminal (TTY) for a given process ID.
func getTTYFromPID(pid int) (string, error) {
	fdDir := fmt.Sprintf("/proc/%d/fd", pid)
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return "", fmt.Errorf("failed to read fd directory: %w", err)
	}

	for _, entry := range entries {
		link, err := os.Readlink(filepath.Join(fdDir, entry.Name()))
		if err != nil {
			continue
		}
		if strings.HasPrefix(link, "/dev/tty") || strings.HasPrefix(link, "/dev/pts") {
			return filepath.Base(link), nil
		}
	}
	return "?", nil
}

// formatTime formats a Unix timestamp into a human-readable time string.
func formatTime(sec int64) string {
	return time.Unix(sec, 0).UTC().Format("15:04")
}

// formatDuration formats a duration into a human-readable string (e.g., "1:23").
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// displayHeader prints the header of the `w` output with colors.
func displayHeader(info SystemInfo, method string) {
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf(" %s up %s,  load average: %s (%s)\n",
		cyan(info.CurrentTime),
		yellow(info.Uptime),
		yellow(info.LoadAvg),
		method,
	)
	fmt.Println(color.New(color.FgHiWhite).Sprint("USER     TTY      FROM             LOGIN@   IDLE   JCPU   PCPU WHAT"))
}

// displaySessions prints the list of user sessions with colors.
func displaySessions(sessions []UserSession) {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()

	for _, session := range sessions {
		fmt.Printf("%-8s %-8s %-16s %-8s %-6s %-6s %-6s %s\n",
			green(session.User),
			blue(session.TTY),
			magenta(session.From),
			session.LoginAt,
			session.Idle,
			session.JCPU,
			session.PCPU,
			session.What,
		)
	}
}

func main() {
	// Retrieve system information
	info, err := getSystemInfo()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Parse user sessions
	sessions, method, err := parseUtmp()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Display the output with colors
	displayHeader(info, method)
	displaySessions(sessions)
}
