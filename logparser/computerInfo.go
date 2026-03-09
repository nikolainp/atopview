package logparser

type computerInfo struct {
	id   int
	name string

	properties map[fieldKey]*dataValue
}

type fieldKey struct {
	label, name, subName string
}

///////////////////////////////////////////////////////////////////////////////

func newComputerInfo(id int, name string) *computerInfo {
	obj := new(computerInfo)

	obj.id = id
	obj.name = name
	obj.properties = make(map[fieldKey]*dataValue)

	return obj
}

func (obj *computerInfo) getID() int {
	return obj.id
}
func (obj *computerInfo) getName() string {
	return obj.name
}

func (obj *computerInfo) setProperty(key fieldKey, value float64) {
	if data, ok := obj.properties[key]; ok {
		data.set(value)
	} else {
		obj.properties[key] = newDataValue(value)
	}
}

func (obj *computerInfo) getProperties() map[fieldKey]*dataValue {
	return obj.properties
}

///////////////////////////////////////////////////////////////////////////////
