package conf

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"path/filepath"
	//"os"
	"os"
	"strings"
)

type ProConf struct {
	FilePath string
	FileName string
	File     string

	ViperDefault *viper.Viper
	ViperUser    *viper.Viper

	ToLog    bool
	ToStdout bool

	ConfUserExist bool
}

func NewConf(file string) *ProConf {
	newProConf := ProConf{
		FilePath:      filepath.Dir(file),
		FileName:      strings.Split(filepath.Base(file), ".")[0],
		File:          file,
		ViperDefault:  viper.New(),
		ViperUser:     viper.New(),
		ConfUserExist: false,
	}

	err := newProConf.Refresh()
	if err != nil {
		fmt.Printf(err.Error())
	}
	return &newProConf
}

func (c *ProConf) Get(key string, defaultValue interface{}) interface{} {
	if c.ConfUserExist {
		ret := c.ViperUser.Get(key)
		if ret == nil {
			ret = c.ViperDefault.Get(key)
			if ret == nil {
				ret = defaultValue
			}
		}

		return ret
	} else {
		ret := c.ViperDefault.Get(key)
		if ret == nil {
			ret = defaultValue
		}
		return ret
	}
}

func (c *ProConf) Set(key string, value interface{}) error {
	c.ViperUser.Set(key, value)
	err := c.ViperUser.WriteConfig()

	if err != nil {
		fmt.Printf(err.Error())
		return err
	}

	return nil
}

func (c *ProConf) Refresh() error {
	errMsg := ""
	c.CheckConfUserExist()

	c.ViperDefault.SetConfigName(c.FileName)
	c.ViperDefault.AddConfigPath(c.FilePath)
	errDefault := c.ViperDefault.ReadInConfig()

	if c.ConfUserExist {
		c.ViperUser.SetConfigName(c.FileName + ".user")
		c.ViperUser.AddConfigPath(c.FilePath)
		errUser := c.ViperUser.ReadInConfig()

		if errUser != nil {
			errMsg += errUser.Error()
			errMsg += ","
		}
	}

	if errDefault != nil {
		errMsg := ""
		if errDefault != nil {
			errMsg += errDefault.Error()
		}
	}

	if errMsg == "" {
		return nil
	} else {
		err := errors.New(errMsg + "\n")
		return err
	}
}

func (c *ProConf) CheckConfUserExist() {
	if _, err := os.Stat(c.File); os.IsNotExist(err) {
		c.ConfUserExist = false
	} else {
		c.ConfUserExist = true
	}
}
