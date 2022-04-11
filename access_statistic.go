package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type AccessStatistic struct {
	gorm.Model
	Time            time.Time
	IP              string `gorm:"type:varchar(15);index"`
	UserAgent       string `gorm:"type:text"`
	Path            string `gorm:"type:text"`
	Method          string `gorm:"type:text"`
	Status          int
	Referer         string `gorm:"type:text"`
	UserID          *uint
	User            *User
	AdministratorID *uint
	Administrator   *Administrator

	Headers        json.RawMessage `gorm:"type:jsonb"`
	Country        string          `gorm:"type:varchar(127)"`
	Region         string          `gorm:"type:varchar(255)"`
	City           string          `gorm:"type:varchar(255)"`
	OS             string          `gorm:"type:varchar(255)"`
	OSVersion      string          `gorm:"type:varchar(255)"`
	Browser        string          `gorm:"type:varchar(255)"`
	BrowserVersion string          `gorm:"type:varchar(255)"`
	Device         string          `gorm:"type:varchar(127)"`

	Data json.RawMessage `gorm:"type:jsonb"`
}

func AddStatisticInBatch(stats []AccessStatistic) error {
	return db.Create(&stats).Error
}

func GetStatisticInBatchBefore(before uint) ([]AccessStatistic, error) {
	result := make([]AccessStatistic, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Where("id < ?", before).
		Limit(10).
		Order("id desc").
		Find(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return result, nil
}

func GetStatisticInBatchAfter(after uint) ([]AccessStatistic, error) {
	result := make([]AccessStatistic, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Where("id > ?", after).
		Limit(10).
		Order("id asc").
		Find(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return result, nil
}

func GetStatisticInBatchBeforeAfter(before, after uint) ([]AccessStatistic, error) {
	if before >= after {
		return []AccessStatistic{}, nil
	}
	result := make([]AccessStatistic, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Where("id < ? and id > ?", before, after).
		Limit(10).
		Order("id desc").
		Find(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return result, nil
}

func GetLatestStatistic() ([]AccessStatistic, error) {
	result := make([]AccessStatistic, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Limit(10).
		Order("id desc").
		Find(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return result, nil
}

func GetAccessCountAfterCustomTime(time time.Time) uint {
	var res int64 = 0
	_ = db.
		Model(&AccessStatistic{}).
		Where("time > ?", time).
		Count(&res)
	return uint(res)
}

func GetAccessCountWithIn24Hours() uint {
	return GetAccessCountAfterCustomTime(time.Now().Add(-24 * time.Hour))
}

type ASReport struct {
	Time    time.Time `gorm:"column:t" json:"time"`
	Count   int       `gorm:"column:cnt" json:"count"`
	IdStart uint      `gorm:"column:id_start" json:"id_start"`
	IdEnd   uint      `gorm:"column:id_end" json:"id_end"`
}

func GetDailyReportWithCustomDays(days int) ([]ASReport, error) {
	res := make([]ASReport, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Select("date_trunc('day', time) t, COUNT(*) cnt, MAX(id) id_start, MIN(id) id_end").
		Group("t").
		Order("t desc").
		Limit(days).
		Find(&res)
	return res, tx.Error
}

func GetHourlyReportWithCustomHours(hours int) ([]ASReport, error) {
	res := make([]ASReport, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Select("date_trunc('hour', time) t, COUNT(*) cnt, MAX(id) id_start, MIN(id) id_end").
		Group("t").
		Order("t desc").
		Limit(hours).
		Find(&res)
	return res, tx.Error
}

func GetMonthlyReportWithCustomMonths(months int) ([]ASReport, error) {
	res := make([]ASReport, 0)
	tx := db.
		Model(&AccessStatistic{}).
		Select("date_trunc('month', time) t, COUNT(*) cnt, MAX(id) id_start, MIN(id) id_end").
		Group("t").
		Order("t desc").
		Limit(months).
		Find(&res)
	return res, tx.Error
}

func GetLatestHourlyReport() ([]ASReport, error) {
	return GetHourlyReportWithCustomHours(24)
}

func GetLatestDailyReport() ([]ASReport, error) {
	return GetDailyReportWithCustomDays(30)
}

func GetLatestMonthlyReport() ([]ASReport, error) {
	return GetMonthlyReportWithCustomMonths(3)
}

type CatagoryCount struct {
	Catagory string `gorm:"column:catagory" json:"catagory"`
	Count    uint   `gorm:"column:cnt" json:"count"`
}
type StasticsSummary struct {
	TotalCount uint            `json:"total_count"`
	Browsers   []CatagoryCount `json:"browsers"`
	OSs        []CatagoryCount `json:"os"`
	Devices    []CatagoryCount `json:"devices"`
	Locations  []CatagoryCount `json:"locations"`
	API        []CatagoryCount `json:"api"`

	Reports []ASReport `json:"reports"`
}

func GetOverallStasticsSummary(days int) StasticsSummary {
	browsers := []CatagoryCount{}
	os := []CatagoryCount{}
	devices := []CatagoryCount{}
	locations := []CatagoryCount{}
	api := []CatagoryCount{}

	t := time.Now().Add(-24 * time.Hour * time.Duration(days))
	t.Truncate(24 * time.Hour)

	tx := db.Begin()
	tx.Model(&AccessStatistic{}).
		Select("browser catagory, COUNT(*) cnt").
		Where("time > ?", t).
		Where("browser is not null").
		Where("browser != ''").
		Group("browser").
		Order("cnt desc").
		Find(&browsers)
	tx.Model(&AccessStatistic{}).
		Select("os catagory, COUNT(*) cnt").
		Where("time > ?", t).
		Where("os is not null").
		Where("os != ''").
		Group("os").
		Order("cnt desc").
		Find(&os)
	tx.Model(&AccessStatistic{}).
		Select("device catagory, COUNT(*) cnt").
		Where("time > ?", t).
		Where("device is not null").
		Where("device != ''").
		Group("device").
		Order("cnt desc").
		Find(&devices)
	tx.Model(&AccessStatistic{}).
		Select("concat(country, ',', city) catagory, COUNT(*) cnt").
		Where("time > ?", t).
		Where("country != ''").
		Where("city != ''").
		Group("country, city").
		Order("cnt desc").
		Find(&locations)
	tx.Model(&AccessStatistic{}).
		Select("path catagory, COUNT(*) cnt").
		Where("time > ?", t).
		Group("path").
		Order("cnt desc").
		Limit(10).
		Find(&api)
	tx.Commit()

	reports, _ := GetDailyReportWithCustomDays(days)

	return StasticsSummary{
		TotalCount: GetAccessCountAfterCustomTime(t),
		Browsers:   browsers,
		OSs:        os,
		Devices:    devices,
		Locations:  locations,
		Reports:    reports,
		API:        api,
	}
}
