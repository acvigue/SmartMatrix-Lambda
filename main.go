package smartmatrixlambda

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"tidbyt.dev/pixlet/runtime"
)

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

var ginEngine *gin.Engine
var ginLambda *ginadapter.GinLambda

func init() {
	r := gin.Default()

	r.POST("/schema", PostSchema)
	r.POST("/render", PostRender)
	r.GET("/apps", GetApps)
	r.GET("/app/:id", GetApp)

	ginEngine = r
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	ginadapter.New(ginEngine)
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	//if has environment variable, run as lambda
	if os.Getenv("_LAMBDA_SERVER_PORT") != "" {
		lambda.Start(Handler)
	} else {
		ginEngine.Run(":8080")
	}
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
		path := `./TidbytCommunity/apps/` + applet.AppletName
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
