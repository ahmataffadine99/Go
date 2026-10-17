package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"ecommerce-cli/internal/app"
)

const version = "1.0.0"

func main() {
	baseURL := flag.String("url", "http://localhost:8080", "adresse du serveur HTTP")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: client [-url adresse] [interactive|status|version]")
		flag.PrintDefaults()
	}
	flag.Parse()

	switch flag.Arg(0) {
	case "", "interactive":
		app.RunClientApp(*baseURL, os.Stdin, os.Stdout)
	case "status":
		checkStatus(*baseURL)
	case "version":
		fmt.Println("E-Commerce CLI Client v" + version)
	default:
		flag.Usage()
		os.Exit(1)
	}
}

func checkStatus(baseURL string) {
	fmt.Println("Vérification du serveur Backend...")

	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(baseURL + "/api/products")
	if err != nil {
		fmt.Printf("Serveur injoignable: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Serveur en ligne et opérationnel.")
	} else {
		fmt.Printf("Serveur en ligne mais statut inattendu: %d\n", resp.StatusCode)
	}
}
