package config

import (
	"fmt"
	"reflect"
	"strings"
)

// PrintVerbose returns a formatted representation of the
// current configuration suitable for debug logging.
//
// Fields marked as sensitive are omitted.
func (o *IngestOptions) PrintVerbose() string {
	v := reflect.ValueOf(o).Elem()
	t := v.Type()

	var sb strings.Builder
	sb.WriteString("Configuration values:\n")
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		// Skip sensitive fields
		if field.Tag.Get("sensitive") == "true" || field.Tag.Get("print") == "false" {
			continue
		}

		envName := field.Tag.Get("env")
		if envName == "" {
			envName = field.Name
		}
		value := v.Field(i).Interface()
		fmt.Fprintf(&sb, "  %-20s : %v\n", envName, value)
	}
	return sb.String()
}
