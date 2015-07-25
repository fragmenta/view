package helpers

import (
	"fmt"
    "time"
    "strings"
	got "html/template"
	
)




// ARRAYS 

// Given a set of interface pointers as variadic args, generate and return a single array
func Array(args...interface{}) []interface{} {
    return []interface{}{args}
}


func CommaSeparatedArray(args []string) string {
    result := ""
    for _, v := range args {
        if len(result) > 0 {
           result = fmt.Sprintf("%s,%s",result,v) 
        } else {
           result = v
        }
        
    }
    return result
}

func Empty(a []interface{}) bool { 
    return len(a) > 0
}

func NotEmpty(a []interface{}) bool { 
    return len(a) > 0
}



// MAPS

// Return an empty map[string]interface{} for use as a context - perhaps call this nothing?
// I think Empty should be reserved as above
func EmptyMap() map[string]interface{} { 
    return map[string]interface{}{}
}


// Set a map key and return the map
func Map(m map[string]interface{},k string, v interface{}) map[string]interface{} {
    m[k] = v
    return m 
}

// Set a map key and return an empty string
func Set(m map[string]interface{},k string, v interface{}) string {
    m[k] = v
    return "" // Render nothing, we want no side effects
}

func SetIf(m map[string]interface{},k string, v interface{},t bool) string {
    if t {
       m[k] = v
    } else {
       m[k] = ""
    }
    return "" // Render nothing, we want no side effects
}



// Append all args to an array, and return that array
func Append(m []interface{},args...interface{}) []interface{} {
    for _, v := range args {
        m = append(m,v)
    }
    return m
}


// Given a set of interface pointers as variadic args, generate and return a map to the values
// This is currently unused as we just use simpler Map add above to add to context
func CreateMap(args...interface{}) map[string]interface{} {
    m := make(map[string]interface{},0)
    
    key := ""
    for _, v := range args {
        if len(key) == 0 {
            key = string(v.(string))
        } else {
            m[key] = v
        }
    }
    
    return m
}


// Does this array of ints contain the given int?
func Contains(list []int64,item int64) bool {
    for _, b := range list {
       if b == item {
           return true
       }
    }
    return false
}

// FIXME - danger global - what happens when running multiple threads
// better instead to use $i of range function...
var i int
func Counter() bool  {
    i = i + 1
    return (i % 2 == 1)
}

func CounterReset() string {
    i = 0
    return ""
}

// Return a formatted time string given a time and optional format
func Time(time time.Time, formats ...string) got.HTML {
	layout := "Jan 2, 2006 at 15:04"
	if len(formats) > 0 {
		layout = formats[0]
	}
	value := fmt.Sprintf(time.Format(layout))
	return got.HTML(Escape(value))
}

// Return a formatted date string given a time and optional format
// Date format layouts are for the date 2006-01-02
func Date(t time.Time, formats ...string) got.HTML {
    
	//layout := "2006-01-02" // Jan 2, 2006
	layout := "Jan 2, 2006"
    if len(formats) > 0 {
		layout = formats[0]
	}
	value := fmt.Sprintf(t.Format(layout))
	return got.HTML(Escape(value))
}



// Return a formatted date string in 2006-01-02
func UTCDate(t time.Time) got.HTML {
    return Date(t,"2006-01-02")
}

// Return a formatted date string in 2006-01-02
func UTCNow() got.HTML {
    return Date(time.Now(),"2006-01-02")
}

// Truncate text to a given length
func Truncate(s string, l int64) string {
    return s
}


// CSV escape (replace , with ,,)
func CSV(s got.HTML) string {
	return strings.Replace(string(s),",",",,",-1)
}


// ESCAPING

func Escape(s string) string {
	return got.HTMLEscapeString(s)
}
func EscapeUrl(s string) string {
	return got.URLQueryEscaper(s)
}


