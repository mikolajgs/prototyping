package prototyping

import (
	"errors"
)

func validateConfig(cfg *Config) error {
	// todo: proper validation
	if cfg.DatabaseDSN == "" {
		return errors.New("database dsn is missing")
	}
	return nil
}
