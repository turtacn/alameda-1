package metrics

func IsLabelNeedTransform(src, dest string) bool {
	return src != dest
}
