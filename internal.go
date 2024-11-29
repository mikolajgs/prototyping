package prototyping

import (
	"errors"
)

func validateDbConfig(dbCfg *DbConfig) error {
	if dbCfg == nil {
		return errors.New("database config is missing")
	}
	// TODO: regexps
	if dbCfg.Host == "" {
		return errors.New("database host is empty")
	}
	if dbCfg.User == "" {
		return errors.New("database user is empty")
	}
	if dbCfg.Pass == "" {
		return errors.New("database password is empty")
	}
	if dbCfg.Name == "" {
		return errors.New("database name is empty")
	}
	if dbCfg.Port == "" {
		dbCfg.Port = "5432"
	}
	return nil
}

func validateHttpConfig(apiCfg *HttpConfig) error {
	// TODO
	return nil
}
