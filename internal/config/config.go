package config

type NetAddress struct {
	ServerURL string
	BaseURL   string
}

func New() *NetAddress {
	return &NetAddress{ServerURL: "localhost:8080", BaseURL: "http://localhost:8080"}

}
