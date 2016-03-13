package repo

type wsMsg struct {
	Type string `json:"t"`
	Data wsData `json:"d"`
}

type wsData map[string]interface{}

func (d wsData) getInternalData() wsData {
	iv, ok := d["d"]
	if !ok {
		eLogFields("Message does not contain internal data", map[string]interface{}{"key": "d", "data": d})
		return d
	}

	return wsData(iv.(map[string]interface{}))
}

func (d wsData) getString(k string) string {
	iv, ok := d[k]
	if !ok {
		eLogFields("Message does not contain data", map[string]interface{}{"key": k, "data": d})
	}

	return iv.(string)
}

func (d wsData) getFloat(k string) float64 {
	iv, ok := d[k]
	if !ok {
		eLogFields("Message does not contain string", map[string]interface{}{"key": k, "data": d})
	}

	return iv.(float64)
}

func (d wsData) getInt(k string) int64 {
	iv, ok := d[k]
	if !ok {
		eLogFields("Message does not contain string", map[string]interface{}{"key": k, "data": d})
	}

	return iv.(int64)
}
