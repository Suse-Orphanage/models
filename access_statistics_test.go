package models

import (
	"encoding/json"
	"testing"
)

func TestReport(t *testing.T) {
	dcs := getDcs()

	err := Connect(dcs)
	if err != nil {
		t.Error("failed to connect to database.")
		t.Fatal(err)
	}
	rpt, err := GetLatestDailyReport()
	if err != nil {
		t.Error(err)
	}
	t.Log(rpt)
}

func TestSummary(t *testing.T) {
	dcs := getDcs()

	err := Connect(dcs)
	if err != nil {
		t.Error("failed to connect to database.")
		t.Fatal(err)
	}
	rpt := GetOverallStasticsSummary(60)
	if rpt.TotalCount == 0 {
		t.Fail()
	}
	t.Log(json.Marshal(rpt))
}
