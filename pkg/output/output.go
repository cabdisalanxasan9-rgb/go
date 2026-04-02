package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func Write(path, format string, data any) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	switch strings.ToLower(format) {
	case "json":
		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case "txt":
		_, err := fmt.Fprintf(file, "%+v\n", data)
		return err
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}
