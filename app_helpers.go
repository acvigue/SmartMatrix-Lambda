package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"tidbyt.dev/pixlet/runtime"
)

type AppletManifest struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Summary     string `yaml:"summary" json:"summary"`
	Description string `yaml:"desc" json:"desc"`
	Author      string `yaml:"author" json:"author"`
	FileName    string `yaml:"fileName"`
	PackageName string `yaml:"packageName"`
}

type TidbytTokenFile struct {
	Token      string `yaml:"token"`
	ExpiryTime int64  `yaml:"expiryTime"`
}

type AppletSourceType string

const (
	AppletSourceTypeExternal AppletSourceType = "external"
	AppletSourceTypeTidbyt   AppletSourceType = "tidbyt"
	AppletSourceTypeInternal AppletSourceType = "internal"
)

type AppletSource struct {
	Type       AppletSourceType `json:"type"`
	AppletName string           `json:"appletName"`
	AppletURL  string           `json:"appletURL"`
}

func GetManifestForApp(appID string) (*AppletManifest, error) {
	AppsPath := os.Getenv("APPS_PATH")
	path := AppsPath + appID + `/manifest.yaml`

	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	manifest := AppletManifest{}

	err = yaml.Unmarshal(dat, &manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

func CreateApplet(applet AppletSource) (*runtime.Applet, error) {
	//if is url, download the source

	if applet.Type == AppletSourceTypeExternal {
		req, err := http.NewRequest(http.MethodGet, applet.AppletURL, nil)
		if err != nil {
			return nil, fmt.Errorf("could not create request to download applet source")
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("could not download applet source")
		}

		if res.StatusCode != 200 {
			fmt.Printf("server returned bad response code: %s\n", res.Status)
			os.Exit(1)
		}

		src, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read applet source")
		}

		app, err := runtime.NewApplet("applet", src)
		if err != nil {
			return nil, fmt.Errorf("could not create applet")
		}
		return app, nil
	} else if applet.Type == AppletSourceTypeInternal {
		AppsPath := os.Getenv("APPS_PATH")
		path := AppsPath + applet.AppletName
		_, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("applet not found: %s", path)
		}

		fs := os.DirFS(path)
		applet, err := runtime.NewAppletFromFS(filepath.Base(path), fs)
		if err != nil {
			return nil, fmt.Errorf("could not create applet")
		}
		return applet, nil
	}

	return nil, fmt.Errorf("invalid applet source type")
}

func GetTidbytRendererToken() (string, error) {
	token := TidbytTokenFile{}

	dat, err := os.ReadFile("/tmp/tidbyt-renderer-token")
	if err == nil {
		err = yaml.Unmarshal(dat, &token)
	}

	if err != nil || token.ExpiryTime < time.Now().Unix() {
		//fetch new token
		refreshToken := os.Getenv("TIDBYT_REFRESH_TOKEN")
		apiKey := os.Getenv("TIDBYT_API_KEY")

		url := `https://securetoken.googleapis.com/v1/token?key=` + apiKey
		body := `grant_type=refresh_token&refresh_token=` + refreshToken

		req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
		if err != nil {
			return "", err
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}

		if res.StatusCode != 200 {
			return "", err
		}

		//parse response as json
		decoder := json.NewDecoder(res.Body)
		var t map[string]interface{}
		err = decoder.Decode(&t)

		if err != nil {
			return "", err
		}

		token.Token = t["access_token"].(string)
		token.ExpiryTime = time.Now().Unix() + 3600

		//store token
		dat, err := yaml.Marshal(token)
		if err != nil {
			return "", err
		}

		err = os.WriteFile("/tmp/tidbyt-renderer-token", dat, 0644)
		if err != nil {
			return "", err
		}
	}

	return token.Token, nil
}
