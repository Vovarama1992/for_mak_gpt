package speech

import (
	"os/exec"
	"strconv"
	"strings"
)

func AudioDuration(path string) (float64, error) {
	out, err := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	).Output()
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}
