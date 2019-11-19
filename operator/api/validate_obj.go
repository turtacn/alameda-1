package api

type ValidatingObject struct {
	Namespace           string
	Name                string
	Kind                string
	Labels              map[string]string
	SelectorMatchLabels map[string]string
}
