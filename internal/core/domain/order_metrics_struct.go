package domain

type OrderMetrics struct {
	OrderID          string `json:"order_id"`
	WeightGrams      int32  `json:"weight_grams"`
	LengthCm         int32  `json:"length_cm"`
	WidthCm          int32  `json:"width_cm"`
	HeightCm         int32  `json:"height_cm"`
	VolumetricWeight int32  `json:"volumetric_weight"`
}
