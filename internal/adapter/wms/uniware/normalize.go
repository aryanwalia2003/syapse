package uniware

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/errors"
)

const ProviderNameUniware = "UNIWARE"

type UniwareNormalizer struct{}

func NewUniwareNormalizer() domain.WMSNormalizer {
	return &UniwareNormalizer{}
}

func (u *UniwareNormalizer) ProviderName() string {
	return ProviderNameUniware
}

type uniwareInboundPayload struct {
	ReferenceCode   string         `json:"reference_code"`
	OrderDate       string         `json:"order_date"` // ISO 8601
	TotalAmount     float64        `json:"total_amount"`
	PaymentMode     string         `json:"payment_mode"`   // "COD" or "Prepaid"
	PackageWeight   float64        `json:"package_weight"` // in kg
	PickupDetails   uniwareAddress `json:"pickup_details"`
	DeliveryDetails uniwareAddress `json:"delivery_details"`
	OrderItems      []uniwareItem  `json:"order_items"`
}

type uniwareAddress struct {
	Name         string  `json:"name"`
	ContactNum   string  `json:"contact_num"`
	AddressLine1 string  `json:"address_line_1"`
	AddressLine2 string  `json:"address_line_2"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	PinCode      string  `json:"pin_code"`
	Country      string  `json:"country"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

type uniwareItem struct {
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`  // in INR
	Weight   float64 `json:"weight"` // in kg
}

func (u *UniwareNormalizer) NormalizeInboundOrder(ctx context.Context, rawPayload []byte) (*domain.NormalizedOrder, error) {
	var inbound uniwareInboundPayload
	if err := json.Unmarshal(rawPayload, &inbound); err != nil {
		return nil, errors.Wrap(err, errors.CodeValidation, "failed to unmarshal Uniware payload")
	}

	orderDate, err := time.Parse(time.RFC3339, inbound.OrderDate)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeValidation,
			fmt.Sprintf("invalid order_date format from Uniware: %q", inbound.OrderDate),
		)
	}

	paymentMode, err := mapUniwarePaymentMode(inbound.PaymentMode)
	if err != nil {
		return nil, err
	}

	// Monetary values from Uniware arrive as float INR; we store as integer paise.
	totalAmountPaise := rupeesToPaise(inbound.TotalAmount)

	// Package weight from Uniware arrives in kg; we store in grams.
	packageWeightGrams := kgToGrams(inbound.PackageWeight)

	normalizedItems := make([]domain.NormalizedOrderItem, 0, len(inbound.OrderItems))
	for _, rawItem := range inbound.OrderItems {
		normalizedItems = append(normalizedItems, domain.NormalizedOrderItem{
			SKU:         rawItem.SKU,
			Name:        rawItem.Name,
			Quantity:    rawItem.Quantity,
			PricePaise:  rupeesToPaise(rawItem.Price),
			WeightGrams: kgToGrams(rawItem.Weight),
		})
	}

	return &domain.NormalizedOrder{
		ReferenceCode:      inbound.ReferenceCode,
		SourceProvider:     ProviderNameUniware,
		OrderDate:          orderDate,
		CanonicalStatus:    domain.CanonicalStatusPending, // All inbound WMS orders start as PENDING
		PackageWeightGrams: packageWeightGrams,
		Financials: domain.NormalizedFinancials{
			PaymentMode:      paymentMode,
			TotalAmountPaise: totalAmountPaise,
			Currency:         "INR",
		},
		PickupAddress:   mapUniwareAddress(inbound.PickupDetails),
		DeliveryAddress: mapUniwareAddress(inbound.DeliveryDetails),
		Items:           normalizedItems,
	}, nil
}

func mapUniwarePaymentMode(raw string) (domain.PaymentMode, error) {
	switch raw {
	case "COD":
		return domain.PaymentModeCOD, nil
	case "Prepaid":
		return domain.PaymentModePrepaid, nil
	default:
		return "", errors.New(errors.CodeValidation,
			fmt.Sprintf("unknown Uniware payment_mode: %q", raw),
		)
	}
}

func mapUniwareAddress(raw uniwareAddress) domain.NormalizedAddress {
	return domain.NormalizedAddress{
		ContactName:  raw.Name,
		Phone:        raw.ContactNum,
		AddressLine1: raw.AddressLine1,
		AddressLine2: raw.AddressLine2,
		City:         raw.City,
		State:        raw.State,
		PinCode:      raw.PinCode,
		Country:      raw.Country,
		Latitude:     raw.Latitude,
		Longitude:    raw.Longitude,
	}
}

func rupeesToPaise(rupees float64) int64 {
	return int64(math.Round(rupees * 100))
}

func kgToGrams(kg float64) int32 {
	return int32(math.Round(kg * 1000))
}
