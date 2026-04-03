package loginext

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/errors"
)

const ProviderNameLoginext = "LOGINEXT"

// LoginextNormalizer implements domain.DISNormalizer for the Loginext DIS.
// All Loginext-specific field names and status codes are sealed inside this file.
type LoginextNormalizer struct{}

// NewLoginextNormalizer creates a ready-to-use Loginext DIS adapter.
func NewLoginextNormalizer() domain.DISNormalizer {
	return &LoginextNormalizer{}
}

func (l *LoginextNormalizer) ProviderName() string {
	return ProviderNameLoginext
}

// loginextWebhookPayload mirrors the exact JSON structure Loginext sends in webhooks.
// Private to this package.
type loginextWebhookPayload struct {
	OrderNumber string `json:"orderNo"`      // This is the Synapse AWB we sent them
	Status      string `json:"status"`       // Loginext's internal status string
	Reason      string `json:"reason"`       // Optional failure reason
}

func (l *LoginextNormalizer) NormalizeStatusUpdate(ctx context.Context, rawPayload []byte) (*domain.NormalizedStatusUpdate, error) {
	var inbound loginextWebhookPayload
	if err := json.Unmarshal(rawPayload, &inbound); err != nil {
		return nil, errors.Wrap(err, errors.CodeValidation, "failed to unmarshal Loginext webhook payload")
	}

	canonicalStatus, err := mapLoginextStatus(inbound.Status)
	if err != nil {
		return nil, err
	}

	return &domain.NormalizedStatusUpdate{
		ProviderAWB:        inbound.OrderNumber,
		NewCanonicalStatus: canonicalStatus,
		ProviderRawStatus:  inbound.Status,
		SourceProvider:     ProviderNameLoginext,
	}, nil
}

// mapLoginextStatus is the single place in the entire codebase that knows
// how Loginext status strings map to Synapse canonical statuses.
// When Loginext adds a new status, this is the only function that changes.
func mapLoginextStatus(loginextStatus string) (domain.CanonicalOrderStatus, error) {
	statusMap := map[string]domain.CanonicalOrderStatus{
		"NOTDISPATCHED":    domain.CanonicalStatusReadyToShip,
		"INTRANSIT":        domain.CanonicalStatusInTransit,
		"PICKEDUP":         domain.CanonicalStatusPickedUp,
		"OUT_FOR_DELIVERY": domain.CanonicalStatusOutForDelivery,
		"DELIVERED":        domain.CanonicalStatusDelivered,
		"ATTEMPTED_FAILED": domain.CanonicalStatusAttemptFailed,
		"CANCELLED":        domain.CanonicalStatusCancelled,
		"RETURNED":         domain.CanonicalStatusReturned,
	}

	canonical, found := statusMap[loginextStatus]
	if !found {
		// Unknown statuses become EXCEPTION — we never silently drop them.
		// The raw status is preserved in ProviderRawStatus for investigation.
		return domain.CanonicalStatusException, errors.New(
			errors.CodeValidation,
			fmt.Sprintf("Loginext status %q has no canonical mapping; defaulting to EXCEPTION", loginextStatus),
		)
	}

	return canonical, nil
}
