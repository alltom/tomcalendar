package datespec

import (
	"encoding/json"
	"github.com/go-test/deep"
	"testing"
	"time"
)

func TestJson(t *testing.T) {
	specs := []struct {
		spec DateSpec
		want map[string]interface{}
	}{
		{&DailyDateSpec{}, map[string]interface{}{"type": "daily"}},
		{&UnionDateSpec{[]DateSpec{&DailyDateSpec{}}}, map[string]interface{}{"type": "union", "specs": []interface{}{map[string]interface{}{"type": "daily"}}}},
		{&EveryNthDayDateSpec{4}, map[string]interface{}{"type": "everyNth", "n": float64(4), "baseDate": map[string]interface{}{"year": float64(1970), "month": float64(1), "day": float64(1)}, "spec": map[string]interface{}{"type": "daily"}}},
		{&WeekdayDateSpec{time.Sunday}, map[string]interface{}{"type": "dayOfWeek", "weekday": float64(1)}},
		{&EveryNthWeekdayDateSpec{time.Sunday, 4}, map[string]interface{}{"type": "everyNth", "n": float64(4), "baseDate": map[string]interface{}{"year": float64(1970), "month": float64(1), "day": float64(1)}, "spec": map[string]interface{}{"type": "dayOfWeek", "weekday": float64(1)}}},
		{&DayOfMonthDateSpec{15}, map[string]interface{}{"type": "dayOfMonth", "day": float64(15)}},
		{&YearlyDateSpec{2, 28}, map[string]interface{}{"type": "dayOfYear", "month": float64(2), "day": float64(28)}},
		{&SingleDayDateSpec{2022, 3, 7}, map[string]interface{}{"type": "singleDay", "date": map[string]interface{}{"year": float64(2022), "month": float64(3), "day": float64(7)}}},
	}

	for _, tc := range specs {
		sjson, err := json.Marshal(tc.spec)
		if err != nil {
			t.Errorf("json.Marshal(%+v) err: %v", tc.spec, err)
		}

		var got map[string]interface{}
		if err = json.Unmarshal(sjson, &got); err != nil {
			t.Errorf("json.Unmarshal(%q, …) err = %v, want nil", string(sjson), err)
		}

		if diff := deep.Equal(tc.want, got); diff != nil {
			t.Error(diff)
		}
	}
}
