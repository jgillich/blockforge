package stratum

type Protocol string

var (
	ProtocolStandard Protocol = "standard"
	// Implements the NiceHash stratum protocol
	// https://github.com/nicehash/Specifications/blob/master/EthereumStratum_NiceHash_v1.0.0.txt
	ProtocolNicehash Protocol = "nicehash"
)

type Pool struct {
	URL      string   `hcl:"url" json:"url"`
	User     string   `hcl:"user" json:"user"`
	Pass     string   `hcl:"pass" json:"pass"`
	Email    string   `hcl:"email" json:"email"`
	Protocol Protocol `hcl:"protocol" json:"protocol"`
}
