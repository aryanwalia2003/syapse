// internal/core/normalization/normalize_method.go
package normalization

import (
	"context"
	"fmt"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/errors"
)

// NormalizationEngine holds the registered WMS and DIS adapters and dispatches
// raw payloads to the correct one. It is the single entry point for all
// normalization operations in Synapse.
//
// Why an engine and not direct adapter calls?
// The engine enforces the post-normalization validation contract. An adapter
// can implement the interface but still return a partially-filled struct.
// The engine catches that before the data reaches business logic.
type NormalizationEngine struct {
	wmsNormalizers map[string]domain.WMSNormalizer
	disNormalizers map[string]domain.DISNormalizer
}

// NewNormalizationEngine creates an engine pre-loaded with all known adapters.
// To add a new WMS or DIS provider, register it here — no other file changes.
func NewNormalizationEngine(
	wmsAdapters []domain.WMSNormalizer,
	disAdapters []domain.DISNormalizer,
) *NormalizationEngine {
	engine := &NormalizationEngine{
		wmsNormalizers: make(map[string]domain.WMSNormalizer, len(wmsAdapters)),
		disNormalizers: make(map[string]domain.DISNormalizer, len(disAdapters)),
	}

	for _, adapter := range wmsAdapters {
		engine.wmsNormalizers[adapter.ProviderName()] = adapter
	}
	for _, adapter := range disAdapters {
		engine.disNormalizers[adapter.ProviderName()] = adapter
	}

	return engine
}

// NormalizeWMSOrder finds the correct WMS adapter by providerName, delegates
// translation to it, and then validates the resulting NormalizedOrder.
func (e *NormalizationEngine) NormalizeWMSOrder(
	ctx context.Context,
	providerName string,
	rawPayload []byte,
) (*domain.NormalizedOrder, error) {
	adapter, exists := e.wmsNormalizers[providerName]
	if !exists {
		return nil, errors.New(
			errors.CodeValidation,
			fmt.Sprintf("no WMS normalizer registered for provider: %s", providerName),
		)
	}

	normalizedOrder, err := adapter.NormalizeInboundOrder(ctx, rawPayload)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeValidation,
			fmt.Sprintf("WMS adapter [%s] failed to normalize payload", providerName),
		)
	}

	if validationErr := validateNormalizedOrder(normalizedOrder); validationErr != nil {
		return nil, validationErr
	}

	return normalizedOrder, nil
}

// NormalizeDISStatusUpdate finds the correct DIS adapter by providerName, delegates
// translation to it, and validates the canonical status is a known value.
func (e *NormalizationEngine) NormalizeDISStatusUpdate(
	ctx context.Context,
	providerName string,
	rawPayload []byte,
) (*domain.NormalizedStatusUpdate, error) {
	adapter, exists := e.disNormalizers[providerName]
	if !exists {
		return nil, errors.New(
			errors.CodeValidation,
			fmt.Sprintf("no DIS normalizer registered for provider: %s", providerName),
		)
	}

	statusUpdate, err := adapter.NormalizeStatusUpdate(ctx, rawPayload)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeValidation,
			fmt.Sprintf("DIS adapter [%s] failed to normalize status payload", providerName),
		)
	}

	if validationErr := validateNormalizedStatusUpdate(statusUpdate); validationErr != nil {
		return nil, validationErr
	}

	return statusUpdate, nil
}

// validateNormalizedOrder enforces that an adapter has produced a complete,
// semantically valid NormalizedOrder. This is the system's safety net —
// it catches adapter bugs before corrupt data enters the business logic.
func validateNormalizedOrder(order *domain.NormalizedOrder) error {
	if order == nil {
		return errors.New(errors.CodeInternal, "WMS adapter returned a nil NormalizedOrder")
	}
	if order.ReferenceCode == "" {
		return errors.New(errors.CodeValidation, "normalized order is missing required field: reference_code")
	}
	if order.SourceProvider == "" {
		return errors.New(errors.CodeValidation, "normalized order is missing required field: source_provider")
	}
	if order.OrderDate.IsZero() {
		return errors.New(errors.CodeValidation, "normalized order is missing required field: order_date")
	}
	if !order.CanonicalStatus.IsValid() {
		return errors.New(errors.CodeValidation,
			fmt.Sprintf("normalized order has invalid canonical_status: %q", order.CanonicalStatus),
		)
	}
	if !order.Financials.PaymentMode.IsValid() {
		return errors.New(errors.CodeValidation,
			fmt.Sprintf("normalized order has invalid payment_mode: %q", order.Financials.PaymentMode),
		)
	}
	if order.Financials.TotalAmountPaise < 0 {
		return errors.New(errors.CodeValidation, "normalized order total_amount_paise cannot be negative")
	}
	if order.DeliveryAddress.PinCode == "" {
		return errors.New(errors.CodeValidation, "normalized order is missing required field: delivery_address.pin_code")
	}
	if len(order.Items) == 0 {
		return errors.New(errors.CodeValidation, "normalized order must contain at least one item")
	}
	return nil
}

// validateNormalizedStatusUpdate enforces that a DIS adapter has produced a
// valid, actionable status update.
func validateNormalizedStatusUpdate(update *domain.NormalizedStatusUpdate) error {
	if update == nil {
		return errors.New(errors.CodeInternal, "DIS adapter returned a nil NormalizedStatusUpdate")
	}
	if update.ProviderAWB == "" {
		return errors.New(errors.CodeValidation, "normalized status update is missing required field: provider_awb")
	}
	if update.SourceProvider == "" {
		return errors.New(errors.CodeValidation, "normalized status update is missing required field: source_provider")
	}
	if !update.NewCanonicalStatus.IsValid() {
		return errors.New(errors.CodeValidation,
			fmt.Sprintf("DIS adapter [%s] mapped to unknown canonical status: %q (raw: %q)",
				update.SourceProvider, update.NewCanonicalStatus, update.ProviderRawStatus),
		)
	}
	return nil
}
