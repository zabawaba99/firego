package firego

import "strconv"

// StartAt creates a new Firebase reference with the
// requested StartAt configuration.
//
// Reference https://www.firebase.com/docs/rest/guide/retrieving-data.html#section-rest-filtering
func (fb *Firebase) StartAt(value string) *Firebase {
	c := fb.copy()
	if value != "" {
		c.params.Set(startAtParam, value)
	} else {
		c.params.Del(startAtParam)
	}
	return c
}

// EndAt creates a new Firebase reference with the
// requested EndAt configuration.
//
// Reference https://www.firebase.com/docs/rest/guide/retrieving-data.html#section-rest-filtering
func (fb *Firebase) EndAt(value string) *Firebase {
	c := fb.copy()
	if value != "" {
		c.params.Set(endAtParam, value)
	} else {
		c.params.Del(endAtParam)
	}
	return c
}

// OrderBy creates a new Firebase reference with the
// requested OrderBy configuration.
//
// Reference https://www.firebase.com/docs/rest/guide/retrieving-data.html#section-rest-filtering
func (fb *Firebase) OrderBy(value string) *Firebase {
	c := fb.copy()
	if value != "" {
		c.params.Set(orderByParam, value)
	} else {
		c.params.Del(orderByParam)
	}
	return c
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
