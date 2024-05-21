package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"

	_ "github.com/joho/godotenv/autoload"
)

var ginEngine *gin.Engine
var ginLambda *ginadapter.GinLambda

func init() {
	r := gin.Default()

	r.POST("/schema", PostSchema)
	r.POST("/render", PostRender)
	r.GET("/apps", GetApps)
	r.GET("/apps/:id", GetApp)

	ginEngine = r
}

func main() {
	//if has environment variable, run as lambda
	if os.Getenv("_LAMBDA_SERVER_PORT") != "" {
		lambda.Start(Handler)
	} else {
		ginEngine.Run(":8080")
	}
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	ginadapter.New(ginEngine)
	return ginLambda.ProxyWithContext(ctx, req)
}
