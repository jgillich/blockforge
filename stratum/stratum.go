package stratum

type Protocol string

var (
	ProtocolJsonrpc Protocol = "jsonrpc"
	// Implements the NiceHash stratum protocol
	// https://github.com/nicehash/Specifications/blob/master/EthereumStratum_NiceHash_v1.0.0.txt
	ProtocolNicehash Protocol = "nicehash"
)

type Pool struct {
	URL      string   `yaml:"url" json:"url"`
	User     string   `yaml:"user" json:"user"`
	Pass     string   `yaml:"pass" json:"pass"`
	Email    string   `yaml:"email,omitempty" json:"email"`
	Protocol Protocol `yaml:"protocol,omitempty" json:"protocol"`
}
