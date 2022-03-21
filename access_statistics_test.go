package models

import (
	"testing"
)

func TestReport(t *testing.T) {
	dcs := getDcs()

	err := Connect(dcs)
	if err != nil {
		t.Error("failed to connect to database.")
		t.Fatal(err)
	}
	rpt, err := GetDailyReport()
	if err != nil {
		t.Error(err)
	}
	t.Log(rpt)
}
