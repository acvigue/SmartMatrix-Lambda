package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

func GetApps(c *gin.Context) {
	apps := map[string]AppletManifest{}

	path := `./TidbytCommunity/apps`
	files, err := os.ReadDir(path)

	if err != nil {
		c.JSON(404, gin.H{"error": "No apps found"})
		return
	}

	for _, file := range files {
		manifest, err := GetManifestForApp(file.Name())
		if err != nil {
			continue
		}

		apps[file.Name()] = *manifest
	}

	c.JSON(200, gin.H{"apps": apps})
}

func GetApp(c *gin.Context) {
	appID := c.Param("id")

	manifest, err := GetManifestForApp(appID)
	if err != nil {
		c.JSON(404, gin.H{"error": "App not found"})
		return
	}

	c.JSON(200, manifest)
}
