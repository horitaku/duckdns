package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	duckDNSURL = "https://www.duckdns.org/update"
)

// Config holds the configuration for DuckDNS updates
type Config struct {
	Domain string
	Token  string
	IP     string
	Verbose bool
}

func main() {
	config := parseFlags()

	if err := validateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	if err := updateDuckDNS(config); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to update DuckDNS: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("DuckDNS update successful")
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.Domain, "domain", os.Getenv("DUCKDNS_DOMAIN"), "DuckDNS domain (can also use DUCKDNS_DOMAIN env var)")
	flag.StringVar(&config.Token, "token", os.Getenv("DUCKDNS_TOKEN"), "DuckDNS token (can also use DUCKDNS_TOKEN env var)")
	flag.StringVar(&config.IP, "ip", "", "IP address to set (optional, uses your public IP if not specified)")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -domain=mydomain -token=mytoken\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -domain=mydomain -token=mytoken -ip=1.2.3.4\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  DUCKDNS_DOMAIN=mydomain DUCKDNS_TOKEN=mytoken %s\n", os.Args[0])
	}

	flag.Parse()

	return config
}

func validateConfig(config *Config) error {
	if config.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if config.Token == "" {
		return fmt.Errorf("token is required")
	}
	return nil
}

func updateDuckDNS(config *Config) error {
	// Build the URL
	url := fmt.Sprintf("%s?domains=%s&token=%s", duckDNSURL, config.Domain, config.Token)
	
	if config.IP != "" {
		url += fmt.Sprintf("&ip=%s", config.IP)
	}

	if config.Verbose {
		fmt.Printf("Sending request to DuckDNS...\n")
		if config.IP != "" {
			fmt.Printf("Setting IP to: %s\n", config.IP)
		} else {
			fmt.Printf("Using automatic IP detection\n")
		}
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Send the request
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	response := strings.TrimSpace(string(body))

	if config.Verbose {
		fmt.Printf("Response: %s\n", response)
	}

	// Check if the update was successful
	if response != "OK" {
		return fmt.Errorf("update failed with response: %s", response)
	}

	return nil
}
