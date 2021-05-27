package hstore

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm/dialects/postgres"
)

// Hstore ..
type Hstore struct {
	postgres.Hstore

	// Use cache to save more complex calculations  and reuse them
	// exmaple: GetAsCalcs
	cache map[string]interface{}
}

// NewHstore ..
func NewHstore() Hstore {
	return Hstore{
		Hstore: postgres.Hstore{},
	}
}

// Len ..
func (hstore *Hstore) Len() int {
	return len(hstore.Hstore)
}

// // Value for SQL interface
// func (hstore Hstore) Value() (driver.Value, error) {
// 	return hstore.Hstore, nil
// }

// save value to cache
func (hstore *Hstore) saveToCache(key string, v interface{}) {
	if hstore.cache == nil {
		hstore.cache = map[string]interface{}{}
	}
	hstore.cache[key] = v
}

// load value from cache
func (hstore *Hstore) loadFromCache(key string) interface{} {
	if hstore.cache == nil {
		return nil
	}

	if v, isCached := hstore.cache[key]; isCached {
		return v
	}

	return nil
}

// IsEmpty ..
func (hstore *Hstore) IsEmpty() bool {
	isEmpty := hstore.Hstore == nil || hstore.Len() == 0

	// clear cache if main map is empty
	if isEmpty && hstore.cache != nil {
		hstore.cache = nil
	}

	return isEmpty
}

// InitIfEmpty ..
func (hstore *Hstore) InitIfEmpty() *Hstore {
	if hstore.IsEmpty() {
		hstore.Hstore = postgres.Hstore{}
	}
	return hstore
}

// Delete ..
func (hstore *Hstore) Delete(key string) {
	delete(hstore.Hstore, key)
}

// DeleteByRegex ..
func (hstore *Hstore) DeleteByRegex(pattern string) {
	for key := range hstore.Hstore {
		if matched, _ := regexp.MatchString(pattern, key); matched {
			hstore.Delete(key)
		}
	}
}

// Print ..
func (hstore *Hstore) Print() {
	fmt.Println(strings.Repeat(".", 80))
	for key, val := range hstore.Hstore {
		color.Magenta("%25s ==> %s\n", key, *val)
	}
	fmt.Println(strings.Repeat(".", 80))
}

// Set ..
func (hstore *Hstore) Set(key string, val interface{}) {
	s := fmt.Sprintf("%v", val)
	hstore.Hstore[key] = &s
}

// SetInt ..
func (hstore *Hstore) SetInt(key string, val int) {
	s := fmt.Sprintf("%d", val)
	hstore.Hstore[key] = &s
}

// SetFloat ..
func (hstore *Hstore) SetFloat(key string, val float64, decimals int) {
	s := fmt.Sprintf("%."+fmt.Sprintf("%d", decimals)+"f", val)
	s = strings.TrimRight(s, "0") // 100.00 ==> 100.
	s = strings.TrimRight(s, ".") // 100.00 ==> 100
	if s == "" {
		s = "0"
	}
	hstore.Hstore[key] = &s
}

// Append value with separator
func (hstore *Hstore) Append(key, val, sep string) {
	s := hstore.Get(key)
	if s != "" {
		s += sep
	}
	s += val
	hstore.Hstore[key] = &s
}

// Get ..
func (hstore *Hstore) Get(key string) string {
	s := hstore.Hstore[key]
	if s == nil {
		return ""
	}

	return *s
}

// GetInt ..
func (hstore *Hstore) GetInt(key string) int {
	n := hstore.GetFloat(key)
	return int(n)
}

// GetFloat ..
func (hstore *Hstore) GetFloat(key string) float64 {
	s := hstore.Hstore[key]
	if s == nil {
		return 0.0
	}

	n, _ := strconv.ParseFloat(*s, 10)
	return n
}

// GetTime ..
func (hstore *Hstore) GetTime(key string) time.Time {
	s := hstore.Hstore[key]
	if s == nil {
		return time.Time{}
	}

	t, _ := dateparse.ParseAny(*s)
	// if err != nil {
	// 	log.Printf("Hstore: GetTime: %s", err)
	// }
	return t
}

// GetAsSlice ..
func (hstore *Hstore) GetAsSlice(key, sep string) []string {
	s := hstore.Hstore[key]
	if s == nil {
		return nil
	}

	arr := strings.Split(*s, sep)
	return arr
}

// GetAsMap - a-->b,c-->d  ===> map[a:b, c:d]
func (hstore *Hstore) GetAsMap(key, sepItem, sepKeyVal string) map[string]string {
	var m = map[string]string{}

	items := hstore.GetAsSlice(key, sepItem)
	for _, s := range items {
		arr := strings.SplitN(s, sepKeyVal, 2)
		if len(arr) == 2 {
			m[arr[0]] = arr[1]
		}
	}

	return m
}

// Have ..
func (hstore *Hstore) Have(key string) bool {
	s := hstore.Hstore[key]
	if s == nil {
		return false
	}

	return true
}

// Merge with another *Hstore
func (hstore *Hstore) Merge(hstore2 *Hstore) *Hstore {
	for key, val := range hstore2.Hstore {
		hstore.Set(key, *val)
	}
	return hstore
}
