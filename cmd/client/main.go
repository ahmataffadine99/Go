package main

import (
	"os"

	"ecommerce-cli/internal/app"
)

func main() {
	baseURL := "http://localhost:8080"
	app.RunClientApp(baseURL, os.Stdin, os.Stdout)
}
