package hstore

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestHstore(t *testing.T) {
	var hs Hstore
	// log.Printf("%+v %+v %+v, %+v", hs, hs.Hstore, hs.Len(), hs.Hstore == nil)

	if !hs.IsEmpty() || hs.Hstore != nil {
		t.Fatalf("Hstore must be empty")
	}

	hs.InitIfEmpty()

	if !hs.IsEmpty() || hs.Hstore == nil {
		t.Fatalf("Hstore still should be empty but initialized")
	}

	hs.Set("aaa", 111)
	if hs.IsEmpty() {
		t.Fatalf("Hstore should not be empty now")
	}

}

//
// func TestHstoreDB(t *testing.T) {
// 	db, mock := NewMockDatabase()
// 	DB = db
// 	DB.LogMode(true)
//
// 	hs := NewHstore()
// 	hs.Set("hkXXey", "hvaXXl")
//
// 	// mock.ExpectBegin()
// 	rows := sqlmock.NewRows([]string{"hs"}).AddRow(hs) // new ID=1
// 	mock.ExpectQuery(regexp.QuoteMeta(`SELECT ? as hs`)).
// 		WithArgs(`"hkey"=>"hval"`).
// 		WillReturnRows(rows)
// 	// mock.ExpectCommit()
//
// 	DB.Exec(`SELECT 1 as 'a', ? as hs`, hs)
//
// }

func TestHstoreValues(t *testing.T) {
	timetest := time.Now()

	hs := NewHstore()
	hs.Set("aaa", "111")
	hs.SetInt("bbb", 222)
	hs.SetFloat("ccc1", 0.345, 1)
	hs.SetFloat("ccc2", 0.345, 2)
	hs.SetFloat("ccc3", 0.345, 3)
	hs.SetFloat("ccc4", 12.0, 4)
	hs.SetFloat("ccc5", 0.0, 5)
	hs.SetFloat("ccc6", 100, 6)
	hs.SetFloat("ccc7", 100.00, 7)
	hs.SetDayAvg("aaa", 0)    // "aaa_day_avg"
	hs.SetDayAvg("bbb", 1000) // "bbb_day_avg"
	hs.SetInt("k", 123456)
	hs.SetInt("m", 12345678)
	hs.Set("mydate", timetest)

	// Field counts must appear in hstore
	cnt := 14

	if hs.Len() != cnt {
		t.Fatalf("Must be stored %d items. Found: %d", cnt, hs.Len())
	}

	if d := hs.GetInt("aaa"); d != 111 {
		t.Fatalf("Integer `111` must be found. Found: %d", d)
	}

	if s := hs.Get("ccc1"); s != "0.3" {
		t.Fatalf("This must saved with 1 decimal number. Found: %s", s)
	}
	if s := hs.Get("ccc2"); s != "0.34" {
		t.Fatalf("This must saved with 2 decimal number. Found: %s", s)
	}
	if s := hs.Get("ccc3"); s != "0.345" {
		t.Fatalf("This must saved with 2 decimal number. Found: %s", s)
	}
	if s := hs.Get("ccc4"); s != "12" {
		t.Fatalf("Zeroes must be removed. Found: %s", s)
	}
	if s := hs.Get("ccc5"); s != "0" {
		t.Fatalf("Zeroes must be removed. Found: %s", s)
	}
	if s := hs.Get("ccc6"); s != "100" {
		t.Fatalf("Not all zeroes must be removed. Found: %s", s)
	}
	if s := hs.Get("ccc7"); s != "100" {
		t.Fatalf("Not all zeroes must be removed. Found: %s", s)
	}

	if f := hs.GetFloat("bbb_day_avg"); f != 0.2 {
		t.Fatalf("Average must be %.5f. Found: %.5f", 0.2, f)
	}

	if mytime := hs.GetTime("mydate"); mytime.UnixNano() != timetest.UnixNano() {
		t.Fatalf("Time should be [%v]. Found:[%v]", timetest, mytime)
	}

	if f := hs.GetFloat("xxx"); f != 0.0 {
		t.Fatalf("Non-existent key must return 0.0. Found: %v", f)
	}
	if f := hs.GetFloat("mydate"); f != 0.0 {
		t.Fatalf("Non-existent key must return 0.0. Found: %v", f)
	}
	if mytime := hs.GetTime("xxx"); !mytime.IsZero() {
		t.Fatalf("Non-existent key must return zero time. Found: %v", mytime)
	}

	// Deleting simple
	hs.Delete("aaa")
	if s := hs.Get("aaa"); s != "" {
		t.Fatalf("Item `aaa` should be deleted. Found[%s]", s)
	}
	if hs.Len() != cnt-1 {
		t.Fatalf("After delete there should be %d items. Found: %d", cnt-1, hs.Len())
	}

	// Deleting regex
	hs.DeleteByRegex("^ccc.+")
	if s := hs.Get("ccc1"); s != "" {
		t.Fatalf("Item `ccc1` should be deleted. Found[%s]", s)
	}
	if hs.Len() != cnt-8 {
		t.Fatalf("After delete there should be %d items. Found: %d", cnt-8, hs.Len())
	}

	// Cache
	hs.saveToCache("test.cache", 102)
	vcache := hs.loadFromCache("test.cache")
	if vcache == nil {
		t.Fatalf("There should be cached value 102. Found: %v", vcache)
	}
	if fmt.Sprintf("%T", vcache) != "int" {
		t.Fatalf("Cached value type should be int. Found: %T", vcache)
	}
	if vcache.(int) != 102 {
		t.Fatalf("Cached value should be 102. Found: %v", vcache)
	}

	vcache = hs.loadFromCache("no.cache.key")
	if vcache != nil {
		t.Fatalf("Not cached value should be empty. Found: %v", vcache)
	}

	// Append
	hs.Append("words", "apple", ",")
	if s := hs.Get("words"); s != "apple" {
		t.Fatalf("Appended value should be `apple`. Found: %s", s)
	}
	hs.Append("words", "banana", ",")
	if s := hs.Get("words"); s != "apple,banana" {
		t.Fatalf("Appended value should be `apple,banana`. Found: %s", s)
	}
	hs.Append("words", "lemon", ",")
	if s := hs.Get("words"); s != "apple,banana,lemon" {
		t.Fatalf("Appended value should be `apple,banana,lemon`. Found: %s", s)
	}

	// GetSlice
	if arr := hs.GetAsSlice("words", ","); strings.Join(arr, "-") != "apple-banana-lemon" {
		t.Fatalf("Slice value should be `apple banana lemon`. Found: %v", arr)
	}
	if arr := hs.GetAsSlice("no-key", ";"); len(arr) != 0 {
		t.Fatalf("No key slice should be empty. Found: %v", arr)
	}
}
