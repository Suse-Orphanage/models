package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderType uint

const (
	OrderTypeUnknown OrderType = iota
	// 购买会员
	OrderTypeSubscription
	// 会员自动续费
	OrderTypeSubscriptionAutoRenew
	// 购买小时数
	OrderTypeBuyCredits
	// 购买商品
	OrderTypeBuyProduct
)

func (t *OrderType) MarshalJSON() ([]byte, error) {
	str := ""
	switch *t {
	case OrderTypeSubscription:
		str = "subscription"
	case OrderTypeSubscriptionAutoRenew:
		str = "subscription_auto_renew"
	case OrderTypeBuyCredits:
		str = "buy_credits"
	case OrderTypeBuyProduct:
		str = "buy_product"
	}
	return []byte(`"` + str + `"`), nil
}

type OrderStatus uint

const (
	OrderStatusPending OrderStatus = iota
	OrderStatusPaid
	OrderStatusCancelled
)

func (status *OrderStatus) MarshalJSON() ([]byte, error) {
	str := ""
	switch *status {
	case OrderStatusPending:
		str = "pending"
	case OrderStatusPaid:
		str = "paid"
	}
	return []byte(`"` + str + `"`), nil
}

type Order struct {
	gorm.Model
	TimestamppedID  string    `gorm:"index" json:"oid"`
	Date            time.Time `json:"date"`
	Price           Price     `json:"price"`
	DiscountedPrice Price     `json:"discounted_price"`
	Amount          uint      `json:"amount"`
	Type            OrderType `json:"type"`
	GoodID          uint      `json:"-"`
	Good            Good      `json:"good"`
	CouponID        *uint     `json:"-"`
	Coupon          *Coupon   `json:"-"`
	Affiliate       User      `json:"-"`
	AffiliateID     uint      `json:"-"`

	Data string `json:"-"`

	Status OrderStatus `gorm:"notNull;type:int;default:0" json:"status"`
}

func CreateOrder(o *Order) error {
	tx := db.Create(o)
	return tx.Error
}

func GetOrderByID(id string) (Order, error) {
	var o Order
	tx := db.Preload("Affiliate").Where("timestampped_id = ?", id).First(&o)
	return o, tx.Error
}

// mark paid only marks current order to paid status,
// to keep atomic operation, use CommitPaid instead.
func (o *Order) MarkPaid() error {
	tx := db.Model(o).Update("status", OrderStatusPaid)
	return tx.Error
}

// CommitPaid commits the order to paid status,
// and increase user's credits in one transaction.
func (o *Order) CommitPaid() error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err := tx.
		Model(o).
		Update("status", OrderStatusPaid).
		Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.
		Model(User{}).
		Where("id = ?", o.AffiliateID).
		Update("remaining_credit", gorm.Expr("remaining_credit + ? * 100", o.Amount)).
		Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (u *User) ListOrders() ([]Order, error) {
	orders := []Order{}
	tx := db.Preload("Good").Where("affiliate_id = ?", u.ID).Order("id desc").Find(&orders)
	return orders, tx.Error
}

func GetNetRevenu() Price {
	var result = struct {
		Revenu Price `gorm:"column:revenu" json:"revenu"`
	}{}
	tx := db.
		Where("status = ? AND created_at > ?", OrderStatusPaid, time.Now().AddDate(0, -1, 0)).
		Select("SUM(price) revenu").
		Scan(&result)
	if tx.Error != nil {
		return 0
	}
	return result.Revenu
}
