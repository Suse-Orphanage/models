package models

import (
	"os"
	"testing"
)

func TestMigrate(t *testing.T) {
	var dcs string = ""
	dcs = os.Getenv("DB_CONNECTION_STRING")
	if dcs == "" {
		dcs = "host=localhost user=sheey password=quaephietheiHah5auxop0uuPhufoquu dbname=roomy port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	}
	err := Migrate(dcs)
	if err != nil {
		t.Error(err)
	}
}
