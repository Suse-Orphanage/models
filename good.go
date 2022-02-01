package models

import "gorm.io/gorm"

type GoodType = uint

const (
	GoodTypeCredits = iota
	GoodTypeSubscription
	GoodTypeProduct
)

type GoodStatus = uint

const (
	GoodStatusOnSale = iota
	GoodStatusSoldOut
	GoodStatusDeleted
)

type Good struct {
	gorm.Model
	Name   string     `gorm:"type:varchar(128)"`
	Price  Price      `json:"price"`
	Type   GoodType   `json:"good_type"`
	Status GoodStatus `json:"status"`

	Description string `json:"description"`
	Image       string `json:"image"`
	SaleCount   uint   `json:"sale_count"`
}

func GetBuiltinGoods() *[]Good {
	return &[]Good{
		{
			Model:       gorm.Model{ID: 1},
			Name:        "Credit",
			Price:       ToPrice(1),
			Type:        GoodTypeCredits,
			Description: "1 Credit",
		},
		{
			Model:       gorm.Model{ID: 2},
			Name:        "月卡",
			Price:       ToPrice(29),
			Type:        GoodTypeSubscription,
			Description: "一个月的会员",
		},
		{
			Model:       gorm.Model{ID: 2},
			Name:        "季卡",
			Price:       ToPrice(69),
			Type:        GoodTypeSubscription,
			Description: "三个月的会员",
		},
	}
}

func GetCreditGoodID() uint {
	return 1
}

func GetMonthlySubscriptionGoodID() uint {
	return 2
}

func GetSeasonSubscriptionGoodID() uint {
	return 3
}

func GetGoodByID(id uint) (*Good, error) {
	var good Good
	tx := db.First(&good, "id = ?", id)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &good, nil
}

func GetCreditGood() *Good {
	good, _ := GetGoodByID(GetCreditGoodID())
	return good
}

func GetMonthlySubscriptionGood() *Good {
	good, _ := GetGoodByID(GetMonthlySubscriptionGoodID())
	return good
}

func GetSeasonSubscriptionGood() *Good {
	good, _ := GetGoodByID(GetSeasonSubscriptionGoodID())
	return good
}
