package main

import (
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/breez/breez-lnurl/cache"
	"github.com/breez/breez-lnurl/dns"
	"github.com/breez/breez-lnurl/persist"
)

func main() {
	storage, err := persist.NewPgStore(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to create postgres store: %v", err)
	}

	externalURL, err := parseURLFromEnv("SERVER_EXTERNAL_URL", "http://localhost:8080")
	if err != nil {
		log.Fatalf("failed to parse external server URL %v", err)
	}

	var dnsService dns.DnsService
	// Check for the new BOLT12_ENABLED flag.
	bolt12Enabled := os.Getenv("BOLT12_ENABLED") == "true"

	if bolt12Enabled {
		cfApiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
		cfZoneId := os.Getenv("CLOUDFLARE_ZONE_ID")

		if cfApiToken != "" && cfZoneId != "" {
			log.Println("Using Cloudflare DNS service for BOLT12.")
			rootDomain := os.Getenv("ROOT_DOMAIN")
			if rootDomain == "" {
				log.Fatalf("ROOT_DOMAIN environment variable must be set when using Cloudflare")
			}

			// Initialize the Cloudflare DNS service.
			dnsService, err = dns.NewCloudflareDns(cfApiToken, cfZoneId, rootDomain)
			if err != nil {
				log.Fatalf("failed to create Cloudflare DNS service: %v", err)
			}
		} else {
			log.Println("BOLT12 is enabled, but Cloudflare variables are not set. DNS records will not be created.")
			dnsService = dns.NewNoDns()
		}
	} else {
		log.Println("BOLT12 DNS record creation is disabled.")
		dnsService = dns.NewNoDns()
	}

	internalURL, err := parseURLFromEnv("SERVER_INTERNAL_URL", "http://localhost:8080")
	if err != nil {
		log.Fatalf("failed to parse internal server URL %v", err)
	}

	cacheService := cache.NewCache(time.Minute)

	NewServer(internalURL, externalURL, storage, dnsService, cacheService).Serve()
}

func parseURLFromEnv(envKey string, defaultURL string) (*url.URL, error) {
	serverURLStr := os.Getenv(envKey)
	if serverURLStr == "" {
		serverURLStr = defaultURL
	}

	if !strings.HasPrefix(serverURLStr, "http://") && !strings.HasPrefix(serverURLStr, "https://") {
		serverURLStr = "https://" + serverURLStr
	}

	return url.Parse(serverURLStr)
}
