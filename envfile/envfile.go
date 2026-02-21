// Package envfile provides a simple dotenv file parser.
package envfile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Load reads a dotenv-formatted file and sets the key-value pairs as
// environment variables. Lines starting with # are comments. Values
// may be optionally quoted with single or double quotes.
func Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("loading env file %q: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("%s:%d: expected KEY=VALUE", path, lineNum)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Strip surrounding quotes.
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		os.Setenv(key, value)
	}
	return scanner.Err()
}

// LoadIfExists calls Load only when the file exists. A missing file
// is silently ignored.
func LoadIfExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return Load(path)
}
