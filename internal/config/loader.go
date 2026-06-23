package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Load populates the configuration from all supported sources.
//
// Values are loaded in the following order:
//
//  1. Built-in defaults
//  2. Environment variables
//  3. Optional ConfigMap files
//
// The resulting configuration is then validated.
func (o *IngestOptions) Load() error {

	// ENV Variables
	if err := o.loadFromEnv(); err != nil {
		return err
	}

	// ConfigMaps
	if err := o.loadAuthPolicies(); err != nil {
		return err
	}

	return o.validate()
}

// loadFromEnv applies environment variable overrides
// to the current configuration instance.
func (o *IngestOptions) loadFromEnv() error {

	v := reflect.ValueOf(o).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {

		field := t.Field(i)

		envName := field.Tag.Get("env")
		if envName == "" {
			envName = field.Name
		}

		val := os.Getenv(envName)
		if val == "" {
			continue // keep default
		}

		switch v.Field(i).Kind() {

		case reflect.String:
			v.Field(i).SetString(val)

		case reflect.Slice:
			v.Field(i).Set(reflect.ValueOf(parseDelimitedList(val, ";")))
		}
	}

	return nil
}

// loadAuthPolicies resolves the correct file path and loads policies.
func (o *IngestOptions) loadAuthPolicies() error {

	filePath := o.AuthPoliciesConfigMapPath
	if filePath == "" {
		filePath = filepath.Join(DefaultConfigBasePath, DefaultAuthPoliciesConfigMap)
	}

	return o.LoadAuthPolicies(filePath)
}

// validate verifies that all fields marked with
// mandatory:"true" contain valid values.
func (o *IngestOptions) validate() error {

	v := reflect.ValueOf(o).Elem()
	t := v.Type()

	var missing []string

	for i := 0; i < v.NumField(); i++ {

		field := t.Field(i)

		if field.Tag.Get("mandatory") != "true" {
			continue
		}

		envName := field.Tag.Get("env")
		if envName == "" {
			envName = field.Name
		}

		value := v.Field(i)

		switch value.Kind() {

		case reflect.String:
			if strings.TrimSpace(value.String()) == "" {
				missing = append(missing, envName)
			}

		case reflect.Slice:
			if value.Len() == 0 {
				missing = append(missing, envName)
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing mandatory environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}
