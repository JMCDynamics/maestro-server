package dtos

type ResponseNewPeer struct {
	VpnAddress   string `json:"vpnAddress"`
	ConfigOutput string `json:"-"`
}
