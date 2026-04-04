package partition

import "github.com/tidwall/gjson"

// ExtractVendorOrderID extracts the order ID or AWB from the JSON payload
// using the given gjson path. Returns empty string if not found or invalid.
func ExtractVendorOrderID(payload []byte, path string) string {
	if !gjson.ValidBytes(payload) {
		return ""
	}
	result := gjson.GetBytes(payload, path)
	if !result.Exists() {
		return ""
	}
	return result.String()
}
