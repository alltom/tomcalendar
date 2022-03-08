package parser

import (
	"encoding/json"
	"github.com/alltom/tomcalendar/pkg/datespec"
	"github.com/go-test/deep"
	"testing"
)

func TestJson(t *testing.T) {
	entry := &Entry{"foo", &datespec.DailyDateSpec{}}
	want := map[string]interface{}{"text": "foo", "spec": map[string]interface{}{"type": "daily"}}

	sjson, err := json.Marshal(entry)
	if err != nil {
		t.Errorf("json.Marshal(%+v) err: %v", entry, err)
	}

	var got map[string]interface{}
	if err = json.Unmarshal(sjson, &got); err != nil {
		t.Errorf("json.Unmarshal(%q, …) err = %v, want nil", string(sjson), err)
	}

	if diff := deep.Equal(want, got); diff != nil {
		t.Error(diff)
	}
}
