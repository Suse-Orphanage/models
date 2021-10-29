package models

type ProductType string

const (
	ProductTypeSubscriptionPlan ProductType = "SubscriptionPlan"
)

// IProduct 接口，所有商品都应该实现这个接口
type IProduct interface {
	GetName() string
	GetBasePrice() Price
	GetSalePrice() Price
	IsOnSale() bool

	GetType() ProductType
	GetID() uint

	// 返回值是原价，折扣后的价格，错误
	ApplyCoupon(coupon *Coupon, u *User) (Price, Price, error)
}
