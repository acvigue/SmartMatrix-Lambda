package main

import "github.com/gin-gonic/gin"

type PostSchemaParams struct {
	Source AppletSource `json:"appletSource"`
}

func PostSchema(c *gin.Context) {
	var params PostSchemaParams
	if err := c.BindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	app, err := CreateApplet(params.Source)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.String(200, "%s", string(app.SchemaJSON))
}
