package domain

type WMSIntegration struct {
	ID             string `json:"id"`
	BrandID        string `json:"brand_id"`
	WMSProviderID  string `json:"wms_provider_id"`
	APICredentials []byte `json:"api_credentials"` // Encrypted JSON
	IsActive       bool   `json:"is_active"`
}

type DISIntegration struct {
	ID             string `json:"id"`
	BrandID        string `json:"brand_id"`
	DISProviderID  string `json:"dis_provider_id"`
	APICredentials []byte `json:"api_credentials"` // Encrypted JSON
	WebhookSecret  string `json:"webhook_secret"`
	IsActive       bool   `json:"is_active"`
}

