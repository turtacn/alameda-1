package objects

//Role: super, domain-admin, user, agent
type UserInfo struct {
	Name       string
	DomainName string
	Role       string
	Namespace  string
	Token      string

	Password    string
	Cookie      string
	CookieValue string
}
