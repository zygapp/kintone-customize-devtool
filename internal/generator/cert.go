package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kintone/kcdev/internal/config"
)

func GenerateCerts(projectDir string) error {
	certsDir := filepath.Join(projectDir, config.ConfigDir, "certs")
	if err := os.MkdirAll(certsDir, 0755); err != nil {
		return err
	}

	keyPath := filepath.Join(certsDir, "localhost-key.pem")
	certPath := filepath.Join(certsDir, "localhost.pem")

	if _, err := os.Stat(keyPath); err == nil {
		return nil
	}

	opensslConf := `[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = localhost

[v3_req]
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
IP.2 = ::1
`

	confPath := filepath.Join(certsDir, "openssl.cnf")
	if err := os.WriteFile(confPath, []byte(opensslConf), 0644); err != nil {
		return err
	}
	defer os.Remove(confPath)

	cmd := exec.Command("openssl", "req",
		"-x509",
		"-newkey", "rsa:2048",
		"-keyout", keyPath,
		"-out", certPath,
		"-days", "365",
		"-nodes",
		"-config", confPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("証明書生成エラー: %w\n%s", err, string(output))
	}

	return nil
}

func CertsExist(projectDir string) bool {
	certsDir := filepath.Join(projectDir, config.ConfigDir, "certs")
	keyPath := filepath.Join(certsDir, "localhost-key.pem")
	certPath := filepath.Join(certsDir, "localhost.pem")

	if _, err := os.Stat(keyPath); err != nil {
		return false
	}
	if _, err := os.Stat(certPath); err != nil {
		return false
	}
	return true
}
