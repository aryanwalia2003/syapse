
package domain

type WMSProvider struct {
	ID          string `json:"id"`
	Name        string `json:"name"` // e.g., EasyEcom
	IsSupported bool   `json:"is_supported"`
}

type DISProvider struct {
	ID          string `json:"id"`
	Name        string `json:"name"` // e.g., Loginext
	IsSupported bool   `json:"is_supported"`
}
