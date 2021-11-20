package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderType uint

const (
	// 自动续费
	OrderTypeSubscription = iota
	// 平时的付费
	OrderTypeBilling
)

type Order struct {
	gorm.Model
	TimestamppedID  string    `gorm:"index" json:"oid"`
	Date            time.Time `json:"date"`
	Price           Price     `json:"price"`
	DiscountedPrice Price     `json:"discounted_price"`
	Type            OrderType `json:"type"`
	Coupon          Coupon    `json:"-"`
	Affiliate       User      `json:"-"`
	AffiliateID     uint      `json:"-"`
}

func CreateOrder(o Order) error {
	tx := db.Create(o)
	return tx.Error
}

func GetOrderByID(id string) (Order, error) {
	var o Order
	tx := db.Where("timestampped_id = ?", id).First(&o)
	return o, tx.Error
}

func MarkOrderPaid(id string) error {
	var o Order
	tx := db.Model(&o).Where("timestampped_id = ?", id).Update("paid", true)
	return tx.Error
}
