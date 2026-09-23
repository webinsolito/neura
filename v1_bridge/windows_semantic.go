package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type WindowsLaunchReceipt struct {
	App        string `json:"app"`
	Executable string `json:"executable"`
	PID        int    `json:"pid"`
	Verified   bool   `json:"verified"`
}

func resolveAllowedWindowsApp(name string) (string, error) {
	if runtime.GOOS != "windows" {
		return "", errors.New("windows semantic control unavailable on this OS")
	}
	name = strings.ToLower(strings.TrimSpace(name))
	allowed := map[string]string{
		"notepad": "notepad.exe",
	}
	exeName, ok := allowed[name]
	if !ok {
		return "", errors.New("windows app is not allowlisted")
	}
	root := os.Getenv("SystemRoot")
	if root == "" || !filepath.IsAbs(root) {
		return "", errors.New("invalid SystemRoot")
	}
	exe := filepath.Join(root, "System32", exeName)
	st, err := os.Stat(exe)
	if err != nil || st.IsDir() {
		return "", errors.New("allowlisted Windows executable unavailable")
	}
	return exe, nil
}

func verifyWindowsPID(ctx context.Context, pid int) bool {
	if runtime.GOOS != "windows" || pid <= 0 {
		return false
	}
	root := os.Getenv("SystemRoot")
	if root == "" || !filepath.IsAbs(root) {
		return false
	}
	tasklist := filepath.Join(root, "System32", "tasklist.exe")
	ctx2, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx2, tasklist, "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH").Output()
	if err != nil {
		return false
	}
	s := strings.TrimSpace(string(out))
	return s != "" && !strings.HasPrefix(strings.ToUpper(s), "INFO:") && strings.Contains(s, strconv.Itoa(pid))
}

func (r *ToolRegistry) launchAllowedWindowsApp(ctx context.Context, name string) (WindowsLaunchReceipt, error) {
	exe, err := resolveAllowedWindowsApp(name)
	if err != nil {
		return WindowsLaunchReceipt{}, err
	}
	cmd := exec.Command(exe)
	if err := cmd.Start(); err != nil {
		return WindowsLaunchReceipt{}, err
	}
	rec := WindowsLaunchReceipt{App: strings.ToLower(strings.TrimSpace(name)), Executable: exe, PID: cmd.Process.Pid}
	for i := 0; i < 10; i++ {
		if verifyWindowsPID(ctx, rec.PID) {
			rec.Verified = true
			return rec, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	_ = cmd.Process.Kill()
	return WindowsLaunchReceipt{}, errors.New("Windows app launch postcondition verification failed")
}
