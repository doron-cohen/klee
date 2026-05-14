package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// parseDotEnvFiles reads each .env file in order; later files win for duplicate keys.
// Missing files are silently skipped.
func parseDotEnvFiles(paths []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("opening dotenv file %s: %w", path, err)
		}
		entries, parseErr := parseDotEnv(f)
		closeErr := f.Close()
		if parseErr != nil {
			return nil, fmt.Errorf("parsing dotenv file %s: %w", path, parseErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("closing dotenv file %s: %w", path, closeErr)
		}
		for k, v := range entries {
			result[k] = v
		}
	}
	return result, nil
}

func parseDotEnv(f *os.File) (map[string]string, error) {
	result := make(map[string]string)
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			return nil, fmt.Errorf("line %d: missing '=' in %q", lineNum, line)
		}
		key := line[:idx]
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNum)
		}
		result[key] = line[idx+1:]
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
