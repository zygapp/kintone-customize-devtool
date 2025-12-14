package kintone

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	domain   string
	username string
	password string
	client   *http.Client
}

func NewClient(domain, username, password string) *Client {
	return &Client{
		domain:   domain,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) baseURL() string {
	return fmt.Sprintf("https://%s", c.domain)
}

func (c *Client) authHeader() string {
	auth := base64.StdEncoding.EncodeToString([]byte(c.username + ":" + c.password))
	return auth
}

type FileUploadResponse struct {
	FileKey string `json:"fileKey"`
}

func (c *Client) UploadFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.baseURL()+"/k/v1/file.json", body)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-Cybozu-Authorization", c.authHeader())
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ファイルアップロードエラー: %s - %s", resp.Status, string(respBody))
	}

	var result FileUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.FileKey, nil
}

type CustomizeScope string

const (
	ScopeAll    CustomizeScope = "ALL"
	ScopeAdmin  CustomizeScope = "ADMIN"
	ScopeNone   CustomizeScope = "NONE"
)

type FileCustomization struct {
	Type string `json:"type"`
	File *File  `json:"file,omitempty"`
	URL  string `json:"url,omitempty"`
}

type File struct {
	FileKey string `json:"fileKey"`
}

type CustomizeRequest struct {
	App     int                          `json:"app"`
	Scope   CustomizeScope               `json:"scope"`
	Desktop *CustomizeDesktopMobile      `json:"desktop,omitempty"`
	Mobile  *CustomizeDesktopMobile      `json:"mobile,omitempty"`
}

type CustomizeDesktopMobile struct {
	JS  []FileCustomization `json:"js"`
	CSS []FileCustomization `json:"css"`
}

type CustomizeFiles struct {
	JSFileKey  string
	CSSFileKey string
}

func (c *Client) UpdateCustomize(appID int, desktopFiles, mobileFiles *CustomizeFiles) error {
	customize := CustomizeRequest{
		App:   appID,
		Scope: ScopeAll,
	}

	// Desktop
	if desktopFiles != nil {
		jsFiles := []FileCustomization{}
		cssFiles := []FileCustomization{}

		if desktopFiles.JSFileKey != "" {
			jsFiles = []FileCustomization{
				{
					Type: "FILE",
					File: &File{FileKey: desktopFiles.JSFileKey},
				},
			}
		}

		if desktopFiles.CSSFileKey != "" {
			cssFiles = []FileCustomization{
				{
					Type: "FILE",
					File: &File{FileKey: desktopFiles.CSSFileKey},
				},
			}
		}

		customize.Desktop = &CustomizeDesktopMobile{
			JS:  jsFiles,
			CSS: cssFiles,
		}
	} else {
		customize.Desktop = &CustomizeDesktopMobile{
			JS:  []FileCustomization{},
			CSS: []FileCustomization{},
		}
	}

	// Mobile
	if mobileFiles != nil {
		jsFiles := []FileCustomization{}
		cssFiles := []FileCustomization{}

		if mobileFiles.JSFileKey != "" {
			jsFiles = []FileCustomization{
				{
					Type: "FILE",
					File: &File{FileKey: mobileFiles.JSFileKey},
				},
			}
		}

		if mobileFiles.CSSFileKey != "" {
			cssFiles = []FileCustomization{
				{
					Type: "FILE",
					File: &File{FileKey: mobileFiles.CSSFileKey},
				},
			}
		}

		customize.Mobile = &CustomizeDesktopMobile{
			JS:  jsFiles,
			CSS: cssFiles,
		}
	} else {
		customize.Mobile = &CustomizeDesktopMobile{
			JS:  []FileCustomization{},
			CSS: []FileCustomization{},
		}
	}

	return c.updateCustomizeRequest(customize)
}

func (c *Client) UpdateCustomizeWithURL(appID int, jsURLs []string, targetDesktop, targetMobile bool) error {
	js := make([]FileCustomization, len(jsURLs))
	for i, url := range jsURLs {
		js[i] = FileCustomization{
			Type: "URL",
			URL:  url,
		}
	}

	customize := CustomizeRequest{
		App:   appID,
		Scope: ScopeAll,
	}

	if targetDesktop {
		customize.Desktop = &CustomizeDesktopMobile{
			JS:  js,
			CSS: []FileCustomization{},
		}
	} else {
		customize.Desktop = &CustomizeDesktopMobile{
			JS:  []FileCustomization{},
			CSS: []FileCustomization{},
		}
	}

	if targetMobile {
		customize.Mobile = &CustomizeDesktopMobile{
			JS:  js,
			CSS: []FileCustomization{},
		}
	} else {
		customize.Mobile = &CustomizeDesktopMobile{
			JS:  []FileCustomization{},
			CSS: []FileCustomization{},
		}
	}

	return c.updateCustomizeRequest(customize)
}

func (c *Client) updateCustomizeRequest(customize CustomizeRequest) error {

	body, err := json.Marshal(customize)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.baseURL()+"/k/v1/preview/app/customize.json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("X-Cybozu-Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("カスタマイズ更新エラー: %s - %s", resp.Status, string(respBody))
	}

	return nil
}

type DeployRequest struct {
	Apps []DeployApp `json:"apps"`
}

type DeployApp struct {
	App int `json:"app"`
}

func (c *Client) DeployApp(appID int) error {
	deployReq := DeployRequest{
		Apps: []DeployApp{{App: appID}},
	}

	body, err := json.Marshal(deployReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL()+"/k/v1/preview/app/deploy.json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("X-Cybozu-Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("デプロイ開始エラー: %s - %s", resp.Status, string(respBody))
	}

	return nil
}

type DeployStatusResponse struct {
	Apps []DeployStatusApp `json:"apps"`
}

type DeployStatusApp struct {
	App    string `json:"app"`
	Status string `json:"status"`
}

func (c *Client) WaitForDeploy(appID int) error {
	for i := 0; i < 60; i++ {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/k/v1/preview/app/deploy.json?apps=%d", c.baseURL(), appID), nil)
		if err != nil {
			return err
		}

		req.Header.Set("X-Cybozu-Authorization", c.authHeader())

		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}

		var status DeployStatusResponse
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			resp.Body.Close()
			return err
		}
		resp.Body.Close()

		if len(status.Apps) > 0 && status.Apps[0].Status == "SUCCESS" {
			return nil
		}

		if len(status.Apps) > 0 && status.Apps[0].Status == "FAIL" {
			return fmt.Errorf("デプロイ失敗")
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("デプロイタイムアウト")
}
