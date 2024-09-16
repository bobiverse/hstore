package hstore

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm/dialects/postgres"
)

// Compile-time assertion to check if User implements DBModel
var _ sql.Scanner = (*Hstore)(nil)

// Hstore ..
type Hstore struct {
	postgres.Hstore

	// Use cache to save more complex calculations  and reuse them
	// exmaple: GetAsCalcs
	cache map[string]any
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

// Value for SQL interface
func (hstore Hstore) Value() (driver.Value, error) {
	hval, err := hstore.Hstore.Value()
	if err != nil {
		return nil, err
	}

	if hval == nil {
		return nil, nil
	}
	// log.Printf("[%s] %T", string(hval.([]uint8)), val)

	return string(hval.([]uint8)), nil
}

// Scan for SQL interface
func (hstore *Hstore) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var val []uint8

	switch v := value.(type) {
	case []uint8:
		val = v
	case string:
		val = []uint8(v)
	case *string:
		val = []uint8(*v)
	default:
		val = []uint8{}
	}

	if err := hstore.Hstore.Scan(val); err != nil {
		return err
	}

	return nil
}

// save value to cache
func (hstore *Hstore) saveToCache(key string, v any) {
	if hstore.cache == nil {
		hstore.cache = map[string]any{}
	}
	hstore.cache[key] = v
}

// load value from cache
func (hstore *Hstore) loadFromCache(key string) any {
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
	return hstore.Hstore == nil || hstore.Len() == 0
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
func (hstore *Hstore) Set(key string, val any) {
	s := ""

	if val != nil {
		s = fmt.Sprintf("%v", val)
	}

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
// `a|b` + `c` => `a|b|c`
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

	n, _ := strconv.ParseFloat(*s, 32)
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
	return hstore.Hstore[key] != nil
}

// Merge with another *Hstore
func (hstore *Hstore) Merge(hstore2 *Hstore) *Hstore {
	for key, val := range hstore2.Hstore {
		hstore.Set(key, *val)
	}
	return hstore
}

// Gorm

// GormDataType gorm common data type
func (hstore Hstore) GormDataType() string {
	return "hstore"
}

// // GormDBDataType gorm db data type
// func (hstore Hstore) GormDBDataType(db *gorm.DB, field *schema.Field) string {
// 	log.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
// 	switch db.Dialector.Name() {
// 	case "postgres":
// 		return "HSTORE"
// 	}
// 	return ""
// }
