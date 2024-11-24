package prototyping

import (
	"errors"
	"regexp"
)

func validateDBConfig(dbCfg *DBConfig) error {
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

func validateAPIConfig(apiCfg *APIConfig) error {
	// TODO
	return nil
}

func validateConstructors(c *map[string]func() interface{}) error {
	re := regexp.MustCompile(`[a-zA-Z0-9_]+`)
	for k, _ := range *c {
		if !re.MatchString(k) {
			return errors.New("constructor name is invalid")
		}
	}
	return nil
}
