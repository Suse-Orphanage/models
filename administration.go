package models

import (
	"errors"
	"time"
)

// ================== Administator ==================
func AddAdministrator(username, password, email string) error {
	return CreateAdmin(username, password, email)
}

func DeleteAdministrator(id uint) error {
	return db.Model(&Administrator{}).Where("id = ?", id).Delete(&Administrator{}).Error
}

func ListAdministrator(limit, page uint) ([]*Administrator, error) {
	result := make([]*Administrator, 0)
	tx := db.Limit(int(limit)).Offset(int(limit * (page - 1))).Find(&result)
	return result, tx.Error
}

func UpdateAdministrator(admin *Administrator) error {
	admin.Password = encryptPassword(admin.Password, admin.Salt)
	return db.Save(admin).Error
}

// ================== Checkin ==================
func ListCheckin(limit, page uint) ([]*CheckIn, error) {
	result := make([]*CheckIn, 0)
	tx := db.Limit(int(limit)).Offset(int(limit * (page - 1))).Find(&result)
	return result, tx.Error
}

// =================== Configuration =======================
func ListConfiguration(limit, page uint) ([]*DynamicConfiguration, error) {
	result := make([]*DynamicConfiguration, 0)
	tx := db.Limit(int(limit)).Offset(int(limit * (page - 1))).Find(&result)
	return result, tx.Error
}

func SetConfigurationValue(name, value string) error {
	return db.
		Model(&DynamicConfiguration{}).
		Where("name = ?", name).
		Update("value", value).
		Error
}

func DeleteConfiguration(name string) error {
	return db.
		Model(&DynamicConfiguration{}).
		Where("name = ?", name).
		Delete(&DynamicConfiguration{}).
		Error
}

func AddConfiguration(name, value string) error {
	return db.Create(&DynamicConfiguration{
		Name:  name,
		Value: value,
	}).Error
}

// =================== Event =======================
func ListEvent(limit, page uint) ([]*Event, error) {
	result := make([]*Event, 0)
	tx := db.Limit(int(limit)).Offset(int(limit * (page - 1))).Find(&result)
	return result, tx.Error
}

func AddEvent(desc, cover, url string, begin, end time.Time) error {
	return db.Create(&Event{
		Desc:      String2Jsonb(desc),
		BeginTime: begin,
		EndTime:   end,
		Cover:     cover,
		Url:       url,
	}).Error
}

func UpdateEvent(evt *Event) error {
	return db.Save(evt).Error
}

func DeleteEvent(id uint) error {
	return db.Model(&Event{}).Where("id = ?", id).Delete(&Event{}).Error
}

func GetEvent(id uint) (*Event, error) {
	result := &Event{}
	err := db.Where("id = ?", id).First(result).Error
	return result, err
}

// ================== Good ====================

func ListGoods(limit, page uint) ([]*Good, error) {
	result := make([]*Good, 0)
	err := db.Limit(int(limit)).Offset(int(limit * (page - 1))).Find(&result).Error
	return result, err
}

// =================== Devices =======================

func ListDevices(limit, page uint) ([]*Device, error) {
	result := make([]*Device, 0)
	tx := db.Limit(int(limit)).Offset(int(limit * (page - 1))).Find(&result)
	return result, tx.Error
}

func AddDevice(name string, t DeviceType, deviceId string, seatId *uint) error {
	if len(deviceId) != 128 {
		return errors.New("device id is not valid")
	}
	device := &Device{
		Name:     name,
		Type:     t,
		Status:   DeviceStatusUnregistered,
		DeviceID: deviceId,
	}
	if seatId != nil {
		device.SeatID = *seatId
	}
	return db.Create(device).Error
}

func UpdateDevice(dev *Device) error {
	return db.Save(dev).Error
}

func DeleteDevice(id uint) error {
	return db.Model(&Device{}).Where("id = ?", id).Delete(&Device{}).Error
}

func GetDevice(id uint) (*Device, error) {
	result := &Device{}
	err := db.Where("id = ?", id).First(result).Error
	return result, err
}

// =================== Order =======================
func ListOrder(limit, page uint) ([]*Order, error) {
	result := make([]*Order, 0)
	tx := db.Limit(int(limit)).Offset(int(limit * (page - 1))).Find(&result)
	return result, tx.Error
}

func GetOrder(id uint) (*Order, error) {
	result := &Order{}
	err := db.Preload("Affiliate").Where("id = ?", id).First(result).Error
	return result, err
}

// ================== Seat ====================
func ListSeat(limit, page uint) ([]*Seat, error) {
	result := make([]*Seat, 0)
	tx := db.Limit(int(limit)).Offset(int(limit * (page - 1))).Find(&result)
	return result, tx.Error
}

func AddSeat(storeId uint, label string) error {
	return db.Create(&Seat{
		StoreID:       storeId,
		Label:         label,
		CurrentStatus: SeatStatusEnumVacancy,
	}).Error
}

func UpdateSeat(seat *Seat) error {
	return db.Save(seat).Error
}

func DeleteSeat(id uint) error {
	return db.Model(&Seat{}).Where("id = ?", id).Delete(&Seat{}).Error
}

// ================== Store ====================
func ListStore(limit, page uint) ([]*Store, error) {
	result := make([]*Store, 0)
	tx := db.Limit(int(limit)).Offset(int(limit * (page - 1))).Find(&result)
	return result, tx.Error
}

func AddStore(location string, openingHours, openingWeekdays uint) error {
	return db.Create(&Store{
		Location:        location,
		Status:          StoreStatusClosed,
		OpeningHours:    openingHours,
		OpeningWeekdays: openingWeekdays,
		SeatCount:       0,
	}).Error
}

func UpdateStore(store *Store) error {
	return db.Save(store).Error
}

func DeleteStore(id uint) error {
	return db.Model(&Store{}).Where("id = ?", id).Delete(&Store{}).Error
}

// ==================== Thread ====================
func ListThread(limit, page uint) ([]uint, error) {
	result := make([]uint, 0)
	tx := db.
		Select("id").
		Where("level = 1").
		Limit(int(limit)).
		Offset(int(limit * (page - 1))).
		Find(&result)
	return result, tx.Error
}
