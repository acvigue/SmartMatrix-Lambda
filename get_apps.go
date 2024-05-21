package smartmatrixlambda

import (
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type App struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Author      string `json:"author"`
}

func GetApps(c *gin.Context) {
	apps := []string{}

	path := `./TidbytCommunity/apps`
	files, err := os.ReadDir(path)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	for _, file := range files {
		apps = append(apps, file.Name())
	}

	c.JSON(200, gin.H{"apps": apps})
}

func GetApp(c *gin.Context) {
	appID := c.Param("id")

	path := `./TidbytCommunity/apps/` + appID + `/manifest.yaml`

	dat, err := os.ReadFile(path)
	if err != nil {
		c.JSON(404, gin.H{"error": "app not found"})
		return
	}

	manifest := AppletManifest{}

	err = yaml.Unmarshal(dat, &manifest)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not unmarshal manifest"})
		return
	}

	app := App{
		ID:          manifest.ID,
		Name:        manifest.Name,
		Description: manifest.Description,
		Author:      manifest.Author,
	}

	c.JSON(200, app)
}
