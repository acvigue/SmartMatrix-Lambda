package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"tidbyt.dev/pixlet/encode"
	"tidbyt.dev/pixlet/globals"
	"tidbyt.dev/pixlet/runtime"
)

type PostRenderParams struct {
	Source  AppletSource      `json:"appletSource"`
	Params  map[string]string `json:"params"`
	Width   int               `json:"width"`
	Height  int               `json:"height"`
	Magnify int               `json:"magnify"`
}

func PostRenderTidbyt(params PostRenderParams) ([]byte, error) {
	authToken, err := GetTidbytRendererToken()
	if err != nil {
		return nil, err
	}

	//turn the params into a string
	confArr := []string{}
	for k, v := range params.Params {
		confArr = append(confArr, fmt.Sprintf("%s=%s", k, v))
	}
	confString := strings.Join(confArr, "&")

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "https://prod.tidbyt.com/app-server/preview/%s.webp?v=%d&%s", params.Source.AppletName, time.Now().Unix(), confString)
	url := buf.String()

	fmt.Printf("URL: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("tidbyt renderer returned status code %d", resp.StatusCode)
	}

	//return response as bytes
	buf.Reset()
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func PostRender(c *gin.Context) {
	var params PostRenderParams
	if err := c.BindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	if params.Source.Type == AppletSourceTypeTidbyt {
		buf, err := PostRenderTidbyt(params)

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.Data(200, "image/webp", buf)
		return
	}

	globals.Width = params.Width
	globals.Height = params.Height

	app, err := CreateApplet(params.Source)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	cache := runtime.NewInMemoryCache()
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)

	roots, err := app.RunWithConfig(ctx, params.Params)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	Filter := func(input image.Image) (image.Image, error) {
		magnify := params.Magnify
		if magnify <= 1 {
			return input, nil
		}
		in, ok := input.(*image.RGBA)
		if !ok {
			return nil, fmt.Errorf("image not RGBA, very weird")
		}

		out := image.NewRGBA(
			image.Rect(
				0, 0,
				in.Bounds().Dx()*magnify,
				in.Bounds().Dy()*magnify),
		)
		for x := 0; x < in.Bounds().Dx(); x++ {
			for y := 0; y < in.Bounds().Dy(); y++ {
				for xx := 0; xx < magnify; xx++ {
					for yy := 0; yy < magnify; yy++ {
						out.SetRGBA(
							x*magnify+xx,
							y*magnify+yy,
							in.RGBAAt(x, y),
						)
					}
				}
			}
		}

		return out, nil
	}

	screens := encode.ScreensFromRoots(roots)
	buf, err := screens.EncodeWebP(15000, Filter)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Data(200, "image/webp", buf)
}
