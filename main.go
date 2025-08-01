package main

import (
	"log"
	"net/url"
	"os"
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
	cfApiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	cfZoneId := os.Getenv("CLOUDFLARE_ZONE_ID")

	if cfApiToken != "" && cfZoneId != "" {
		log.Println("Using Cloudflare DNS service.")
		rootDomain := os.Getenv("ROOT_DOMAIN")
		if rootDomain == "" {
			log.Fatalf("ROOT_DOMAIN environment variable must be set when using Cloudflare")
		}

		dnsService, err = dns.NewCloudflareDns(cfApiToken, cfZoneId, rootDomain)
		if err != nil {
			log.Fatalf("failed to create Cloudflare DNS service: %v", err)
		}
	} else if nameServer := os.Getenv("NAME_SERVER"); nameServer != "" {
		log.Println("Using TSIG-based DNS service.")
		dnsProtocol := os.Getenv("DNS_PROTOCOL")
		tsigKey := os.Getenv("TSIG_KEY")
		tsigSecret := os.Getenv("TSIG_SECRET")
		if len(tsigKey) == 0 || len(tsigSecret) == 0 {
			log.Fatalf("TSIG_KEY and TSIG_SECRET must be set when using standard DNS")
		}
		dnsService = dns.NewDns(externalURL, nameServer, dnsProtocol, tsigKey, tsigSecret)
	} else {
		log.Println("No DNS service configured.")
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
	return url.Parse(serverURLStr)
}
