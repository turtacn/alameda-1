package factory

const (
	ApiPrefix    = "/api/v1"
	EpQuery      = "/query"
	EpQueryRange = "/query_range"
)

type QueryRequestFactoryOpt struct {
	PromAddr string
	PromAuth string
}

type QueryRequestFactoryOpts func(*QueryRequestFactoryOpt)

func PromAddr(addr string) QueryRequestFactoryOpts {
	return func(o *QueryRequestFactoryOpt) {
		o.PromAddr = addr
	}
}

func PromAuth(auth string) QueryRequestFactoryOpts {
	return func(o *QueryRequestFactoryOpt) {
		o.PromAuth = auth
	}
}
