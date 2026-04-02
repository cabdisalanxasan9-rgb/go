package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func Write(path, format string, data any) error {
	if strings.TrimSpace(path) == "" {
		pretty, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(pretty))
		return nil
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case "txt":
		_, err = fmt.Fprintf(f, "%+v\n", data)
		return err
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
