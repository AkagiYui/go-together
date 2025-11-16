package minecraft

import "encoding/json"

type MCBEInfo struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Motd            string `json:"motd"`
	ProtocolVersion int    `json:"protocolVersion"`
	ServerVersion   string `json:"serverVersion"`
	CurrentPlayers  int    `json:"currentPlayers"`
	MaxPlayers      int    `json:"maxPlayers"`
	UniqueID        string `json:"uniqueId"`
	WorldName       string `json:"worldName"`
	GameMode        string `json:"gameMode"`
	PortIPv4        int    `json:"portIpv4"`
	PortIPv6        int    `json:"portIpv6"`
	Delay           int    `json:"delay"`
}

type MCJEInfo struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Description     string `json:"description"`
	ProtocolVersion int    `json:"protocolVersion"`
	ServerVersion   string `json:"serverVersion"`
	CurrentPlayers  int    `json:"currentPlayers"`
	MaxPlayers      int    `json:"maxPlayers"`
	Delay           int    `json:"delay"`
}

type MCJEDescription struct {
	Text string `json:"text,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for MCJEDescription
func (d *MCJEDescription) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		d.Text = str
		return nil
	}

	// If string unmarshal fails, try struct format
	type DescriptionStruct struct {
		Text string `json:"text"`
	}
	var desc DescriptionStruct
	if err := json.Unmarshal(data, &desc); err != nil {
		return err
	}
	d.Text = desc.Text
	return nil
}

type MCJEResponse struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int `json:"max"`
		Online int `json:"online"`
	} `json:"players"`
	Description MCJEDescription `json:"description"`
}
