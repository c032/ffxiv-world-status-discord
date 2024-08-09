package ffxivapi

type WorldsResponse struct {
	Worlds []World `json:"worlds"`
}

type World struct {
	Group                  string `json:"group"`
	Name                   string `json:"name"`
	Category               string `json:"category"`
	ServerStatus           string `json:"serverStatus"`
	CanCreateNewCharacters bool   `json:"canCreateNewCharacters"`
	IsOnline               bool   `json:"isOnline"`
	IsMaintenance          bool   `json:"isMaintenance"`
	IsCongested            bool   `json:"isCongested"`
	IsPreferred            bool   `json:"isPreferred"`
	IsNew                  bool   `json:"isNew"`
}
