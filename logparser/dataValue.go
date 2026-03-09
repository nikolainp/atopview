package logparser

type keyValue struct {
	key   string
	value float64
}

///////////////////////////////////////////////////////////////////////////////

type dataValue struct {
	max, min float64
}

func newDataValue(value float64) *dataValue {
	obj := new(dataValue)
	obj.min = value
	obj.max = value
	return obj
}

func (obj *dataValue) set(value float64) {
	obj.min = min(obj.min, value)
	obj.max = max(obj.max, value)
}

func (obj *dataValue) get() (min float64, max float64) {
	return obj.min, obj.max
}
