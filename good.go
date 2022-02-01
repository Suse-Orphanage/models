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
