package models

import "time"

// ================== Administator ==================
func AddAdministrator(username, password, email string) error {
	return CreateAdmin(username, password, email)
}

func DeleteAdministrator(id uint) error {
	return db.Model(&Administrator{}).Where("id = ?", id).Delete(&Administrator{}).Error
}

func ListAdministrator(limit, page uint) []*Administrator {
	result := make([]*Administrator, 0)
	_ = db.Limit(int(limit)).Offset(int(limit * (page - 1))).Find(&result)
	return result
}

func UpdateAdministrator(admin *Administrator) error {
	return db.Save(admin).Error
}

// ================== Checkin ==================
func ListCheckin(limit, page uint) []*CheckIn {
	result := make([]*CheckIn, 0)
	_ = db.Limit(int(limit)).Offset(int(limit * (page - 1))).Find(&result)
	return result
}

// =================== Configuration =======================
func ListConfiguration(limit, page uint) []*DynamicConfiguration {
	result := make([]*DynamicConfiguration, 0)
	_ = db.Limit(int(limit)).Offset(int(limit * (page - 1))).Find(&result)
	return result
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
func ListEvent(limit, page uint) []*Event {
	result := make([]*Event, 0)
	_ = db.Limit(int(limit)).Offset(int(limit * (page - 1))).Find(&result)
	return result
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

func UpadteEvent(evt *Event) error {
	return db.Save(evt).Error
}

func DeleteEvent(id uint) error {
	return db.Model(&Event{}).Where("id = ?", id).Delete(&Event{}).Error
}

// ================== Good ====================
