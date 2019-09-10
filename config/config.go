package config

import (
	"bytes"
	"fmt"
	"github.com/danlamanna/rivet/girder"
	"log"
	"os"
	"path"

	"github.com/burntsushi/toml"
)

type Config struct {
	ConfigVersion int        `toml:"config_version"`
	Profiles      []*Profile `toml:"profiles"`
}

type Profile struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
	Auth string `toml:"auth"`
}

// Read the default profile and return it, or nil
func ReadDefaultProfile(ctx *girder.Context) (*Profile, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	config := new(Config)
	configFile := path.Join(homeDir, ".rivet", "config.toml")
	ctx.Logger.Debugf("attempting to load config file %s", configFile)
	if _, err := os.Stat(configFile); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to access config file %s, err: %s", configFile, err)
	}
	ctx.Logger.Debugf("loaded config file %s", configFile)

	_, err = toml.DecodeFile(configFile, config)
	if err != nil {
		return nil, err
	}

	if len(config.Profiles) == 0 {
		return nil, nil
	}

	return config.Profiles[0], nil

}

func WriteDefaultProfile(auth string, url string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	config := new(Config)
	config.ConfigVersion = 1
	config.Profiles = make([]*Profile, 1)
	config.Profiles[0] = &Profile{Name: "default", URL: url, Auth: auth}

	configFile := path.Join(homeDir, ".rivet", "config.toml")
	os.MkdirAll(path.Join(homeDir, ".rivet"), 0755)
	f, err := os.Create(configFile)
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(config); err != nil {
		log.Fatal(err)
	}
	_, err = f.Write(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	return nil
}
