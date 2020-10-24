package main

type ConfCommCtr struct {
	ApiKey            string             `json:"api_key"`
	Server            string             `json:"server"`
	Port              string             `json:"port"`
	Ssl               bool               `json:"ssl"`
	Cert              ConfCertificate    `json:"cert,omitempty"`
	ClientRoot        string             `json:"client_root"`
	AuthorizedDevices []AuthorizedDevice `json:"authorized_devices"`
	CorsOrigin        bool               `json:"cors_origin"`
}

type AuthorizedDevice struct {
	ApiKey   string `json:"api_key"`
	Name     string `json:"name"`
	MacAddr  string `json:"mac_addr"`
	IsOnline bool   `json:"is_online"`
	Enabled  bool   `json:"enabled"`
}

type ConfCertificate struct {
	SslKey  string `json:"ssl_key"`
	SslCert string `json:"ssl_cert,"`
}
