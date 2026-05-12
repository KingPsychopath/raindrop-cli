package cli

import (
	"reflect"
	"testing"
)

func TestCSV(t *testing.T) {
	got := CSV(" go, docs ,, api ")
	want := []string{"go", "docs", "api"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CSV() = %#v, want %#v", got, want)
	}
}

func TestJSONObject(t *testing.T) {
	got, err := JSONObject(`{"important":true,"tags":["go"]}`)
	if err != nil {
		t.Fatal(err)
	}
	if got["important"] != true {
		t.Fatalf("important = %#v, want true", got["important"])
	}
}

func TestInt64CSVRejectsBadID(t *testing.T) {
	if _, err := Int64CSV("1,nope"); err == nil {
		t.Fatal("expected error")
	}
}
