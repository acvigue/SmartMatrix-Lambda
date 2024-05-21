package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// when no 'PORT' environment variable defined, run lambda
	if os.Getenv("PORT") == "" {
		lambda.Start(Handler)
		return
	}
	// otherwise start a local server
	StartLocal()
}
