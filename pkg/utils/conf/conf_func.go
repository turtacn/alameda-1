package conf

var pConfig *ProConf = nil

func ConfigInit(file string) {
	if pConfig == nil {
		pConfig = NewConf(file)
	}
}

func Get(key string, defaultValue interface{}) interface{} {
	ret := pConfig.Get(key, defaultValue)
	return ret
}

func Set(key string, value interface{}) error {
	ret := pConfig.Set(key, value)
	return ret
}

func Refresh() error {
	ret := pConfig.Refresh()
	return ret
}
