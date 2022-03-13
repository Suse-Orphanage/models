package models

import (
	"os"
	"testing"
)

func getDcs() string {
	var dcs string = ""
	dcs = os.Getenv("DB_CONNECTION_STRING")
	if dcs == "" {
		dcs = "host=localhost user=sheey password=quaephietheiHah5auxop0uuPhufoquu dbname=roomy port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	}

	return dcs
}

func TestMigrate(t *testing.T) {
	dcs := getDcs()

	Connect(dcs)
	err := db.Migrator().DropTable(&Thread{})
	if err != nil {
		t.Error(err)
	}

	err = Migrate(dcs)
	if err != nil {
		t.Error(err)
	}
}
