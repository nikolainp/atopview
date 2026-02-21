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
}

type dataField struct {
	name        string
	isSubName   bool
	isScale     bool
	isNeedScale bool
	description string
}

func getDataDescription() map[entryLabel]dataDescription {
	data := map[entryLabel]dataDescription{
		labelCPUTotal: {
			label: "CPU (total)",
			fields: []dataField{
				{name: "", isScale: true, description: "total number of clock-ticks per second for this machine"},
				{name: "", description: "number of processors"},
				{name: "system", isNeedScale: true, description: "consumption for all CPUs in system mode"},
				{name: "user", isNeedScale: true, description: "consumption for all CPUs in user mode"},
				{name: "user nice", isNeedScale: true, description: "consumption for all CPUs in user mode for niced processes"},
				{name: "idle", isNeedScale: true, description: "consumption for all CPUs in idle mode"},
				{name: "wait", isNeedScale: true, description: "consumption for all CPUs in wait mode"},
				{name: "irq", isNeedScale: true, description: "consumption for all CPUs in irq mode"},
				{name: "softirq", isNeedScale: true, description: "consumption for all CPUs in softirq mode"},
				{name: "steal", isNeedScale: true, description: "consumption for all CPUs in steal mode"},
				{name: "guest", isNeedScale: true, description: "consumption for all CPUs in guest mode overlapping user mode"},
				{name: "frequency", description: "frequency of all CPUs"},
				{name: "frequency %", description: "frequency percentage of all CPUs"},
				{name: "instructions", description: "instructions executed by all CPUs and cycles for all CPUs"},
			},
		},
		labelCPU: {
			label: "CPU (core)",
			fields: []dataField{
				{name: "", isScale: true, description: "total number of clock-ticks per second for this machine"},
				{name: "", isSubName: true, description: "processor-number"},
				{name: "system", isNeedScale: true, description: "consumption for this CPU in system  mode "},
				{name: "user", isNeedScale: true, description: "consumption for this CPU in user mode"},
				{name: "user nicec", isNeedScale: true, description: "consumption for this CPU in user mode for niced processes"},
				{name: "idle", isNeedScale: true, description: "consumption for this CPU in idle mode"},
				{name: "wait", isNeedScale: true, description: "consumption for this CPU in wait mode"},
				{name: "irq", isNeedScale: true, description: "consumption for this CPU in irq mode"},
				{name: "softirq", isNeedScale: true, description: "consumption for this CPU in softirq mode (clock-ticks"},
				{name: "steal", isNeedScale: true, description: "consumption for this CPU in steal mode"},
				{name: "guest", isNeedScale: true, description: "consumption for this CPU in guest mode overlapping user mode"},
				{name: "frequency", description: "frequency of this CPU"},
				{name: "frequency %", description: "frequency percentage of this CPU"},
				{name: "instructions", description: "instructions executed by this CPU and cycles for this CPU"},
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

func (obj *dataDescription) getSubName(data [][]byte) string {

	length := min(len(data), len(obj.fields))
	for i := 0; i < length; i++ {
		if obj.fields[i].isSubName {
			return string(data[i])
		}
	}

	return ""
}

func (obj *dataDescription) getCounters(interval int64, data [][]byte) ([]struct {
	key   string
	value float64
}, error) {

	var err error
	var scale float64 = 1

	length := min(len(data), len(obj.fields))
	res := make([]struct {
		key   string
		value float64
	}, 0, length)
	for i := 0; i < length; i++ {
		var value float64

		if obj.fields[i].isSubName {
			continue
		}

		if value, err = bytesToFloat64(data[i]); err != nil {
			return nil, err
		}

		if obj.fields[i].isScale {
			scale = value
		}

		if obj.fields[i].name == "" {
			continue
		}

		if obj.fields[i].isNeedScale {
			value = 100 * value / (scale * float64(interval))
		}

		res = append(res, struct {
			key   string
			value float64
		}{obj.fields[i].name, value})
	}

	return res, err
}

func (obj *dataDescription) getProperties([][]byte) {

}
