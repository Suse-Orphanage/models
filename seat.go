package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type SeatStatusEnum uint

const (
	SeatStatusEnumVacancy SeatStatusEnum = iota
	SeatStatusEnumOccupied
)

type Seat struct {
	gorm.Model
	Label         string
	StoreID       uint
	CurrentStatus SeatStatusEnum
	Status        []SeatStatus
	Devices       []Device
}

// deprecated
type SeatStatus struct {
	gorm.Model
	SeatID uint
	Date   uint `gorm:"index"`
	Hour   uint
	Status SeatStatusEnum
}

// deprecated
type SeatSeatusDescriber struct {
	From   uint       `json:"from"`
	Till   uint       `json:"till"`
	Status SeatStatus `json:"status"`
}

// deprecated
type CombinedSeatStatus struct {
	ID     uint                  `json:"id"`
	Status []SeatSeatusDescriber `json:"status"`
}

// deprecated
func (s *Seat) GetNewDayOfStatus(day time.Time) []SeatStatus {
	dayIdentifier := uint(day.Year())*1000 + uint(day.Month())*10 + uint(day.Day())
	status := make([]SeatStatus, 24)
	for i := range status {
		status[i].SeatID = s.ID
		status[i].Date = dayIdentifier
		status[i].Hour = uint(i + 1)
	}

	return status
}

// deprecated
func (s *Seat) CombineStatus(day time.Time) (*CombinedSeatStatus, error) {
	dayIdentifier := uint(day.Year())*1000 + uint(day.Month())*10 + uint(day.Day())
	status := make([]SeatStatus, 0, 24)
	tx := db.Model(&SeatStatus{}).Where("SeatID = ? AND Date = ?", s.ID, dayIdentifier).Find(&status)

	if tx.Error != nil || len(status) == 0 {
		return nil, errors.New("查询座位状态失败")
	}

	ret := &CombinedSeatStatus{
		ID:     s.ID,
		Status: make([]SeatSeatusDescriber, 0),
	}

	currStatus := status[0].Status
	ptr := 0
	for i, sta := range status {
		if sta.Status != currStatus {
			ret.Status = append(ret.Status, SeatSeatusDescriber{
				From: uint(ptr),
				Till: uint(i),
			})

			ptr = i
		}
	}

	return ret, nil
}

func GetSeatByIDWithDevices(id uint) *Seat {
	seat := &Seat{}
	tx := db.Preload("Devices").First(seat, "id = ?", id)
	if tx.Error != nil {
		return nil
	}
	return seat
}

func GetSeatByID(id uint) *Seat {
	seat := &Seat{}
	tx := db.First(seat, "id = ?", id)
	if tx.Error != nil {
		return nil
	}
	return seat
}

func (s *Seat) SetStatus(status SeatStatusEnum) error {
	s.CurrentStatus = status
	tx := db.Save(s)
	return tx.Error
}

type SeatStatusInTimeRange struct {
	StartTime time.Time
	EndTime   time.Time
	Status    SeatStatusEnum
}

type SeatStatusSeries []SeatStatusInTimeRange

type SeatStatusInADay struct {
	Seat   Seat
	Status SeatStatusSeries
}
