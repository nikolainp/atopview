package logparser

import "fmt"

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
	labelNET1
	labelNET2
	labelNUM
	labelPRG
	labelPRC
	labelPRE
	labelPRM
	labelPRD
	labelPRN
)

type dataDescription struct {
	label    string
	isSystem bool
	fields   []dataField
	counts   []countField
	scale    func(value float64, intetrval int64, scale float64) float64
	subName  func(dataEntry) string
}

type subField struct {
	name        string
	enable      bool
	description string
}

type dataField struct {
	subField

	isCounter   bool
	isProperty  bool
	isSubName   bool
	isScale     bool
	isNeedScale bool

	// TODO время на середину интервала для счетчиков cpu, disk, ...
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
	pidList := newPidList()

	data := map[entryLabel]dataDescription{
		labelCPUTotal: {
			label: "CPU(total)", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "total number of clock-ticks per second for this machine"}, isScale: true},
				{subField: subField{name: "processors", description: "number of processors"},
					isProperty: true},
				{subField: subField{name: "system", description: "consumption for all CPUs in system mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "user", enable: true, description: "consumption for all CPUs in user mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "user nice", description: "consumption for all CPUs in user mode for niced processes"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "idle", description: "consumption for all CPUs in idle mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "wait", description: "consumption for all CPUs in wait mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "irq", description: "consumption for all CPUs in irq mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "softirq", description: "consumption for all CPUs in softirq mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "steal", description: "consumption for all CPUs in steal mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "guest", description: "consumption for all CPUs in guest mode overlapping user mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "frequency", description: "frequency of all CPUs"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "frequency %", description: "frequency percentage of all CPUs"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "instructions", description: "instructions executed by all CPUs and cycles for all CPUs"},
					isCounter: true},
			},
			counts: []countField{
				{subField: subField{name: "all", enable: true, description: "system + user"}, counting: countCPU},
			},
			scale: func(value float64, intetrval int64, scale float64) float64 {
				return 100 * value / (float64(intetrval) * scale)
			},
		},
		labelCPU: {
			label: "CPU(core)", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "total number of clock-ticks per second for this machine"}, isScale: true},
				{subField: subField{name: "", description: "processor-number"}, isSubName: true},
				{subField: subField{name: "system", description: "consumption for this CPU in system  mode "},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "user", enable: true, description: "consumption for this CPU in user mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "user nicec", description: "consumption for this CPU in user mode for niced processes"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "idle", description: "consumption for this CPU in idle mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "wait", description: "consumption for this CPU in wait mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "irq", description: "consumption for this CPU in irq mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "softirq", description: "consumption for this CPU in softirq mode (clock-ticks"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "steal", description: "consumption for this CPU in steal mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "guest", description: "consumption for this CPU in guest mode overlapping user mode"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "frequency", description: "frequency of this CPU"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "frequency %", description: "frequency percentage of this CPU"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "instructions", description: "instructions executed by this CPU and cycles for this CPU"},
					isCounter: true},
			},
			counts: []countField{
				{subField: subField{name: "all", enable: true, description: "system + user"}, counting: countCPU},
			},
			scale: func(value float64, intetrval int64, scale float64) float64 {
				return 100 * value / (float64(intetrval) * scale)
			},
		},
		labelCPL: {
			//CPL chuwi 1771747662 2026/02/22 11:07:42 588957 4 1.49 0.79 0.70 71097733 58186219
			//Subsequent fields:
			// number of processors,
			// load average for last minute,
			// load average for last five minutes,
			// load average for last fifteen minutes,
			// number  of  context-switches, and
			// number of device interrupts.

			label: "CPU Load", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "nCPU", description: "number of processors"}},
				{subField: subField{name: "avg1", description: "averaged over 1 minutes"},
					isCounter: true},
				{subField: subField{name: "avg5", description: "averaged over 5 minutes "},
					isCounter: true},
				{subField: subField{name: "avg15", description: "averaged over 15 minutes"},
					isCounter: true},
				{subField: subField{name: "cws", description: "number of context switches"},
					isCounter: true},
				{subField: subField{name: "intr", description: "number of device interrupts"},
					isCounter: true},
			},
		},
		labelMEM: {
			// MEM chuwi 1771750664 2026/02/22 11:57:44 600
			// 4096 2994813 439157 1320927 54305 151396 1608 79862 0 281349 0 0 2097152 0 0 0 0 0 342 66 19098 1073741824 0 0 1502448 0
			// MEM chuwi 1771751090 2026/02/22 12:04:50 426 4096 2994813 438952 1320946 54598 151386 2033 79858 0 280040 0 0 2097152 0 0 0 0 0 86 66 19251 1073741824 0 0 1503911 0
			// MEM chuwi 1771766468 2026/02/22 16:21:08 607762 4096 2994813 424679 1321638 54646 151624 1840 79858 0 280626 0 0 2097152 0 0 0 0 0 86 66 19758 1073741824 0 0 1489792 0

			label: "Memory", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "pageSize", description: "page size for this machine (in bytes)"},
					isProperty: true, isScale: true},
				{subField: subField{name: "physical", description: "size of physical memory (in bytes)"},
					isProperty: true, isNeedScale: true},
				{subField: subField{name: "free", enable: true, description: "size of free memory (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "cache", description: "size of page cache (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "bufferCache", description: "size of buffer cache (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "slab", description: "size of slab (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "dirty", description: "dirty pages in cache (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "reclaimableSlab", description: "reclaimable part of slab (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "vmwareBalloon", description: "total size of vmware's balloon pages (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "shared", description: "total size of shared memory (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "residentShared", description: "size of resident shared memory (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "swappedShared", description: "size of swapped shared memory (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "smallerHugePage", description: "smaller huge page size (in bytes)"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "totalSizeSmallerHugePage", description: "total size of smaller huge pages (huge pages)"},
					isCounter: true},
				{subField: subField{name: "freeSmallerHugePage", description: "size of free smaller huge pages (huge pages)"},
					isCounter: true},
				{subField: subField{name: "ARC", description: "size of ARC (cache) of ZFSonlinux (in bytes)"},
					isCounter: true},
				{subField: subField{name: "sharingKSM", description: "size of sharing pages for KSM (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "sharedKSM", description: "size of shared pages for KSM (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "TCP", description: "size of memory used for TCP sockets (in bytes)"},
					isCounter: true, isProperty: true, isNeedScale: true},
				{subField: subField{name: "UDP", description: "size of memory used for UDP sockets (in bytes)"},
					isCounter: true, isProperty: true, isNeedScale: true},
				{subField: subField{name: "pagetables", description: "size of pagetables (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "largerHugePage", description: "larger huge page size (in bytes)"},
					isCounter: true},
				{subField: subField{name: "totalLargerHugePage", description: "total size of larger huge pages (huge pages)"},
					isCounter: true},
				{subField: subField{name: "freeLargerHugePage", description: "size of free larger huge pages (huge pages)"},
					isCounter: true},
				{subField: subField{name: "available", enable: true, description: "size of available memory (in bytes) for new workloads without swapping"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "anonymousHugePages", description: "size of anonymous transparent huge pages (in bytes)"},
					isCounter: true, isNeedScale: true},
			},
		},
		labelSWP: {
			label: "Swap", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "page  size  for  this machine (in bytes)"}, isScale: true},
				{subField: subField{name: "swap", description: "size of swap (in bytes)"},
					isCounter: true, isProperty: true, isNeedScale: true},
				{subField: subField{name: "freeSwap", enable: true, description: "size of free swap (in bytes)"},
					isCounter: true, isProperty: true, isNeedScale: true},
				{subField: subField{name: "cacheSwap", description: "size of swap cache (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "committed", description: "size of committed space (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "committedLimit", description: "limit for committed space (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "sizeSwap", description: "size of the swap cache (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "realZswap", description: "real (decompressed) size of the pages stored in zswap (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "storageZswap", description: "size of compressed storage used for zswap (in bytes)"},
					isCounter: true, isNeedScale: true},
			},
		},
		labelPAG: {
			label: "Page", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "page size for this machine (in bytes)"}, isScale: true},
				{subField: subField{name: "pageScans", description: "number of page scans"},
					isCounter: true},
				{subField: subField{name: "allocstalls", description: "number of allocstalls"},
					isCounter: true},
				{subField: subField{name: "", description: "0 (future use)"}},
				{subField: subField{name: "swapins", description: "number of swapins"},
					isCounter: true},
				{subField: subField{name: "swapouts", description: "number of swapouts"},
					isCounter: true},
				{subField: subField{name: "oomkills", description: "number of oomkills (-1 when counter not present)"},
					isCounter: true},
				{subField: subField{name: "compactions", description: "number of process stalls to run memory compaction"},
					isCounter: true},
				{subField: subField{name: "migrates", description: "number of pages successfully migrated in total"},
					isCounter: true},
				{subField: subField{name: "migratesNUMA", description: "number of NUMA pages migrated"},
					isCounter: true},
				{subField: subField{name: "blockRead", description: "number of pages read from block devices"},
					isCounter: true},
				{subField: subField{name: "blockWrite", description: "number of pages written to block devices"},
					isCounter: true},
				{subField: subField{name: "zswapins", description: "number of swapins from zswap"},
					isCounter: true},
				{subField: subField{name: "zswapouts", description: "number of swapouts to zswap"},
					isCounter: true},
			},
		},
		labelPSI: {
			label: "Pressure Stall", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "PSI statistics present on this system (n or y)"}},
				{subField: subField{name: "", description: "CPU some avg10"}},
				{subField: subField{name: "", description: "CPU some avg60"}},
				{subField: subField{name: "", description: "CPU some avg300"}},
				{subField: subField{name: "", description: "CPU some accumulated microseconds during interval"}},
				{subField: subField{name: "", description: "memory some avg10"}},
				{subField: subField{name: "", description: "memory some avg60"}},
				{subField: subField{name: "", description: "memory some avg300"}},
				{subField: subField{name: "", description: "memory some accumulated microseconds during interval"}},
				{subField: subField{name: "", description: "memory full avg10"}},
				{subField: subField{name: "", description: "memory full avg60,"}},
				{subField: subField{name: "", description: "memory full avg300"}},
				{subField: subField{name: "", description: "memory full accumulated microseconds during interval"}},
				{subField: subField{name: "", description: "I/O some avg10"}},
				{subField: subField{name: "", description: "I/O some avg60"}},
				{subField: subField{name: "", description: "I/O some avg300"}},
				{subField: subField{name: "", description: "I/O some accumulated microseconds during interval"}},
				{subField: subField{name: "", description: "I/O full avg10"}},
				{subField: subField{name: "", description: "I/O full avg60"}},
				{subField: subField{name: "", description: "I/O full avg300"}},
				{subField: subField{name: "", description: "I/O full accumulated microseconds during interva"}},
			},
		},
		labelDSK: {
			label: "Disk", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "name"}, isSubName: true},
				{subField: subField{name: "ios", description: "number of milliseconds spent for I/O"},
					isCounter: true},
				{subField: subField{name: "reads", description: "number of reads issued"},
					isCounter: true},
				{subField: subField{name: "sectorRead", description: "number of sectors transferred for reads"},
					isCounter: true},
				{subField: subField{name: "writes", description: "number of writes issued"},
					isCounter: true},
				{subField: subField{name: "sectorWrite", description: "number of sectors transferred for write"},
					isCounter: true},
				{subField: subField{name: "discards", description: "number of discards issued (-1 if not supported)"},
					isCounter: true},
				{subField: subField{name: "sectorDiscards", description: "number of sectors transferred for discards"},
					isCounter: true},
				{subField: subField{name: "inFlight", description: "number of requests currently in flight (not yet completed)"},
					isCounter: true},
				{subField: subField{name: "queue", description: "average queue depth while the disk was busy"},
					isCounter: true, isProperty: true},
			},
		},
		labelNFC: {
			label: "Network Filesystem (client)", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "RPC", description: "number of transmitted RPCs"},
					isCounter: true},
				{subField: subField{name: "readRPC", description: "number of transmitted read RPCs "},
					isCounter: true},
				{subField: subField{name: "writeRPC", description: "number of transmitted write RPCs"},
					isCounter: true},
				{subField: subField{name: "retransmissionsRPC", description: "number of RPC retransmissions"},
					isCounter: true},
				{subField: subField{name: "authorizationRPC", description: "number of authorization refreshes"},
					isCounter: true},
			},
		},
		labelNFS: {
			label: "Network Filesystem (server)", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "RPC", description: "number of handled RPCs"},
					isCounter: true},
				{subField: subField{name: "readRPC", description: "number of received read RPCs"},
					isCounter: true},
				{subField: subField{name: "writePRC", description: "number of received write RPCs"},
					isCounter: true},
				{subField: subField{name: "readBytes", description: "number of bytes read by clients"},
					isCounter: true},
				{subField: subField{name: "writeBytes", description: "number of bytes written by clients"},
					isCounter: true},
				{subField: subField{name: "badRPC", description: "number of RPCs with bad format"},
					isCounter: true},
				{subField: subField{name: "badAuthRPC", description: "number of RPCs with bad authorization"},
					isCounter: true},
				{subField: subField{name: "badRPC", description: "number of RPCs from bad client"},
					isCounter: true},
				{subField: subField{name: "requests", description: "total number of handled network requests"},
					isCounter: true},
				{subField: subField{name: "TCP", description: "number of handled network requests via TCP"},
					isCounter: true},
				{subField: subField{name: "UPD", description: "number of handled network requests via UDP"},
					isCounter: true},
				{subField: subField{name: "connection", description: "number of handled TCP connections"},
					isCounter: true},
				{subField: subField{name: "cacheHits", description: "number of hits on reply cache"},
					isCounter: true},
				{subField: subField{name: "cacheMiss", description: "number of misses on reply cache"},
					isCounter: true},
				{subField: subField{name: "uncached", description: "number of uncached request"},
					isCounter: true},
			},
		},
		labelNET1: {
			// NET chuwi 1771749462 2026/02/22 11:37:42 600
			// 1-upper 2-14032 3-8538 4-2307 5-1618 6-16921 7-10824 8-16921 9-0 10-0 11-8 12-98 13-0 14-15 15-187 16-1 17-46 18-0
			label: "NET(total)", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "the verb \"upper\""}},
				{subField: subField{name: "recvTCP", description: "number of packets received by TCP"},
					isCounter: true},
				{subField: subField{name: "tranTCP", description: "number of packets transmitted by TCP"},
					isCounter: true},
				{subField: subField{name: "recvTCP", description: "number of packets received by UDP"},
					isCounter: true},
				{subField: subField{name: "tranTCP", description: "number of packets transmitted by UDP,"},
					isCounter: true},
				{subField: subField{name: "recvIP", description: "number of packets received by IP"},
					isCounter: true},
				{subField: subField{name: "tranIP", description: "number of packets transmitted by IP,"},
					isCounter: true},
				{subField: subField{name: "higherIP", description: "number of packets delivered to higher layers by IP"},
					isCounter: true},
				{subField: subField{name: "forwIP", description: "number of packets forwarded by IP"},
					isCounter: true},
				{subField: subField{name: "inErrUDP", description: "number of input errors (UDP)"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "noportErrUDP", description: "number of noport errors (UDP),"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "activetTCP", description: "number of active opens (TCP),"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "passiveTCP", description: "number of passive opens (TCP),"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "estabTCP", description: "number of established connections at this moment (TCP),"},
					isCounter: true},
				{subField: subField{name: "retranTCP", description: "number of retransmitted segments(TCP),"},
					isCounter: true},
				{subField: subField{name: "inErrTCP", description: "number of input errors (TCP),"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "outErrTCP", description: "number of output resets (TCP)"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "checkErrTCP", description: "number of checksum errors on received packets (TCP)"},
					isCounter: true, isProperty: true},
			},
		},
		labelNET2: {
			label: "NET", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "name of the interface"}, isSubName: true},
				{subField: subField{name: "recvPackets", description: "number of packets received by the interface"},
					isCounter: true},
				{subField: subField{name: "recvBytes", description: "number of bytes received by the interface"},
					isCounter: true},
				{subField: subField{name: "packets", description: "number of packets transmitted  by the interface"},
					isCounter: true},
				{subField: subField{name: "bytes", description: "number of bytes transmitted by the interface"},
					isCounter: true},
				{subField: subField{name: "speed", description: "interface speed"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "duplex", description: "duplex mode (0=half, 1=full)"},
					isCounter: true, isProperty: true},
			},
		},
		labelNUM: {
			label: "NUMA", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "NUMA  node  number"}, isSubName: true},
				{subField: subField{name: "pageSize", description: "page size for this machine (in bytes)"},
					isProperty: true, isScale: true},
				{subField: subField{name: "fragmentation", description: "the fragmentation percentage of this node"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "physical", description: "size of physical memory (in bytes)"},
					isCounter: true, isProperty: true, isNeedScale: true},
				{subField: subField{name: "free", description: "size of free memory (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "activeUsed", description: "recently (active) used memory (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "inactiveUsed", description: "less recently (inactive) used memory (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "cached", description: "size of cached file data (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "dirty", description: "dirty pages in cache (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "slabKernel", description: "slab memory being used for kernel mallocs (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "slabReclaim", description: "slab memory that is reclaimable (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "shared", description: "shared memory including tmpfs (in bytes)"},
					isCounter: true, isNeedScale: true},
				{subField: subField{name: "totalHugePages", description: "total huge pages (huge pages)"},
					isCounter: true, isProperty: true},
				{subField: subField{name: "freeHugePages", description: "free huge pages (huge pages)"},
					isCounter: true},
			},
		},
		labelPRG: {
			label: "Program", isSystem: false,
			fields: []dataField{
				{subField: subField{name: "", description: "PID (unique ID of task),"}},
				{subField: subField{name: "", description: "name (between parenthesis or underscores for spaces)"}},
				{subField: subField{name: "", description: "state"}},
				{subField: subField{name: "", description: "real uid"}},
				{subField: subField{name: "", description: "real gid"}},
				{subField: subField{name: "", description: "TGID (group number of related tasks/threads)"}},
				{subField: subField{name: "threads", description: "total number of threads"},
					isCounter: true},
				{subField: subField{name: "", description: "exit code (in case of fatal signal: signal number + 256)"}},
				{subField: subField{name: "", description: "start time (epoch)"}},
				{subField: subField{name: "", description: "full command line  (between  parenthesis  or underscores for spaces)"}},
				{subField: subField{name: "", description: "PPID"}},
				{subField: subField{name: "threadsRun", description: "number of threads in state 'running' (R)"},
					isCounter: true},
				{subField: subField{name: "threadsSleep", description: "number of threads in state 'interruptible sleeping' (S)"},
					isCounter: true},
				{subField: subField{name: "threadsDead", description: "number of threads in state 'uninterruptible sleeping' (D)"},
					isCounter: true},
				{subField: subField{name: "", description: "effective uid"}},
				{subField: subField{name: "", description: "effective gid"}},
				{subField: subField{name: "", description: "saved uid"}},
				{subField: subField{name: "", description: "saved gid"}},
				{subField: subField{name: "", description: "filesystem uid"}},
				{subField: subField{name: "", description: "filesystem gid"}},
				{subField: subField{name: "", description: "elapsed time of terminated process (hertz)"}},
				{subField: subField{name: "", description: "is_process (y/n)"}},
				{subField: subField{name: "", description: "OpenVZ virtual pid (VPID)"}},
				{subField: subField{name: "", description: "OpenVZ container id (CTID)"}},
				{subField: subField{name: "", description: "container/pod name (CID/POD)"}},
				{subField: subField{name: "", description: "indication if the task is newly started during this interval ('N')"}},
				{subField: subField{name: "", description: "cgroup v2 path name (between parenthesis or underscores for spaces)"}},
				{subField: subField{name: "", description: "end time (epoch or 0 if still active)"}},
				{subField: subField{name: "threadIdle", description: "number of threads in state 'idle' (I)"}},
			},
			subName: pidList.getProgrammPid,
		},
		labelPRC: {
			label: "Process", isSystem: false,
			fields: []dataField{
				{subField: subField{name: "", description: "PID"}},
				{subField: subField{name: "", description: "name (between parenthesis or underscores for spaces)"}},
				{subField: subField{name: "", description: "state"}},
				{subField: subField{name: "", description: "total number of clock-ticks per second for this machine"}},
				{subField: subField{name: "", description: "CPU-consumption in user mode (clockticks)"}},
				{subField: subField{name: "", description: "CPU-consumption in system mode (clockticks)"}},
				{subField: subField{name: "", description: "nice value"}},
				{subField: subField{name: "", description: "priority"}},
				{subField: subField{name: "", description: "realtime priority"}},
				{subField: subField{name: "", description: "scheduling policy"}},
				{subField: subField{name: "", description: "current CPU (-1 for exited process)"}},
				{subField: subField{name: "", description: "sleep average"}},
				{subField: subField{name: "", description: "TGID (group number of related tasks/threads)"}},
				{subField: subField{name: "", description: "is_process (y/n)"}},
				{subField: subField{name: "", description: "runqueue delay in nanoseconds for this thread or for all threads (in case of process)"}},
				{subField: subField{name: "", description: "wait channel of this thread (between parenthesis or underscores for spaces)"}},
				{subField: subField{name: "", description: "block I/O delay (clockticks)"}},
				{subField: subField{name: "", description: "cgroup v2 'cpu.max' calculated as  percentage  (-3  means  no cgroup v2 support, -2 means undefined and -1 means maximum)"}},
				{subField: subField{name: "", description: "cgroup v2 most restrictive 'cpu.max' in upper directories calculated as percentage (-3 means no cgroup v2 support, -2 means undefined and -1 means maximum)"}},
				{subField: subField{name: "", description: "number of voluntary context switches"}},
				{subField: subField{name: "", description: "number of involuntary context switches"}},
			},
			subName: pidList.getProcessPid,
		},
		labelPRE: {
			label: "Process(GPU)", isSystem: false,
			fields: []dataField{
				{subField: subField{name: "", description: "PID"}},
				{subField: subField{name: "", description: "name (between parenthesis or underscores for spaces)"}},
				{subField: subField{name: "", description: "process state"}},
				{subField: subField{name: "", description: "GPU state (A for active, E for exited, N for no GPU user)"}},
				{subField: subField{name: "", description: "number  of GPUs  used  by  this process"}},
				{subField: subField{name: "", description: "bitlist reflecting used GPUs"}},
				{subField: subField{name: "", description: "GPU busy percentage during interval"}},
				{subField: subField{name: "", description: "memory busy percentage during interval"}},
				{subField: subField{name: "", description: "memory occupation (KiB) at this moment cumulative memory occupation (KiB) during interval, "}},
				{subField: subField{name: "", description: "number of samples taken during interval"}},
			},
			subName: pidList.getProcessPid,
		},
		labelPRM: {
			label: "Process(Memory)", isSystem: false,
			fields: []dataField{
				{subField: subField{name: "", description: "PID"}},
				{subField: subField{name: "", description: "name (between parenthesis or underscores for spaces)"}},
				{subField: subField{name: "", description: "state"}},
				{subField: subField{name: "", description: "page size for this machine (in bytes)"}},
				{subField: subField{name: "", description: "virtual  memory  size  (KiB)"}},
				{subField: subField{name: "", description: "resident memory  size  (KiB)"}},
				{subField: subField{name: "", description: "shared  text  memory size (KiB)"}},
				{subField: subField{name: "", description: "virtual memory growth (KiB)"}},
				{subField: subField{name: "", description: "resident memory growth (KiB)"}},
				{subField: subField{name: "", description: "number of minor page faults"}},
				{subField: subField{name: "", description: "number of major page faults"}},
				{subField: subField{name: "", description: "virtual library exec size (KiB)"}},
				{subField: subField{name: "", description: "virtual data size (KiB)"}},
				{subField: subField{name: "", description: "virtual stack size (KiB)"}},
				{subField: subField{name: "", description: "swap space used (KiB)"}},
				{subField: subField{name: "", description: "TGID (group number of related tasks/threads)"}},
				{subField: subField{name: "", description: "is_process (y/n)"}},
				{subField: subField{name: "", description: "proportional set size (KiB) if in 'R' option is specified"}},
				{subField: subField{name: "", description: "virtually locked memory space (KiB)"}},
				{subField: subField{name: "", description: "cgroup v2 'memory.max' in KiB (-3 means no cgroup v2  support, -2 means undefined and -1 means maximum)"}},
				{subField: subField{name: "", description: "cgroup v2 most restrictive 'memory.max' in upper directories in KiB (-3 means no cgroup v2 support, -2 means undefined and -1 means maximum)"}},
				{subField: subField{name: "", description: "cgroup v2 'memory.swap.max' in KiB (-3 means no cgroup v2 support, -2 means undefined and -1 means maximum)"}},
				{subField: subField{name: "", description: "cgroup  v2  most restrictive 'memory.swap.max' in upper directories in KiB (-3 means no cgroup v2 support, -2 means undefined and -1 means maximum)"}},
			},
			subName: pidList.getProcessPid,
		},
		labelPRD: {
			label: "Process(Disk)", isSystem: false,
			fields: []dataField{
				{subField: subField{name: "", description: "PID"}},
				{subField: subField{name: "", description: "name (between parenthesis or underscores for spaces)"}},
				{subField: subField{name: "", description: "state, obsoleted kernel patch installed ('n')"}},
				{subField: subField{name: "", description: "standard io statistics used ('y' or 'n')"}},
				{subField: subField{name: "", description: "number of reads on disk"}},
				{subField: subField{name: "", description: "cumulative number of sectors read"}},
				{subField: subField{name: "", description: "number of writes on disk"}},
				{subField: subField{name: "", description: "cumulative number of sectors written "}},
				{subField: subField{name: "", description: "cancelled number of written  sectors"}},
				{subField: subField{name: "", description: "TGID (group number of related tasks/threads)"}},
				{subField: subField{name: "", description: "obsoleted value ('n')"}},
				{subField: subField{name: "", description: "is_process (y/n)"}},
			},
			subName: pidList.getProcessPid,
		},
		labelPRN: {
			label: "Process(Net)", isSystem: false,
			fields: []dataField{
				{subField: subField{name: "", description: "PID"}},
				{subField: subField{name: "", description: "name (between parenthesis or underscores for spaces)"}},
				{subField: subField{name: "", description: "state"}},
				{subField: subField{name: "", description: "kernel module netatop or netatop-bpf installed ('y' or 'n')"}},
				{subField: subField{name: "", description: "number of TCP-packets transmitted"}},
				{subField: subField{name: "", description: "cumulative size of TCP-packets transmitted"}},
				{subField: subField{name: "", description: "number of TCP-packets received"}},
				{subField: subField{name: "", description: "cumulative size of TCP-packets  received"}},
				{subField: subField{name: "", description: "number  of  UDP-packets transmitted"}},
				{subField: subField{name: "", description: "cumulative  size of UDP-packets transmitted"}},
				{subField: subField{name: "", description: "number of UDP-packets received"}},
				{subField: subField{name: "", description: "cumulative size of UDP-packets transmitted"}},
				{subField: subField{name: "", description: "number of raw packets transmitted (obsolete, always 0)"}},
				{subField: subField{name: "", description: "number of raw packets received (obsolete, always 0)"}},
				{subField: subField{name: "", description: "TGID (group number of related tasks/threads)"}},
				{subField: subField{name: "", description: "is_process (y/n)"}},
			},
			subName: pidList.getProcessPid,
		},
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

	if obj.subName == nil {
		length := min(len(data.points), len(obj.fields))
		for i := 0; i < length; i++ {
			if obj.fields[i].isSubName {
				return string(data.points[i])
			}
		}
	} else {
		return obj.subName(data)

	}

	return ""
}

func (obj *dataDescription) getCounters(data dataEntry) (count []keyValue, prop []keyValue, err error) {

	var scale float64 = 1

	length := min(len(data.points), len(obj.fields))
	count = make([]keyValue, 0, length)
	prop = make([]keyValue, 0, length)

	for i := 0; i < length; i++ {
		var value float64
		var field = obj.fields[i]

		if field.isSubName {
			continue
		}
		if field.name == "" && !field.isScale {
			continue
		}

		if value, err = bytesToFloat64(data.points[i]); err != nil {
			return nil, nil, err
		}

		if field.isScale {
			scale = value
		}

		if field.isNeedScale {
			if obj.scale == nil {
				value = value * scale
			} else {
				value = obj.scale(value, data.interval, scale)
			}
		}

		if field.isCounter {
			if field.name == "" {
				return nil, nil, fmt.Errorf("Name needed: %s", field.description)
			}

			count = append(count, keyValue{field.name, value})
		}
		if field.isProperty {
			if field.name == "" {
				return nil, nil, fmt.Errorf("Name needed: %s", field.description)
			}

			prop = append(prop, keyValue{field.name, value})
		}
	}
	for _, field := range obj.counts {
		var value float64

		if value, err = field.counting(data); err != nil {
			return nil, nil, err
		}

		count = append(count, keyValue{field.name, value})
	}

	return count, prop, err
}

///////////////////////////////////////////////////////////////////////////////

type pidList struct {
	data map[string]string
}

func newPidList() *pidList {
	obj := new(pidList)

	obj.data = make(map[string]string)

	return obj
}

func (obj *pidList) getProgrammPid(entry dataEntry) string {
	//  9 - start time (epoch)
	// 26 - "indication if the task is newly started during this interval ('N')"
	pid := string(entry.points[0])
	startTime := string(entry.points[8])

	if oldStartTime, ok := obj.data[pid]; ok {
		if startTime != oldStartTime {
			obj.data[pid] = startTime
		}
	} else {
		obj.data[pid] = startTime
	}

	return fmt.Sprintf("%s_%s", pid, startTime)
}

func (obj *pidList) getProcessPid(entry dataEntry) string {
	pid := string(entry.points[0])
	if startTime, ok := obj.data[pid]; ok {
		return fmt.Sprintf("%s_%s", pid, startTime)
	}
	obj.data[pid] = ""

	return fmt.Sprintf("%s_0", pid)
}
