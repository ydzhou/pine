package pine

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-runewidth"
)

func runeRenderedWidth(
	index int,
	data rune,
) int {
	if data == rune('\t') {
		return TABWIDTH - index%TABWIDTH
	}
	if data == ' ' {
		return 1
	}
	return runewidth.RuneWidth(data)
}

func expandHomeDir(path string) (string, error) {
	homeDir, err := os.UserHomeDir()
	fullPath := path
	if path == "~" {
		fullPath = homeDir
	} else if strings.HasPrefix(path, "~/") {
		fullPath = filepath.Join(homeDir, path[2:])
	}
	fullPath = getAbsoluteFilePath(fullPath)
	if err != nil {
		return "", err
	}

	return fullPath, nil
}

func getAbsoluteFilePath(path string) string {
	if absolutePath, err := filepath.Abs(path); err == nil {
		return absolutePath
	} else {
		return "NA"
	}
}

func getFilename(path string) string {
	return filepath.Base(path)
}
