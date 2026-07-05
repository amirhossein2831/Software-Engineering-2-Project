// Package config provides simple 12-factor environment configuration helpers
// shared by every service. Services read all runtime configuration from env
// vars; there are no hardcoded secrets.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Get returns the value of the environment variable named by key, or fallback
// if the variable is unset.
func Get(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// MustGet returns the value of the environment variable named by key, or panics
// if the variable is unset. Use for required secrets/connection strings.
func MustGet(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

// GetInt returns the integer value of the environment variable named by key.
// It returns fallback if the variable is unset or cannot be parsed as an int.
func GetInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
