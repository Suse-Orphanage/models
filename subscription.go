package models

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Price uint64

// SubscriptionTybe 表示订阅方式，有按月，一次性等
type SubscriptionType uint

const (
	SubscriptionTypeOneTime uint = iota
	SubscriptionTypeMonthly
)

// SubscriptionPlan 表示一个订阅方案，是一个 IProduct
type SubscriptionPlan struct {
	gorm.Model
	Name      string
	Type      SubscriptionType
	Expiry    uint  // days of expiry
	Price     Price `gorm:"type:uint"` // 原价
	SalePrice Price `gorm:"type:uint"` // 促销价格
	OnSale    bool
}

// Subscription 表示一个既定的订阅，用户独立
type Subscription struct {
	gorm.Model
	UserID           uint
	User             User `gorm:"foreignKey:user_id"`
	SubscriptionPlan uint
	Plan             SubscriptionPlan `gorm:"foreignKey:subscription_plan"`
	ExpiresAt        time.Time        `gorm:"notNull"`
}

func (sp *SubscriptionPlan) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID     uint             `json:"id"`
		Name   string           `json:"name"`
		Type   SubscriptionType `json:"type"`
		Expiry uint             `json:"expiry"`
		Price  float64          `json:"price"`
	}{
		ID:     sp.ID,
		Name:   sp.Name,
		Type:   sp.Type,
		Expiry: sp.Expiry,
		Price:  sp.Price.ToFloat64(),
	})
}

func (p Price) ToInt() int64 {
	return int64(p)
}

func (p Price) ToFloat64() float64 {
	return float64(p) / 100
}

func ToPrice(p float64) Price {
	return Price(p * 100)
}

func GetAllSubscriptionPlan() []SubscriptionPlan {
	result := make([]SubscriptionPlan, 0)

	tx := db.Find(&result)
	if tx.Error != nil {
		logrus.WithError(tx.Error).Error("failed to retrive subscription classes.")
		return nil
	}
	return result
}

func (sp SubscriptionPlan) GetName() string {
	return sp.Name
}
func (sp SubscriptionPlan) GetBasePrice() Price {
	return sp.Price
}
func (sp SubscriptionPlan) GetSalePrice() Price {
	return sp.SalePrice
}
func (sp SubscriptionPlan) GetType() ProductType {
	return ProductTypeSubscriptionPlan
}
func (sp SubscriptionPlan) GetID() uint {
	return sp.ID
}
func (sp SubscriptionPlan) IsOnSale() bool {
	return sp.OnSale
}

// 应用优惠券，返回原价，折扣后的价格，错误
func (sp SubscriptionPlan) ApplyCoupon(coupon *Coupon, u *User) (Price, Price, error) {
	price := sp.Price
	if sp.OnSale {
		price = sp.SalePrice
	}

	discount, err := coupon.Discount(sp, u)
	if err != nil {
		return 0, 0, nil
	}

	return price, price - discount, nil
}
