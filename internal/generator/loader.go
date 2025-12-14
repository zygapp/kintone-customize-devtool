package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kintone/kcdev/internal/config"
	"github.com/kintone/kcdev/internal/prompt"
)

const (
	loaderSchemaVersion = 1
	devOrigin           = "https://localhost:3000"
)

type LoaderMeta struct {
	SchemaVersion int           `json:"schemaVersion"`
	KcdevVersion  string        `json:"kcdevVersion"`
	GeneratedAt   string        `json:"generatedAt"`
	Dev           DevMeta       `json:"dev"`
	Project       ProjectMeta   `json:"project"`
	Kintone       KintoneMeta   `json:"kintone"`
	Files         FilesMeta     `json:"files"`
}

type DevMeta struct {
	Origin string `json:"origin"`
	Entry  string `json:"entry"`
}

type ProjectMeta struct {
	Name      string `json:"name"`
	Framework string `json:"framework"`
	Language  string `json:"language"`
}

type KintoneMeta struct {
	Domain string `json:"domain"`
	AppID  int    `json:"appId"`
}

type FilesMeta struct {
	LoaderPath   string `json:"loaderPath"`
	LoaderSha256 string `json:"loaderSha256"`
	CertKeyPath  string `json:"certKeyPath"`
	CertCertPath string `json:"certCertPath"`
}

func GenerateLoader(projectDir string, answers *prompt.InitAnswers) error {
	managedDir := filepath.Join(projectDir, config.ConfigDir, "managed")
	if err := os.MkdirAll(managedDir, 0755); err != nil {
		return err
	}

	entry := GetEntryPath(answers.Framework, answers.Language)
	loaderContent := generateLoaderContent(entry)
	loaderPath := filepath.Join(managedDir, "kintone-dev-loader.js")

	if err := os.WriteFile(loaderPath, []byte(loaderContent), 0644); err != nil {
		return err
	}

	loaderHash := sha256.Sum256([]byte(loaderContent))

	meta := &LoaderMeta{
		SchemaVersion: loaderSchemaVersion,
		KcdevVersion:  "0.1.0",
		GeneratedAt:   time.Now().Format(time.RFC3339),
		Dev: DevMeta{
			Origin: devOrigin,
			Entry:  entry,
		},
		Project: ProjectMeta{
			Name:      answers.ProjectName,
			Framework: string(answers.Framework),
			Language:  string(answers.Language),
		},
		Kintone: KintoneMeta{
			Domain: answers.Domain,
			AppID:  answers.AppID,
		},
		Files: FilesMeta{
			LoaderPath:   ".kcdev/managed/kintone-dev-loader.js",
			LoaderSha256: hex.EncodeToString(loaderHash[:]),
			CertKeyPath:  ".kcdev/certs/localhost-key.pem",
			CertCertPath: ".kcdev/certs/localhost.pem",
		},
	}

	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	metaPath := filepath.Join(managedDir, "loader.meta.json")
	return os.WriteFile(metaPath, metaData, 0644)
}

func generateLoaderContent(entry string) string {
	now := time.Now().Format(time.RFC3339)

	return fmt.Sprintf(`// kcdev-loader
// schemaVersion: %d
// generatedAt: %s
// origin: %s

(() => {
  const origin = "%s";
  const t = Date.now();

  // 同期 XHR で IIFE バンドルを取得して実行
  const xhr = new XMLHttpRequest();
  xhr.open("GET", origin + "/customize.js?t=" + t, false);
  xhr.send();
  if (xhr.status === 200) {
    eval(xhr.responseText);
  }

  // HMR: @vite/client を非同期で読み込んでリロード検知
  import(origin + "/@vite/client").catch(() => {});
})();
`, loaderSchemaVersion, now, devOrigin, devOrigin)
}

func LoadLoaderMeta(projectDir string) (*LoaderMeta, error) {
	metaPath := filepath.Join(projectDir, config.ConfigDir, "managed", "loader.meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta LoaderMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func VerifyLoader(projectDir string) (bool, string, error) {
	meta, err := LoadLoaderMeta(projectDir)
	if err != nil {
		return false, "メタデータが見つかりません", nil
	}

	loaderPath := filepath.Join(projectDir, config.ConfigDir, "managed", "kintone-dev-loader.js")
	content, err := os.ReadFile(loaderPath)
	if err != nil {
		return false, "ローダーファイルが見つかりません", nil
	}

	hash := sha256.Sum256(content)
	if hex.EncodeToString(hash[:]) != meta.Files.LoaderSha256 {
		return false, "ローダーが変更されています。再登録が必要です", nil
	}

	return true, "OK（再登録不要）", nil
}
