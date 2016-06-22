package firego

import (
	"strconv"
	"strings"
)

// StartAt creates a new Firebase reference with the
// requested StartAt configuration. The value that is passed in
// is automatically escape if it is a string value.
//
//    StartAt(7)        // -> endAt=7
//    StartAt("foo")    // -> endAt="foo"
//    StartAt(`"foo"`)  // -> endAt="foo"
//
// Reference https://www.firebase.com/docs/rest/guide/retrieving-data.html#section-rest-filtering
func (fb *Firebase) StartAt(value string) *Firebase {
	c := fb.copy()
	if value != "" {
		c.params.Set(startAtParam, escapeString(value))
	} else {
		c.params.Del(startAtParam)
	}
	return c
}

// EndAt creates a new Firebase reference with the
// requested EndAt configuration. The value that is passed in
// is automatically escape if it is a string value.
//
//    EndAt(7)        // -> endAt=7
//    EndAt("foo")    // -> endAt="foo"
//    EndAt(`"foo"`)  // -> endAt="foo"
//
// Reference https://www.firebase.com/docs/rest/guide/retrieving-data.html#section-rest-filtering
func (fb *Firebase) EndAt(value string) *Firebase {
	c := fb.copy()
	if value != "" {
		c.params.Set(endAtParam, escapeString(value))
	} else {
		c.params.Del(endAtParam)
	}
	return c
}

// OrderBy creates a new Firebase reference with the
// requested OrderBy configuration. The value that is passed in
// is automatically escape if it is a string value.
//
//    OrderBy(7)       // -> endAt=7
//    OrderBy("foo")   // -> endAt="foo"
//    OrderBy(`"foo"`) // -> endAt="foo"
//
// Reference https://www.firebase.com/docs/rest/guide/retrieving-data.html#section-rest-filtering
func (fb *Firebase) OrderBy(value string) *Firebase {
	c := fb.copy()
	if value != "" {
		c.params.Set(orderByParam, escapeString(value))
	} else {
		c.params.Del(orderByParam)
	}
	return c
}

// EqualTo sends the query string equalTo so that one can find a single value
//
// Reference https://www.firebase.com/docs/rest/guide/retrieving-data.html#section-rest-filtering
func (fb *Firebase) EqualTo(value string) *Firebase {
	c := fb.copy()
	if value != "" {
		c.params.Set(equalToParam, escapeString(value))
	} else {
		c.params.Del(equalToParam)
	}
	return c
}

func escapeString(s string) string {
	_, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return `"` + strings.Trim(s, `"`) + `"`
	}
	return s
}

// LimitToFirst creates a new Firebase reference with the
// requested limitToFirst configuration.
//
// Reference https://www.firebase.com/docs/rest/api/#section-param-query
func (fb *Firebase) LimitToFirst(value int64) *Firebase {
	c := fb.copy()
	if value > 0 {
		c.params.Set(limitToFirstParam, strconv.FormatInt(value, 10))
	} else {
		c.params.Del(limitToFirstParam)
	}
	return c
}

// LimitToLast creates a new Firebase reference with the
// requested limitToLast configuration.
//
// Reference https://www.firebase.com/docs/rest/api/#section-param-query
func (fb *Firebase) LimitToLast(value int64) *Firebase {
	c := fb.copy()
	if value > 0 {
		c.params.Set(limitToLastParam, strconv.FormatInt(value, 10))
	} else {
		c.params.Del(limitToLastParam)
	}
	return c
}

// Shallow limits the depth of the data returned when calling Value.
// If the data at the location is a JSON primitive (string, number or boolean),
// its value will be returned. If the data is a JSON object, the values
// for each key will be truncated to true.
//
// Reference https://www.firebase.com/docs/rest/api/#section-param-shallow
func (fb *Firebase) Shallow(v bool) {
	if v {
		fb.params.Set(shallowParam, "true")
	} else {
		fb.params.Del(shallowParam)
	}
}

// IncludePriority determines whether or not to ask Firebase
// for the values priority. By default, the priority is not returned.
//
// Reference https://www.firebase.com/docs/rest/api/#section-param-format
func (fb *Firebase) IncludePriority(v bool) {
	if v {
		fb.params.Set(formatParam, formatVal)
	} else {
		fb.params.Del(formatParam)
	}
}
