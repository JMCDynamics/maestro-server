package dtos

type OperatingSystem string

type TypeNodeStatus = string

const (
	WINDOWS OperatingSystem = "WINDOWS"
	LINUX   OperatingSystem = "LINUX"

	UP   TypeNodeStatus = "UP"
	DOWN TypeNodeStatus = "DOWN"
)

type NodeStatus struct {
	Id     string         `json:"id"`
	Status TypeNodeStatus `json:"status"`
}

type VpnInterface struct {
	Address    string `ini:"Address" json:"address"`
	PrivateKey string `ini:"PrivateKey" json:"privateKey"`
	DNS        string `ini:"DNS" json:"dns"`
}

type VpnPeer struct {
	PublicKey    string `ini:"PublicKey" json:"publicKey"`
	PresharedKey string `ini:"PresharedKey" json:"presharedKey"`
	AllowedIPs   string `ini:"AllowedIPs" json:"allowedIPs"`
	Endpoint     string `ini:"Endpoint" json:"endpoint"`
}

type VpnConfig struct {
	Interface VpnInterface `ini:"Interface" json:"interface"`
	Peer      VpnPeer      `ini:"Peer" json:"peer"`
}

type Node struct {
	Id              string          `json:"id"`
	Name            string          `json:"name"`
	VpnAddress      string          `json:"vpnAddress"`
	OperatingSystem OperatingSystem `json:"operatingSystem"`
	Status          TypeNodeStatus  `json:"status"`
	VpnConfig       string          `json:"vpnConfig"`
}
