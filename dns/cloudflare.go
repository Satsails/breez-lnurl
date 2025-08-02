package dns

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/cloudflare/cloudflare-go"
)

// CloudflareDnsService implements the DnsService interface using the official Go library.
type CloudflareDnsService struct {
	api          *cloudflare.API
	zoneID       string
	targetDomain string
}

// Updated function to accept the externalURL
func NewCloudflareDns(apiToken, zoneID, rootDomain string, externalURL *url.URL) (DnsService, error) {
	if rootDomain == "" {
		return nil, fmt.Errorf("rootDomain cannot be empty")
	}
	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudflare client: %w", err)
	}

	return &CloudflareDnsService{
		api:          api,
		zoneID:       zoneID,
		targetDomain: externalURL.Hostname(),
	}, nil
}

func (c *CloudflareDnsService) Set(username, offer string) (uint32, error) {
	ctx := context.Background()
	recordName := fmt.Sprintf("%s.user._bitcoin-payment.%s", username, c.targetDomain)
	log.Printf("Setting Cloudflare DNS TXT record for: %s", recordName)

	rc := cloudflare.ZoneIdentifier(c.zoneID)

	records, _, err := c.api.ListDNSRecords(ctx, rc, cloudflare.ListDNSRecordsParams{
		Type: "TXT",
		Name: recordName,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list existing DNS records: %w", err)
	}

	fullContent := fmt.Sprintf("bitcoin:?lno=%s", offer)
	ttl := 3600

	if len(records) > 0 {
		recordID := records[0].ID
		params := cloudflare.UpdateDNSRecordParams{
			ID:      recordID,
			Content: fullContent,
			TTL:     ttl,
		}
		_, err := c.api.UpdateDNSRecord(ctx, rc, params)
		if err != nil {
			return 0, fmt.Errorf("failed to update DNS record %s: %w", recordID, err)
		}
		log.Printf("Successfully updated DNS record %s", recordName)
	} else {
		params := cloudflare.CreateDNSRecordParams{
			Type:    "TXT",
			Name:    recordName,
			Content: fullContent,
			TTL:     ttl,
		}
		_, err := c.api.CreateDNSRecord(ctx, rc, params)
		if err != nil {
			return 0, fmt.Errorf("failed to create DNS record for %s: %w", err)
		}
		log.Printf("Successfully created DNS record %s", recordName)
	}

	return uint32(ttl), nil
}

func (c *CloudflareDnsService) Remove(username string) error {
	ctx := context.Background()
	recordName := fmt.Sprintf("%s.user._bitcoin-payment.%s", username, c.targetDomain)
	log.Printf("Removing Cloudflare DNS TXT record for: %s", recordName)

	rc := cloudflare.ZoneIdentifier(c.zoneID)

	records, _, err := c.api.ListDNSRecords(ctx, rc, cloudflare.ListDNSRecordsParams{
		Type: "TXT",
		Name: recordName,
	})
	if err != nil {
		return fmt.Errorf("failed to list DNS records for removal: %w", err)
	}

	if len(records) == 0 {
		log.Printf("No record found for %s, nothing to remove.", recordName)
		return nil // Success, as the record doesn't exist.
	}

	recordID := records[0].ID
	err = c.api.DeleteDNSRecord(ctx, rc, recordID)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record %s: %w", recordID, err)
	}

	log.Printf("Successfully removed DNS record %s", recordName)
	return nil
}
