package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"gorm.io/gorm"
)

// ICoupon 是优惠券接口，所有类型的优惠券都需要实现这个接口
type ICoupon interface {
	IfSatifyRestriction(product IProduct, user User) error
	// 返回折扣减去的金额
	Discount(products []IProduct) []IProduct
}

type CouponRestrictions []CouponRestriction

// Scanner for CouponRestrictions
func (c *CouponRestrictions) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("解析 CouponRestrictions 失败", value))
	}

	result := []CouponRestriction{}
	err := json.Unmarshal(bytes, &result)
	*c = result
	return err
}
func (c CouponRestrictions) Value() (driver.Value, error) {
	if len(c) == 0 {
		return nil, nil
	}
	return json.Marshal(c)
}

// RestrictionType 表示优惠券限制类型，有指定商品、商品金额、指定用户等类型
type RestrictionType uint

const (
	// 指定商品
	CouponRestrictionTypeProductLimit RestrictionType = iota
	// 满多少元
	CouponRestrictionTypePriceThreshold
	// 指定用户
	CouponRestrictionTypeSpecifiedUser
)

type CouponRestriction struct {
	Type        RestrictionType
	Restriction interface{}
}

// 表示优惠券类型，如减去固定金额，打折等
type CouponType uint

const (
	// 打折
	CouponTypeDiscountPercentage CouponType = iota
	// 减去固定金额
	CouponTypeDiscountFixedPrice
)

type Coupon struct {
	gorm.Model
	UserID       uint               `json:"-"`
	User         User               `gorm:"foreignKey:user_id" json:"-"`
	Type         CouponType         `gorm:"type:int"`
	Used         bool               `gorm:"notNull;default:0" json:"-"`
	Restrictions CouponRestrictions `gorm:"type:text"`
	DiscountData uint32             `gorm:"type:int"` // 存放优惠券折扣相关数据
}

func IfSatifyProductLimitRestriction(types []ProductType, p IProduct) bool {
	satisfy := false
	for _, t := range types {
		satisfy = satisfy || t == p.GetType()
		if satisfy {
			break
		}
	}

	return satisfy
}

func IfSatifyPriceThreshold(threshold, price Price) bool {
	return threshold.ToInt() <= price.ToInt()
}

func IfSatifySpecifiedUser(restrictedUser, user uint) bool {
	return restrictedUser == user
}

func (coupon *Coupon) IfSatifyRestriction(product IProduct, user *User) error {
	for _, res := range coupon.Restrictions {
		switch res.Type {
		case CouponRestrictionTypeProductLimit:
			if !IfSatifyProductLimitRestriction(res.Restriction.([]ProductType), product) {
				return errors.New("仅指定商品可用")
			}
		case CouponRestrictionTypePriceThreshold:
			if !IfSatifyPriceThreshold(res.Restriction.(Price), product.GetSalePrice()) {
				return errors.New("未达到满减金额")
			}
		case CouponRestrictionTypeSpecifiedUser:
			if !IfSatifySpecifiedUser(res.Restriction.(uint), user.ID) {
				return errors.New("仅指定用户可用")
			}
		}
	}
	return nil
}

// 返回折扣减去的金额
func (coupon *Coupon) Discount(product IProduct, u *User) (Price, error) {
	if err := coupon.IfSatifyRestriction(product, u); err != nil {
		return 0, err
	}

	price := product.GetBasePrice()
	if product.IsOnSale() {
		price = product.GetSalePrice()
	}

	switch coupon.Type {
	case CouponTypeDiscountFixedPrice:
		if coupon.DiscountData > uint32(price.ToInt()) {
			return price, nil
		} else {
			return Price(coupon.DiscountData), nil
		}
	case CouponTypeDiscountPercentage:
		return Price(math.Floor(float64(uint32(price)*coupon.DiscountData) / 100)), nil
	}

	return 0, errors.New("优惠券类型错误")
}

func GetCouponById(id uint) (*Coupon, error) {
	coupon := Coupon{}
	tx := db.Where("id = ?", id).First(&coupon)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &coupon, nil
}
