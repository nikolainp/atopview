package logparser

type entryLabel int

const (
	labelNONE entryLabel = iota
	labelSEP
	labelRESET
	labelCPUTotal
	labelCPU
	labelCPL
	labelMEM
	labelSWP
	labelPAG
	labelPSI
	labelDSK
	labelNFC
	labelNFS
	labelNET
	labelNUM
	labelPRG
	labelPRC
	labelPRE
	labelPRM
	labelPRD
	labelPRN
)

type dataDescription struct {
	label  string
	fields []dataField
	counts []countField
}

type subField struct {
	name        string
	enable      bool
	description string
}

type dataField struct {
	subField
	isSubName   bool
	isScale     bool
	isNeedScale bool
}

type countField struct {
	subField
	counting func(dataEntry) (float64, error)
}

func getDataDescription() map[entryLabel]dataDescription {

	countCPU := func(data dataEntry) (float64, error) {
		var err error
		var scale float64
		var valueSystem, valueUser float64

		if scale, err = bytesToFloat64(data.points[0]); err != nil {
			return 0, err
		}
		if valueSystem, err = bytesToFloat64(data.points[2]); err != nil {
			return 0, err
		}
		if valueUser, err = bytesToFloat64(data.points[3]); err != nil {
			return 0, err
		}

		value := 100 * (valueSystem + valueUser) / (scale * float64(data.interval))

		return value, nil
	}

	data := map[entryLabel]dataDescription{
		labelCPUTotal: {
			label: "CPU(total)",
			fields: []dataField{
				{subField: subField{name: "", description: "total number of clock-ticks per second for this machine"}, isScale: true},
				{subField: subField{name: "", description: "number of processors"}},
				{subField: subField{name: "system", description: "consumption for all CPUs in system mode"}, isNeedScale: true},
				{subField: subField{name: "user", enable: true, description: "consumption for all CPUs in user mode"}, isNeedScale: true},
				{subField: subField{name: "user nice", description: "consumption for all CPUs in user mode for niced processes"}, isNeedScale: true},
				{subField: subField{name: "idle", description: "consumption for all CPUs in idle mode"}, isNeedScale: true},
				{subField: subField{name: "wait", description: "consumption for all CPUs in wait mode"}, isNeedScale: true},
				{subField: subField{name: "irq", description: "consumption for all CPUs in irq mode"}, isNeedScale: true},
				{subField: subField{name: "softirq", description: "consumption for all CPUs in softirq mode"}, isNeedScale: true},
				{subField: subField{name: "steal", description: "consumption for all CPUs in steal mode"}, isNeedScale: true},
				{subField: subField{name: "guest", description: "consumption for all CPUs in guest mode overlapping user mode"}, isNeedScale: true},
				{subField: subField{name: "frequency", description: "frequency of all CPUs"}},
				{subField: subField{name: "frequency %", description: "frequency percentage of all CPUs"}},
				{subField: subField{name: "instructions", description: "instructions executed by all CPUs and cycles for all CPUs"}},
			},
			counts: []countField{
				{subField: subField{name: "all", enable: true, description: "all"}, counting: countCPU},
			},
		},
		labelCPU: {
			label: "CPU(core)",
			fields: []dataField{
				{subField: subField{name: "", description: "total number of clock-ticks per second for this machine"}, isScale: true},
				{subField: subField{name: "", description: "processor-number"}, isSubName: true},
				{subField: subField{name: "system", description: "consumption for this CPU in system  mode "}, isNeedScale: true},
				{subField: subField{name: "user", enable: true, description: "consumption for this CPU in user mode"}, isNeedScale: true},
				{subField: subField{name: "user nicec", description: "consumption for this CPU in user mode for niced processes"}, isNeedScale: true},
				{subField: subField{name: "idle", description: "consumption for this CPU in idle mode"}, isNeedScale: true},
				{subField: subField{name: "wait", description: "consumption for this CPU in wait mode"}, isNeedScale: true},
				{subField: subField{name: "irq", description: "consumption for this CPU in irq mode"}, isNeedScale: true},
				{subField: subField{name: "softirq", description: "consumption for this CPU in softirq mode (clock-ticks"}, isNeedScale: true},
				{subField: subField{name: "steal", description: "consumption for this CPU in steal mode"}, isNeedScale: true},
				{subField: subField{name: "guest", description: "consumption for this CPU in guest mode overlapping user mode"}, isNeedScale: true},
				{subField: subField{name: "frequency", description: "frequency of this CPU"}},
				{subField: subField{name: "frequency %", description: "frequency percentage of this CPU"}},
				{subField: subField{name: "instructions", description: "instructions executed by this CPU and cycles for this CPU"}},
			},
			counts: []countField{
				{subField: subField{name: "all", enable: true, description: "all"}, counting: countCPU},
			},
		},
		labelCPL: {},
		labelMEM: {},
		labelSWP: {},
		labelPAG: {},
		labelPSI: {},
		labelDSK: {},
		labelNFC: {},
		labelNFS: {},
		labelNET: {},
		labelNUM: {},
		labelPRG: {},
		labelPRC: {},
		labelPRE: {},
		labelPRM: {},
		labelPRD: {},
		labelPRN: {},
	}

	return data
}

///////////////////////////////////////////////////////////////////////////////

func (obj *dataDescription) getLabel() string {
	return obj.label
}

func (obj *dataDescription) getDetails(name string) subField {
	for _, field := range obj.fields {
		if field.name == name {
			return field.subField
		}
	}
	for _, field := range obj.counts {
		if field.name == name {
			return field.subField
		}
	}
	return subField{}
}

func (obj *dataDescription) getSubName(data dataEntry) string {

	length := min(len(data.points), len(obj.fields))
	for i := 0; i < length; i++ {
		if obj.fields[i].isSubName {
			return string(data.points[i])
		}
	}

	return ""
}

func (obj *dataDescription) getCounters(data dataEntry) ([]struct {
	key   string
	value float64
}, error) {

	var err error
	var scale float64 = 1

	length := min(len(data.points), len(obj.fields))
	res := make([]struct {
		key   string
		value float64
	}, 0, length)
	for i := 0; i < length; i++ {
		var value float64

		if obj.fields[i].isSubName {
			continue
		}

		if value, err = bytesToFloat64(data.points[i]); err != nil {
			return nil, err
		}

		if obj.fields[i].isScale {
			scale = value
		}

		if obj.fields[i].name == "" {
			continue
		}

		if obj.fields[i].isNeedScale {
			value = 100 * value / (scale * float64(data.interval))
		}

		res = append(res, struct {
			key   string
			value float64
		}{obj.fields[i].name, value})
	}
	for _, count := range obj.counts {
		var value float64

		if value, err = count.counting(data); err != nil {
			return nil, err
		}

		res = append(res, struct {
			key   string
			value float64
		}{count.name, value})
	}

	return res, err
}

func (obj *dataDescription) getProperties([][]byte) {

}
