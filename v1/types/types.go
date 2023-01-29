package types

type RedisConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	DB int "json:db"
	Password string "json:password"
}

type ConfigFile struct {
	ServerBaseUrl string `json:"server_base_url"`
	ServerPort string `json:"server_port"`
	ServerAPIKey string `json:"server_api_key"`
	TimeZone string `json:"time_zone"`
	Redis RedisConfig "json:latitude"
	IPBlacklist []string "json:ip_blacklist"
	IPInfoToken string `json:"ip_info_token"`
}

type AListResponse struct {
	UUIDS []string `json:"uuids"`
}

type RedisMultiCommand struct {
	Command string `json:"type"`
	Key string `json:"key"`
	Args string `json:"args"`
}