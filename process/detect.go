package process

import (
	"os/exec"
	"strings"
)

func DetectRunning() map[string]bool {
	pids := getPIDs()
	if len(pids) == 0 {
		return make(map[string]bool)
	}
	return getCwds(pids)
}

func getPIDs() []string {
	out, err := exec.Command("pgrep", "-f", "claude|codex").Output()
	if err != nil {
		return nil
	}
	var pids []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		pid := strings.TrimSpace(line)
		if pid != "" {
			pids = append(pids, pid)
		}
	}
	return pids
}

func getCwds(pids []string) map[string]bool {
	running := make(map[string]bool)
	args := []string{"-a", "-d", "cwd", "-Fn"}
	for _, pid := range pids {
		args = append(args, "-p", pid)
	}
	out, err := exec.Command("lsof", args...).Output()
	if err != nil {
		return running
	}
	return ParseLsofOutput(string(out))
}

func ParseLsofOutput(output string) map[string]bool {
	running := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "n/") {
			running[line[1:]] = true
		}
	}
	return running
}

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
