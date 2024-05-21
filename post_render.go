package smartmatrixlambda

import (
	"context"
	"fmt"
	"image"

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

func PostRender(c *gin.Context) {
	var params PostRenderParams
	if err := c.BindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
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
