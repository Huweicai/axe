package process

import (
	"os/exec"
	"strings"
)

// DetectRunning runs pgrep and returns a set of cwds with active claude/codex processes.
func DetectRunning() map[string]bool {
	out, err := exec.Command("pgrep", "-af", "claude|codex").Output()
	if err != nil {
		return make(map[string]bool)
	}
	return ParseProcessList(string(out))
}

// ParseProcessList parses pgrep output and extracts --cwd values.
func ParseProcessList(output string) map[string]bool {
	running := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		for i, p := range parts {
			if p == "--cwd" && i+1 < len(parts) {
				running[parts[i+1]] = true
				break
			}
		}
	}
	return running
}
