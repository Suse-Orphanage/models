package models

type Price uint64

func (p Price) ToInt() int64 {
	return int64(p)
}

func (p Price) ToFloat64() float64 {
	return float64(p) / 100
}

func ToPrice(p float64) Price {
	return Price(p * 100)
}
