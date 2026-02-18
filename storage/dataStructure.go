package storage

func getDataStructure() map[string]metaTable {
	return map[string]metaTable{
		"details": {name: "details",
			columns: []metaColumn{
				{name: "title", datatype: "TEXT"},
				{name: "version", datatype: "TEXT"},
				{name: "processingSize", datatype: "INTEGER"},
				{name: "processingSpeed", datatype: "INTEGER"}, {name: "processingTime", datatype: "DATETIME"},
				{name: "firstEventTime", datatype: "DATETIME"}, {name: "lastEventTime", datatype: "DATETIME"},
			},
		},
		"computerDetails": {name: "computerDetails",
			columns: []metaColumn{
				{name: "detail", datatype: "INTEGER"},
				{name: "computer", datatype: "INTEGER"},
			},
		},
		"computers": {name: "computers",
			columns: []metaColumn{
				{name: "id", datatype: "INTEGER"},
				{name: "name", datatype: "TEXT"},
			},
		},
		"counters": {name: "counters",
			columns: []metaColumn{
				{name: "id", datatype: "INTEGER"},
				{name: "fullName", datatype: "TEXT"},
				{name: "computer", datatype: "INTEGER"},
				{name: "label", datatype: "TEXT"},
				{name: "name", datatype: "TEXT"},
				{name: "subName", datatype: "TEXT"},
			},
		},
		"dataPoints": {name: "dataPoints",
			columns: []metaColumn{
				{name: "timeStamp", datatype: "DATETIME"},
				{name: "counter", datatype: "INTEGER"},
				{name: "value", datatype: "REAL"},
			},
		},
	}
}
