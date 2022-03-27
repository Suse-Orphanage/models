package models

import (
	"time"

	"github.com/sirupsen/logrus"
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

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	SeatCount uint
	SeatMap   []byte

	Seats []Seat
}

// depreacated
type SeatSector struct {
	Label string               `json:"label"`
	Seats []CombinedSeatStatus `json:"seat"`
}

type StoreSeatsStautusSummaryWithSectorLabel struct {
	Label  string             `json:"label"`
	Status []SeatStatusInADay `json:"status"`
}

func getSeatsOfStore(sid uint) []Seat {
	seats := make([]Seat, 0)
	_ = db.Where("store_id = ?", sid).Find(&seats)
	return seats
}

func (s *Store) GetStoreSeatStatus(day time.Time) ([]StoreSeatsStautusSummaryWithSectorLabel, error) {
	seats := getSeatsOfStore(s.ID)

	truncatedDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	status := make([]SeatStatusInADay, len(seats))
	for _, seat := range seats {
		series := GetSeatStatusBySeatID(seat.ID, truncatedDay)
		status = append(status, SeatStatusInADay{
			Seat:   seat,
			Status: series,
		})
	}

	sectors := make(map[string][]SeatStatusInADay, 0)
	for _, st := range status {
		if _, ok := sectors[st.Seat.Label]; !ok {
			sectors[st.Seat.Label] = make([]SeatStatusInADay, 0)
		}
		sectors[st.Seat.Label] = append(sectors[st.Seat.Label], st)
	}

	result := make([]StoreSeatsStautusSummaryWithSectorLabel, 0)
	for sector, sts := range sectors {
		result = append(result, StoreSeatsStautusSummaryWithSectorLabel{
			Label:  sector,
			Status: sts,
		})
	}

	return result, nil
}

func GetSeatStatusBySeatID(seat_id uint, truncatedDay time.Time) SeatStatusSeries {
	sessions := make([]Session, 0)
	tx := db.Where("seat_id = ? AND start_time > ?", seat_id, truncatedDay).Order("start_time asc").Find(&sessions)
	if tx.Error != nil {
		logrus.WithError(tx.Error).Error("error when finding sessions for seat vacancy time range")
		return SeatStatusSeries{}
	}

	result := make(SeatStatusSeries, 0)
	start_time := truncatedDay.Add(time.Hour + 8) // 8:00 AM
	for _, s := range sessions {
		if start_time.Before(s.StartTime.Add(-time.Minute * 5)) {
			result = append(result, SeatStatusInTimeRange{
				StartTime: start_time,
				EndTime:   s.StartTime.Add(-time.Minute * 5),
				Status:    SeatStatusEnumVacancy,
			})
		}
		result = append(result, SeatStatusInTimeRange{
			StartTime: s.StartTime.Add(-time.Minute * 5),
			EndTime:   s.EndTime.Add(time.Minute * 5),
			Status:    SeatStatusEnumOccupied,
		})
		start_time = s.EndTime.Add(time.Minute * 5)
	}

	// if there exists ramaining time after the final session and before 00:00 AM next day, append it.
	if start_time.Before(truncatedDay.Add(time.Hour * 24)) {
		result = append(result, SeatStatusInTimeRange{
			StartTime: start_time,
			EndTime:   truncatedDay.Add(time.Hour * 24),
			Status:    SeatStatusEnumVacancy,
		})
	}

	return result
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
	tx := db.First(store, "id = ?", id)
	if tx.Error != nil {
		return nil
	}
	return store
}
