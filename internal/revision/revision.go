package revision

import (
	"fmt"
	"os"
	"strings"
	"wis2-ingest/internal/config"
)

// AppendRuntimeInfo appends runtime configuration details to the revision file.
// It does not overwrite the existing Docker image version line.
func AppendRuntimeInfo(filePath string, cfg *config.IngestOptions) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open revision file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			// Only log to stderr; we can't return it from defer directly
			fmt.Fprintf(os.Stderr, "warning: failed to close revision file: %v\n", cerr)
		}
	}()

	// Join topics slice into a single string
	topicStr := strings.Join(cfg.Topics, ",")

	content := fmt.Sprintf(
		"\nHost: %s\nPort: %s\nUsername: %s\nTopic: %s\nOutput Directory: %s\n",
		cfg.Host, cfg.Port, cfg.Username, topicStr, cfg.OutputDir,
	)

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("cannot append runtime info to revision file: %w", err)
	}

	return nil
}
