package storage

func getDataStructure() map[string]metaTable {
	return map[string]metaTable{
		"dataFilter": {name: "dataFilter",
			columns: []metaColumn{
				{name: "timeFrom", datatype: "DATETIME"},
				{name: "timeTo", datatype: "DATETIME"},
			},
		},
		"details": {name: "details",
			columns: []metaColumn{
				{name: "title", datatype: "TEXT"},
				{name: "version", datatype: "TEXT"},
				// {name: "processingSize", datatype: "INTEGER"},
				// {name: "processingSpeed", datatype: "INTEGER"}, {name: "processingTime", datatype: "DATETIME"},
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
				{name: "system", datatype: "BOLLEAN"},
				{name: "active", datatype: "BOOLEAN"},
				{name: "fullName", datatype: "TEXT"},
				{name: "computer", datatype: "INTEGER"},
				{name: "label", datatype: "TEXT"},
				{name: "name", datatype: "TEXT"},
				{name: "subName", datatype: "TEXT"},
				{name: "description", datatype: "TEXT"},
			},
		},
		"dataPoints": {name: "dataPoints",
			columns: []metaColumn{
				{name: "timeStamp", datatype: "DATETIME",
					isTimeFrom: true, isTimeTo: true},
				{name: "counter", datatype: "INTEGER"},
				{name: "value", datatype: "REAL"},
			},
			indexes: []string{
				"CREATE UNIQUE INDEX IF NOT EXISTS dataPoints1 ON dataPoints (counter, timeStamp)",
			},
		},
		"computerInfo": {name: "computerInfo",
			columns: []metaColumn{
				{name: "computer", datatype: "INTEGER"},
				{name: "counter", datatype: "INTEGER"},
				{name: "label", datatype: "TEXT", isService: true},
				{name: "name", datatype: "TEXT", isService: true},
				{name: "subName", datatype: "TEXT", isService: true},
				{name: "min", datatype: "REAL", isService: true},
				{name: "max", datatype: "REAL", isService: true},
			},
			postLoad: []string{
				`
UPDATE computerInfo
SET label = up.label, name = up.name, subName = up.subName
FROM counters up
WHERE
    computerInfo.counter = up. id
`,
			},
			pivot: metaPivot{
				calc: `
UPDATE computerInfo
SET min = up.min, max = up.max
FROM (
    SELECT 
        counter, 
        MIN(value) as min, 
        MAX(value) as max
    FROM dataPoints
    WHERE
        counter in (SELECT counter from computerInfo)
        AND timeStamp < (SELECT timeTo from dataFilter)
        AND timeStamp > (SELECT timeFrom from dataFilter)
    GROUP BY counter
    ) as up
WHERE
    computerInfo.counter = up.counter
`,
			},
		},
	}
}
