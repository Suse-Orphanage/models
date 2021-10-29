package models

import (
	"time"

	"gorm.io/gorm"
)

type StoreStatus = uint

const (
	StoreStatusOpening StoreStatus = iota
	StoreStatusClosed
)

type Store struct {
	gorm.Model
	Location        string
	Status          StoreStatus `gorm:"type:int"`
	OpeningHours    uint
	OpeningWeekdays uint

	SeatCount uint
	SeatMap   []byte

	Seats []Seat
}

type SeatSector struct {
	Label string               `json:"label"`
	Seats []CombinedSeatStatus `json:"seat"`
}

func (s *Store) GetStoreSeatStatus(day time.Time) ([]SeatSector, error) {
	seats := s.Seats
	sectorMap := make(map[string][]CombinedSeatStatus)

	for _, s := range seats {
		status, err := s.CombineStatus(day)
		if err != nil {
			return nil, err
		}
		if _, ok := sectorMap[s.Label]; !ok {
			sectorMap[s.Label] = make([]CombinedSeatStatus, 0)
		}
		sectorMap[s.Label] = append(sectorMap[s.Label], *status)
	}

	ret := make([]SeatSector, 0, len(sectorMap))

	for sector, arr := range sectorMap {
		ret = append(ret, SeatSector{
			Label: sector,
			Seats: arr,
		})
	}

	return ret, nil
}

func GetStore() *Store {
	store := &Store{}
	tx := db.First(store)
	if tx.Error != nil {
		return nil
	}
	return store
}

func GetStoreByID(id uint) *Store {
	store := &Store{}
	tx := db.Find(store, "id = ?", id)
	if tx.Error != nil {
		return nil
	}
	return store
}
