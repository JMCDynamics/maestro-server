package adapters

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JMCDynamics/maestro-server/internal/dtos"
	"github.com/JMCDynamics/maestro-server/internal/interfaces"
)

type Wireguard struct {
	endpoint string
}

const (
	path_to_conf             string = "/config/wg_confs/wg0.conf"
	path_to_publickey_server string = "/config/server/publickey-server"
	path_to_peers            string = "/config"
)

func NewWireguardAdapter(endpoint string) interfaces.IVpnGateway {
	return &Wireguard{
		endpoint: endpoint,
	}
}

func (w *Wireguard) Run() error {
	cmd := exec.Command("wg-quick", "up", "/config/wg_confs/wg0.conf")
	return cmd.Run()
}

func (w *Wireguard) GenerateNewPeer(name string) (dtos.ResponseNewPeer, error) {
	nextAddress, err := getNextAddress()
	if err != nil {
		return dtos.ResponseNewPeer{}, err
	}

	peerPath := fmt.Sprintf("%s/peer_%s", path_to_peers, name)
	if err := os.MkdirAll(peerPath, os.ModePerm); err != nil {
		return dtos.ResponseNewPeer{}, fmt.Errorf("unable to create peer's folder: %v", err)
	}

	privateKey, publicKey, presharedKey, err := generateKeys(name)
	if err != nil {
		return dtos.ResponseNewPeer{}, err
	}

	config, err := generatePeerConf(name, nextAddress, privateKey, presharedKey, w.endpoint)
	if err != nil {
		return dtos.ResponseNewPeer{}, err
	}

	if err := addPeerToServer(name, publicKey, presharedKey, nextAddress); err != nil {
		return dtos.ResponseNewPeer{}, err
	}

	return dtos.ResponseNewPeer{
		VpnAddress:   nextAddress,
		ConfigOutput: config,
	}, nil
}

func getNextAddress() (string, error) {
	file, err := os.Open(path_to_conf)
	if err != nil {
		return "", errors.New("unable to open the wg0 conf file")
	}
	defer file.Close()

	var lastIP string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "AllowedIPs") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				ipPart := strings.Split(parts[2], "/")[0]
				lastIP = ipPart
			}
		}
	}

	if lastIP == "" {
		return "10.0.0.2", nil
	}

	parts := strings.Split(lastIP, ".")
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid IP format")
	}

	lastNumber, err := strconv.Atoi(parts[3])
	if err != nil {
		return "", fmt.Errorf("invalid IP format")
	}

	parts[3] = fmt.Sprintf("%d", lastNumber+1)
	return strings.Join(parts, "."), nil
}

func getServerPublicKey() (string, error) {
	key, err := os.ReadFile(path_to_publickey_server)
	if err != nil {
		return "", errors.New("unable to read server's public key")
	}

	return strings.TrimSpace(string(key)), nil
}

func generateKeys(peerName string) (string, string, string, error) {
	privateKeyCmd := exec.Command("wg", "genkey")
	privateKey, err := privateKeyCmd.Output()
	if err != nil {
		return "", "", "", err
	}
	privateKeyStr := strings.TrimSpace(string(privateKey))
	privateKeyPath := fmt.Sprintf("/config/peer_%s/privatekey-peer_%s", peerName, peerName)
	if err := os.WriteFile(privateKeyPath, []byte(privateKeyStr+""), 0600); err != nil {
		return "", "", "", fmt.Errorf("failed to save private key: %v", err)
	}

	publicKeyCmd := exec.Command("wg", "pubkey")
	publicKeyCmd.Stdin = strings.NewReader(privateKeyStr)
	publicKey, err := publicKeyCmd.Output()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate public key: %v", err)
	}
	publicKeyStr := strings.TrimSpace(string(publicKey))
	publicKeyPath := fmt.Sprintf("/config/peer_%s/publickey-peer_%s", peerName, peerName)
	if err := os.WriteFile(publicKeyPath, []byte(publicKeyStr+""), 0600); err != nil {
		return "", "", "", fmt.Errorf("failed to save public key: %v", err)
	}

	presharedKeyCmd := exec.Command("wg", "genkey")
	presharedKey, err := presharedKeyCmd.Output()
	if err != nil {
		return "", "", "", err
	}
	presharedKeyStr := strings.TrimSpace(string(presharedKey))
	presharedKeyPath := fmt.Sprintf("/config/peer_%s/presharedkey-peer_%s", peerName, peerName)
	if err := os.WriteFile(presharedKeyPath, []byte(presharedKeyStr+""), 0600); err != nil {
		return "", "", "", fmt.Errorf("failed to save preshared key: %v", err)
	}

	return privateKeyStr, publicKeyStr, presharedKeyStr, nil
}

func generatePeerConf(name, nextAddress, privateKey, presharedKey, endpoint string) (string, error) {
	serverPublicKey, err := getServerPublicKey()
	if err != nil {
		return "", err
	}

	peerPath := fmt.Sprintf("%s/peer_%s", path_to_peers, name)
	absPath, err := filepath.Abs(peerPath)
	if err != nil {
		return "", fmt.Errorf("error resolving absolute path: %v", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create base directory: %v", err)
	}

	file, err := os.Create(fmt.Sprintf("%s/peer_%s.conf", peerPath, name))
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	config := fmt.Sprintf(`[Interface]
Address = %s
PrivateKey = %s
ListenPort = 51820
DNS = 10.10.0.1

[Peer]
PublicKey = %s
PresharedKey = %s
AllowedIPs = 10.10.0.1/24
Endpoint = %s`,
		nextAddress,
		privateKey,
		serverPublicKey,
		presharedKey,
		strings.ReplaceAll(
			strings.ReplaceAll(endpoint, `“`, ""),
			`”`, "",
		),
	)

	file.WriteString(config)

	return config, nil
}

func addPeerToServer(peerName, publicKey, presharedKey, peerAddress string) error {
	peerConfig := fmt.Sprintf("\n[Peer]\n# peer_%s\nPublicKey = %s\nPresharedKey = %s\nAllowedIPs = %s/32",
		peerName,
		publicKey,
		presharedKey,
		peerAddress,
	)

	f, err := os.OpenFile(path_to_conf, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open wg conf: %v", err)
	}
	defer f.Close()

	if _, err = f.WriteString(peerConfig); err != nil {
		return fmt.Errorf("erro ao escrever no arquivo: %v", err)
	}

	presharedKeyPath := fmt.Sprintf("%s/peer_%s/presharedkey-peer_%s", path_to_peers, peerName, peerName)
	cmd := exec.Command("wg", "set", "wg0", "peer", publicKey, "preshared-key", presharedKeyPath, "allowed-ips", peerAddress+"/32")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to add new peer: %v\noutput: %s", err, string(output))
	}

	cmd = exec.Command("ip", "-4", "route", "add", peerAddress+"/32", "dev", "wg0")

	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to add new peer: %v\noutput: %s", err, string(output))
	}

	cmd = exec.Command("iptables", "-A", "FORWARD", "-i", "wg0", "-j", "ACCEPT")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro ao adicionar regra FORWARD de entrada: %v\noutput: %s", err, string(output))
	}

	cmd = exec.Command("iptables", "-A", "FORWARD", "-o", "wg0", "-j", "ACCEPT")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro ao adicionar regra FORWARD de saída: %v\noutput: %s", err, string(output))
	}

	cmd = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", "eth+", "-j", "MASQUERADE")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro ao adicionar regra de NAT: %v\noutput: %s", err, string(output))
	}

	return nil
}
