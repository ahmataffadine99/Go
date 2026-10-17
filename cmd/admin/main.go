package main

import (
	"flag"
	"os"

	"ecommerce-cli/internal/app"
)

func main() {
	baseURL := flag.String("url", "http://localhost:8080", "adresse du serveur HTTP")
	flag.Parse()

	app.RunAdminApp(*baseURL, os.Stdin, os.Stdout)
}
