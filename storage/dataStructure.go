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
		"computerCounters": {name: "computerCounters",
			columns: []metaColumn{
				{name: "id", datatype: "INTEGER"},
				{name: "active", datatype: "BOOLEAN"},
				{name: "fullName", datatype: "TEXT"},
				{name: "computer", datatype: "INTEGER"},
				{name: "label", datatype: "TEXT"},
				{name: "name", datatype: "TEXT"},
				{name: "subName", datatype: "TEXT"},
				{name: "description", datatype: "TEXT"},
			},
		},
		"processCounters": {name: "processCounters",
			columns: []metaColumn{
				{name: "id", datatype: "INTEGER"},
				{name: "active", datatype: "BOOLEAN"},
				{name: "fullName", datatype: "TEXT"},
				{name: "computer", datatype: "INTEGER"},
				{name: "label", datatype: "TEXT"},
				{name: "name", datatype: "TEXT"},
				{name: "description", datatype: "TEXT"},
			},
		},
		"processCountersData": {name: "processCountersData",
			columns: []metaColumn{
				{name: "counter", datatype: "INTEGER"},
				{name: "process", datatype: "INTEGER"},
				{name: "data", datatype: "INTEGER"},
			},
			indexes: []string{
				"CREATE UNIQUE INDEX IF NOT EXISTS processCountersData1 ON processCountersData (counter, process, data)",
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
				"CREATE INDEX IF NOT EXISTS dataPoints1 ON dataPoints (counter, timeStamp)",
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
FROM computerCounters up
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
        AND timeStamp <= (SELECT timeTo from dataFilter)
        AND timeStamp >= (SELECT timeFrom from dataFilter)
    GROUP BY counter
    ) as up
WHERE
    computerInfo.counter = up.counter
`,
			},
		},
		"processInfo": {name: "processInfo",
			columns: []metaColumn{
				{name: "id", datatype: "INTEGER"},
				{name: "active", datatype: "BOOLEAN"},
				{name: "computer", datatype: "INTEGER"},
				{name: "pid", datatype: "INTEGER"},
				{name: "ppid", datatype: "INTEGER"},
				{name: "name", datatype: "TEXT"},
				{name: "commandLine", datatype: "TEXT"},
				{name: "exitCode", datatype: "TEXT"},
				{name: "startTime", datatype: "DATETIME", isTimeFrom: true},
				{name: "endTime", datatype: "DATETIME", isTimeTo: true},
			},
			pivot: metaPivot{
				columns: `
SELECT name FROM processCounters ORDER BY name;
`,
				create: `
ALTER TABLE processInfo ADD COLUMN %[1]s_min REAL; 
ALTER TABLE processInfo ADD COLUMN %[1]s_max REAL;
`,
				calc: `
UPDATE processInfo
SET %[1]s_min = up.min, %[1]s_max = up.max
FROM (
    SELECT
        process,
        MIN(value) as min,
        MAX(value) as max
    FROM dataPoints dp
        INNER JOIN processCountersData pc
        ON dp.counter = pc.data
        AND pc.counter = (SELECT id from processCounters WHERE name = "%[1]s")
        AND dp.timeStamp <= (SELECT timeTo from dataFilter)
        AND dp.timeStamp >= (SELECT timeFrom from dataFilter)
    GROUP BY process
    ) as up
WHERE
    processInfo.id = up.process
`,
			},
		},
	}
}
