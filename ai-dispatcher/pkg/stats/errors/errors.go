package errors

const (
	ErrorDataPointsNotEnough = "data points not enough"
	ErrorNoDataPoints        = "no data points"
)

func DataPointsNotEnough(err error) bool {
	return err.Error() == ErrorDataPointsNotEnough
}
