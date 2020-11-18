package main

type ConfCommCtr struct {
	ApiKey            string             `json:"api_key"`
	Server            string             `json:"server"`
	Port              string             `json:"port"`
	Ssl               bool               `json:"ssl"`
	Cert              ConfCertificate    `json:"cert,omitempty"`
	TelegramConf      string             `json:"telegram_conf"`
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

type TmeConf struct {
	BotId  string `json:"bot_id"`
	ChatId int64  `json:"chat_id"`
}

type TmeMessage struct {
	ChatId              int64  `json:"chat_id"`
	Text                string `json:"text"`
	DisableNotification bool   `json:"disable_notification"`
}
