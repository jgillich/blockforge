package stratum

type Pool struct {
	URL   string `hcl:"url" json:"url"`
	User  string `hcl:"user" json:"user"`
	Pass  string `hcl:"pass" json:"pass"`
	Email string `hcl:"email" json:"email"`
}
