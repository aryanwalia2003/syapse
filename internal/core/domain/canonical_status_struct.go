package domain

// CanonicalOrderStatus is Synapse's internal, provider-agnostic order status.
// Every WMS and DIS status string MUST be mapped to one of these values.
// Adding a new status here is a deliberate, system-wide decision.
type CanonicalOrderStatus string

// PaymentMode is the canonical payment classification.
type PaymentMode string

const (
	// CanonicalStatusPending means the order has been received but not yet dispatched.
	CanonicalStatusPending CanonicalOrderStatus = "PENDING"

	// CanonicalStatusReadyToShip means the WMS has confirmed the order is packed and ready.
	CanonicalStatusReadyToShip CanonicalOrderStatus = "READY_TO_SHIP"

	// CanonicalStatusPickedUp means the DIS courier has collected the shipment.
	CanonicalStatusPickedUp CanonicalOrderStatus = "PICKED_UP"

	// CanonicalStatusInTransit means the shipment is moving toward the destination.
	CanonicalStatusInTransit CanonicalOrderStatus = "IN_TRANSIT"

	// CanonicalStatusOutForDelivery means the courier is en route to deliver.
	CanonicalStatusOutForDelivery CanonicalOrderStatus = "OUT_FOR_DELIVERY"

	// CanonicalStatusDelivered means the shipment was successfully handed to the recipient.
	CanonicalStatusDelivered CanonicalOrderStatus = "DELIVERED"

	// CanonicalStatusAttemptFailed means a delivery was attempted but could not be completed.
	CanonicalStatusAttemptFailed CanonicalOrderStatus = "ATTEMPT_FAILED"

	// CanonicalStatusCancelled means the order was cancelled before delivery.
	CanonicalStatusCancelled CanonicalOrderStatus = "CANCELLED"

	// CanonicalStatusReturned means the shipment is being or has been sent back to the warehouse.
	CanonicalStatusReturned CanonicalOrderStatus = "RETURNED"

	// CanonicalStatusException means an unclassified or unexpected event occurred.
	CanonicalStatusException CanonicalOrderStatus = "EXCEPTION"
)

const (
	PaymentModeCOD     PaymentMode = "COD"
	PaymentModePrepaid PaymentMode = "PREPAID"
)

// IsValid returns true if the status is one of the known canonical values.
// This is used by the normalization engine to guard against adapters returning garbage.
func (s CanonicalOrderStatus) IsValid() bool {
	switch s {
	case CanonicalStatusPending,
		CanonicalStatusReadyToShip,
		CanonicalStatusPickedUp,
		CanonicalStatusInTransit,
		CanonicalStatusOutForDelivery,
		CanonicalStatusDelivered,
		CanonicalStatusAttemptFailed,
		CanonicalStatusCancelled,
		CanonicalStatusReturned,
		CanonicalStatusException:
		return true
	}
	return false
}

// IsValid returns true if the payment mode is a known canonical value.
func (p PaymentMode) IsValid() bool {
	return p == PaymentModeCOD || p == PaymentModePrepaid
}
