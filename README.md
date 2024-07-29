# hstore
Go package for postgresql hstore type ( https://www.postgresql.org/docs/13/hstore.html )


### 
```go
package main

import (
   "github.com/bobiverse/hstore"
   // ...
)

type MyTest struct {
    Name string
    Data hstore.Hstore
    // ...
}

func main() {
  // Do database things and get values from database with `hstore` type 
  var my MyTest
  my := dummy.GetSqlData("SELECT * FROM dummy LIMIT 1")
  
  // Retrieve different data from hstore field
  myname := my.Data.Get("name") // string
  myage := my.Data.GetInt("age") // int
  mytemperature := my.Data.GetFloat("temp") // float64
  mytime := my.Data.GetTime("birth_time") // try to parse to `time.Time`
  
  // Set new values or overwrites
  my.Data.Set("token", "rnPTD4*xjBG2KR9%jt$a9R2Hh") // new field
  myData.SetIn("age", 21) // set new age value
  
  // Save to database (your code)
  
  
  

}
