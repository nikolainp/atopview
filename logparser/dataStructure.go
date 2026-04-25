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
	active      bool
	isProperty  bool
	description string
}

type dataField struct {
	subField

	isNote      bool
	isSubName   bool
	isScale     bool
	isNeedScale bool
	isProcess   bool

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
				{subField: subField{name: "", description: "total number of clock-ticks per second for this machine"},
					isScale: true},
				{subField: subField{name: "processors", isProperty: true, description: "number of processors"}},
				{subField: subField{name: "system", description: "consumption for all CPUs in system mode"},
					isNeedScale: true},
				{subField: subField{name: "user", active: true, description: "consumption for all CPUs in user mode"},
					isNeedScale: true},
				{subField: subField{name: "user nice", description: "consumption for all CPUs in user mode for niced processes"},
					isNeedScale: true},
				{subField: subField{name: "idle", description: "consumption for all CPUs in idle mode"},
					isNeedScale: true},
				{subField: subField{name: "wait", description: "consumption for all CPUs in wait mode"},
					isNeedScale: true},
				{subField: subField{name: "irq", description: "consumption for all CPUs in irq mode"},
					isNeedScale: true},
				{subField: subField{name: "softirq", description: "consumption for all CPUs in softirq mode"},
					isNeedScale: true},
				{subField: subField{name: "steal", description: "consumption for all CPUs in steal mode"},
					isNeedScale: true},
				{subField: subField{name: "guest", description: "consumption for all CPUs in guest mode overlapping user mode"},
					isNeedScale: true},
				{subField: subField{name: "frequency", isProperty: true, description: "frequency of all CPUs"}},
				{subField: subField{name: "frequency %", isProperty: true, description: "frequency percentage of all CPUs"}},
				{subField: subField{name: "instructions", description: "instructions executed by all CPUs and cycles for all CPUs"}},
			},
			counts: []countField{
				{subField: subField{name: "all", active: true, description: "system + user"}, counting: countCPU},
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
					isNeedScale: true},
				{subField: subField{name: "user", active: true, description: "consumption for this CPU in user mode"},
					isNeedScale: true},
				{subField: subField{name: "user nicec", description: "consumption for this CPU in user mode for niced processes"},
					isNeedScale: true},
				{subField: subField{name: "idle", description: "consumption for this CPU in idle mode"},
					isNeedScale: true},
				{subField: subField{name: "wait", description: "consumption for this CPU in wait mode"},
					isNeedScale: true},
				{subField: subField{name: "irq", description: "consumption for this CPU in irq mode"},
					isNeedScale: true},
				{subField: subField{name: "softirq", description: "consumption for this CPU in softirq mode (clock-ticks"},
					isNeedScale: true},
				{subField: subField{name: "steal", description: "consumption for this CPU in steal mode"},
					isNeedScale: true},
				{subField: subField{name: "guest", description: "consumption for this CPU in guest mode overlapping user mode"},
					isNeedScale: true},
				{subField: subField{name: "frequency", isProperty: true, description: "frequency of this CPU"}},
				{subField: subField{name: "frequency %", isProperty: true, description: "frequency percentage of this CPU"}},
				{subField: subField{name: "instructions", description: "instructions executed by this CPU and cycles for this CPU"}},
			},
			counts: []countField{
				{subField: subField{name: "all", active: true, description: "system + user"}, counting: countCPU},
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
				{subField: subField{name: "avg1", description: "averaged over 1 minutes"}},
				{subField: subField{name: "avg5", description: "averaged over 5 minutes "}},
				{subField: subField{name: "avg15", description: "averaged over 15 minutes"}},
				{subField: subField{name: "cws", description: "number of context switches"}},
				{subField: subField{name: "intr", description: "number of device interrupts"}},
			},
		},
		labelMEM: {
			// MEM chuwi 1771750664 2026/02/22 11:57:44 600
			// 4096 2994813 439157 1320927 54305 151396 1608 79862 0 281349 0 0 2097152 0 0 0 0 0 342 66 19098 1073741824 0 0 1502448 0
			// MEM chuwi 1771751090 2026/02/22 12:04:50 426 4096 2994813 438952 1320946 54598 151386 2033 79858 0 280040 0 0 2097152 0 0 0 0 0 86 66 19251 1073741824 0 0 1503911 0
			// MEM chuwi 1771766468 2026/02/22 16:21:08 607762 4096 2994813 424679 1321638 54646 151624 1840 79858 0 280626 0 0 2097152 0 0 0 0 0 86 66 19758 1073741824 0 0 1489792 0

			label: "Memory", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "pageSize", isProperty: true, description: "page size for this machine (in bytes)"},
					isScale: true},
				{subField: subField{name: "physical", isProperty: true, description: "size of physical memory (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "free", active: true, description: "size of free memory (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "cache", description: "size of page cache (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "bufferCache", description: "size of buffer cache (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "slab", description: "size of slab (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "dirty", description: "dirty pages in cache (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "reclaimableSlab", description: "reclaimable part of slab (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "vmwareBalloon", description: "total size of vmware's balloon pages (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "shared", description: "total size of shared memory (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "residentShared", description: "size of resident shared memory (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "swappedShared", description: "size of swapped shared memory (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "smallerHugePage", isProperty: true, description: "smaller huge page size (in bytes)"}},
				{subField: subField{name: "totalSizeSmallerHugePage", description: "total size of smaller huge pages (huge pages)"}},
				{subField: subField{name: "freeSmallerHugePage", description: "size of free smaller huge pages (huge pages)"}},
				{subField: subField{name: "ARC", description: "size of ARC (cache) of ZFSonlinux (in bytes)"}},
				{subField: subField{name: "sharingKSM", description: "size of sharing pages for KSM (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "sharedKSM", description: "size of shared pages for KSM (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "TCP", isProperty: true, description: "size of memory used for TCP sockets (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "UDP", isProperty: true, description: "size of memory used for UDP sockets (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "pagetables", description: "size of pagetables (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "largerHugePage", description: "larger huge page size (in bytes)"}},
				{subField: subField{name: "totalLargerHugePage", description: "total size of larger huge pages (huge pages)"}},
				{subField: subField{name: "freeLargerHugePage", description: "size of free larger huge pages (huge pages)"}},
				{subField: subField{name: "available", active: true, description: "size of available memory (in bytes) for new workloads without swapping"},
					isNeedScale: true},
				{subField: subField{name: "anonymousHugePages", description: "size of anonymous transparent huge pages (in bytes)"},
					isNeedScale: true},
			},
		},
		labelSWP: {
			label: "Swap", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "page  size  for  this machine (in bytes)"}, isScale: true},
				{subField: subField{name: "swap", isProperty: true, description: "size of swap (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "freeSwap", active: true, isProperty: true, description: "size of free swap (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "cacheSwap", description: "size of swap cache (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "committed", description: "size of committed space (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "committedLimit", description: "limit for committed space (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "sizeSwap", description: "size of the swap cache (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "realZswap", description: "real (decompressed) size of the pages stored in zswap (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "storageZswap", description: "size of compressed storage used for zswap (in bytes)"},
					isNeedScale: true},
			},
		},
		labelPAG: {
			label: "Page", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "page size for this machine (in bytes)"}, isScale: true},
				{subField: subField{name: "pageScans", isProperty: true, description: "number of page scans"}},
				{subField: subField{name: "allocstalls", description: "number of allocstalls"}},
				{subField: subField{name: "", description: "0 (future use)"},
					isNote: true},
				{subField: subField{name: "swapins", description: "number of swapins"}},
				{subField: subField{name: "swapouts", description: "number of swapouts"}},
				{subField: subField{name: "oomkills", isProperty: true, description: "number of oomkills (-1 when counter not present)"}},
				{subField: subField{name: "compactions", description: "number of process stalls to run memory compaction"}},
				{subField: subField{name: "migrates", description: "number of pages successfully migrated in total"}},
				{subField: subField{name: "migratesNUMA", description: "number of NUMA pages migrated"}},
				{subField: subField{name: "blockRead", description: "number of pages read from block devices"}},
				{subField: subField{name: "blockWrite", description: "number of pages written to block devices"}},
				{subField: subField{name: "zswapins", description: "number of swapins from zswap"}},
				{subField: subField{name: "zswapouts", description: "number of swapouts to zswap"}},
			},
		},
		labelPSI: {
			label: "Pressure Stall", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "PSI statistics present on this system (n or y)"},
					isNote: true},
				{subField: subField{name: "cpu10", description: "CPU some avg10"}},
				{subField: subField{name: "cpu60", description: "CPU some avg60"}},
				{subField: subField{name: "cpu300", description: "CPU some avg300"}},
				{subField: subField{name: "", description: "CPU some accumulated microseconds during interval"}},
				{subField: subField{name: "memSome10", description: "memory some avg10"}},
				{subField: subField{name: "memSome60", description: "memory some avg60"}},
				{subField: subField{name: "memSome300", description: "memory some avg300"}},
				{subField: subField{name: "", description: "memory some accumulated microseconds during interval"}},
				{subField: subField{name: "memFull10", description: "memory full avg10"}},
				{subField: subField{name: "memFull60", description: "memory full avg60,"}},
				{subField: subField{name: "memFull300", description: "memory full avg300"}},
				{subField: subField{name: "", description: "memory full accumulated microseconds during interval"}},
				{subField: subField{name: "ioSome10", description: "I/O some avg10"}},
				{subField: subField{name: "ioSome60", description: "I/O some avg60"}},
				{subField: subField{name: "ioSome300", description: "I/O some avg300"}},
				{subField: subField{name: "", description: "I/O some accumulated microseconds during interval"}},
				{subField: subField{name: "ioFull10", description: "I/O full avg10"}},
				{subField: subField{name: "ioFull60", description: "I/O full avg60"}},
				{subField: subField{name: "ioFull300", description: "I/O full avg300"}},
				{subField: subField{name: "", description: "I/O full accumulated microseconds during interva"}},
			},
		},
		labelDSK: {
			label: "Disk", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "name"}, isSubName: true},
				{subField: subField{name: "ios", isProperty: true, description: "number of milliseconds spent for I/O"}},
				{subField: subField{name: "reads", description: "number of reads issued"}},
				{subField: subField{name: "sectorRead", description: "number of sectors transferred for reads"}},
				{subField: subField{name: "writes", description: "number of writes issued"}},
				{subField: subField{name: "sectorWrite", description: "number of sectors transferred for write"}},
				{subField: subField{name: "discards", description: "number of discards issued (-1 if not supported)"}},
				{subField: subField{name: "sectorDiscards", description: "number of sectors transferred for discards"}},
				{subField: subField{name: "inFlight", description: "number of requests currently in flight (not yet completed)"}},
				{subField: subField{name: "queue", isProperty: true, description: "average queue depth while the disk was busy"}},
			},
		},
		labelNFC: {
			label: "Network Filesystem (client)", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "RPC", description: "number of transmitted RPCs"}},
				{subField: subField{name: "readRPC", description: "number of transmitted read RPCs "}},
				{subField: subField{name: "writeRPC", description: "number of transmitted write RPCs"}},
				{subField: subField{name: "retransmissionsRPC", description: "number of RPC retransmissions"}},
				{subField: subField{name: "authorizationRPC", description: "number of authorization refreshes"}},
			},
		},
		labelNFS: {
			label: "Network Filesystem (server)", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "RPC", description: "number of handled RPCs"}},
				{subField: subField{name: "readRPC", description: "number of received read RPCs"}},
				{subField: subField{name: "writePRC", description: "number of received write RPCs"}},
				{subField: subField{name: "readBytes", description: "number of bytes read by clients"}},
				{subField: subField{name: "writeBytes", description: "number of bytes written by clients"}},
				{subField: subField{name: "badRPCformat", description: "number of RPCs with bad format"}},
				{subField: subField{name: "badAuthRPC", description: "number of RPCs with bad authorization"}},
				{subField: subField{name: "badRPCclient", description: "number of RPCs from bad client"}},
				{subField: subField{name: "requests", description: "total number of handled network requests"}},
				{subField: subField{name: "TCP", description: "number of handled network requests via TCP"}},
				{subField: subField{name: "UPD", description: "number of handled network requests via UDP"}},
				{subField: subField{name: "connection", description: "number of handled TCP connections"}},
				{subField: subField{name: "cacheHits", description: "number of hits on reply cache"}},
				{subField: subField{name: "cacheMiss", description: "number of misses on reply cache"}},
				{subField: subField{name: "uncached", description: "number of uncached request"}},
			},
		},
		labelNET1: {
			// NET chuwi 1771749462 2026/02/22 11:37:42 600
			// 1-upper 2-14032 3-8538 4-2307 5-1618 6-16921 7-10824 8-16921 9-0 10-0 11-8 12-98 13-0 14-15 15-187 16-1 17-46 18-0
			label: "NET(total)", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "the verb \"upper\""}},
				{subField: subField{name: "recvTCP", description: "number of packets received by TCP"}},
				{subField: subField{name: "tranTCP", description: "number of packets transmitted by TCP"}},
				{subField: subField{name: "recvUDP", description: "number of packets received by UDP"}},
				{subField: subField{name: "tranUDP", description: "number of packets transmitted by UDP,"}},
				{subField: subField{name: "recvIP", description: "number of packets received by IP"}},
				{subField: subField{name: "tranIP", description: "number of packets transmitted by IP,"}},
				{subField: subField{name: "higherIP", description: "number of packets delivered to higher layers by IP"}},
				{subField: subField{name: "forwIP", description: "number of packets forwarded by IP"}},
				{subField: subField{name: "inErrUDP", isProperty: true, description: "number of input errors (UDP)"}},
				{subField: subField{name: "noportErrUDP", isProperty: true, description: "number of noport errors (UDP),"}},
				{subField: subField{name: "activetTCP", isProperty: true, description: "number of active opens (TCP),"}},
				{subField: subField{name: "passiveTCP", isProperty: true, description: "number of passive opens (TCP),"}},
				{subField: subField{name: "estabTCP", description: "number of established connections at this moment (TCP),"}},
				{subField: subField{name: "retranTCP", description: "number of retransmitted segments(TCP),"}},
				{subField: subField{name: "inErrTCP", isProperty: true, description: "number of input errors (TCP),"}},
				{subField: subField{name: "outErrTCP", isProperty: true, description: "number of output resets (TCP)"}},
				{subField: subField{name: "checkErrTCP", isProperty: true, description: "number of checksum errors on received packets (TCP)"}},
			},
		},
		labelNET2: {
			label: "NET", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "name of the interface"}, isSubName: true},
				{subField: subField{name: "recvPackets", description: "number of packets received by the interface"}},
				{subField: subField{name: "recvBytes", description: "number of bytes received by the interface"}},
				{subField: subField{name: "packets", description: "number of packets transmitted  by the interface"}},
				{subField: subField{name: "bytes", description: "number of bytes transmitted by the interface"}},
				{subField: subField{name: "speed", isProperty: true, description: "interface speed"}},
				{subField: subField{name: "duplex", isProperty: true, description: "duplex mode (0=half, 1=full)"}},
			},
		},
		labelNUM: {
			label: "NUMA", isSystem: true,
			fields: []dataField{
				{subField: subField{name: "", description: "NUMA  node  number"}, isSubName: true},
				{subField: subField{name: "pageSize", isProperty: true, description: "page size for this machine (in bytes)"},
					isScale: true},
				{subField: subField{name: "fragmentation", isProperty: true, description: "the fragmentation percentage of this node"}},
				{subField: subField{name: "physical", isProperty: true, description: "size of physical memory (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "free", description: "size of free memory (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "activeUsed", description: "recently (active) used memory (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "inactiveUsed", description: "less recently (inactive) used memory (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "cached", description: "size of cached file data (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "dirty", description: "dirty pages in cache (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "slabKernel", description: "slab memory being used for kernel mallocs (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "slabReclaim", description: "slab memory that is reclaimable (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "shared", description: "shared memory including tmpfs (in bytes)"},
					isNeedScale: true},
				{subField: subField{name: "totalHugePages", isProperty: true, description: "total huge pages (huge pages)"}},
				{subField: subField{name: "freeHugePages", description: "free huge pages (huge pages)"}},
			},
		},
		labelPRG: {
			label: "Program", isSystem: false,
			fields: []dataField{
				{subField: subField{name: "pid", description: "PID (unique ID of task),"},
					isNote: true},
				{subField: subField{name: "name", description: "name (between parenthesis or underscores for spaces)"},
					isNote: true},
				{subField: subField{name: "", description: "state"},
					isNote: true},
				{subField: subField{name: "", description: "real uid"},
					isNote: true},
				{subField: subField{name: "", description: "real gid"},
					isNote: true},
				{subField: subField{name: "", description: "TGID (group number of related tasks/threads)"},
					isNote: true},
				{subField: subField{name: "threads", description: "total number of threads"}},
				{subField: subField{name: "exitCode", description: "exit code (in case of fatal signal: signal number + 256)"},
					isNote: true},
				{subField: subField{name: "startTime", description: "start time (epoch)"},
					isNote: true},
				{subField: subField{name: "commandLine", description: "full command line  (between  parenthesis  or underscores for spaces)"},
					isNote: true},
				{subField: subField{name: "ppid", description: "PPID"},
					isNote: true},
				{subField: subField{name: "threadsRun", description: "number of threads in state 'running' (R)"}},
				{subField: subField{name: "threadsSleep", description: "number of threads in state 'interruptible sleeping' (S)"}},
				{subField: subField{name: "threadsDead", description: "number of threads in state 'uninterruptible sleeping' (D)"}},
				{subField: subField{name: "", description: "effective uid"}},
				{subField: subField{name: "", description: "effective gid"}},
				{subField: subField{name: "", description: "saved uid"}},
				{subField: subField{name: "", description: "saved gid"}},
				{subField: subField{name: "", description: "filesystem uid"}},
				{subField: subField{name: "", description: "filesystem gid"}},
				{subField: subField{name: "elapsedTime", description: "elapsed time of terminated process (hertz)"},
					isNote: true},
				{subField: subField{name: "", description: "is_process (y/n)"},
					isProcess: true},
				{subField: subField{name: "", description: "OpenVZ virtual pid (VPID)"}},
				{subField: subField{name: "", description: "OpenVZ container id (CTID)"}},
				{subField: subField{name: "", description: "container/pod name (CID/POD)"}},
				{subField: subField{name: "", description: "indication if the task is newly started during this interval ('N')"}},
				{subField: subField{name: "", description: "cgroup v2 path name (between parenthesis or underscores for spaces)"}},
				{subField: subField{name: "endTime", description: "end time (epoch or 0 if still active)"},
					isNote: true},
				{subField: subField{name: "threadIdle", description: "number of threads in state 'idle' (I)"}},
			},
			subName: pidList.getProgrammPid,
		},
		labelPRC: {
			label: "Process", isSystem: false,
			fields: []dataField{
				{subField: subField{name: "", description: "PID"},
					isSubName: true},
				{subField: subField{name: "", description: "name (between parenthesis or underscores for spaces)"},
					isNote: true},
				{subField: subField{name: "", description: "state"},
					isNote: true},
				{subField: subField{name: "", description: "total number of clock-ticks per second for this machine"},
					isScale: true},
				{subField: subField{name: "CPU_user", description: "CPU-consumption in user mode (clockticks)"},
					isNeedScale: true},
				{subField: subField{name: "CPU_system", description: "CPU-consumption in system mode (clockticks)"},
					isNeedScale: true},
				{subField: subField{name: "", description: "nice value"}},
				{subField: subField{name: "", description: "priority"}},
				{subField: subField{name: "", description: "realtime priority"}},
				{subField: subField{name: "", description: "scheduling policy"}},
				{subField: subField{name: "", description: "current CPU (-1 for exited process)"}},
				{subField: subField{name: "", description: "sleep average"}},
				{subField: subField{name: "", description: "TGID (group number of related tasks/threads)"}},
				{subField: subField{name: "", description: "is_process (y/n)"},
					isProcess: true},
				{subField: subField{name: "", description: "runqueue delay in nanoseconds for this thread or for all threads (in case of process)"}},
				{subField: subField{name: "", description: "wait channel of this thread (between parenthesis or underscores for spaces)"}},
				{subField: subField{name: "blockIO", description: "block I/O delay (clockticks)"},
					isNeedScale: true},
				{subField: subField{name: "", description: "cgroup v2 'cpu.max' calculated as  percentage  (-3  means  no cgroup v2 support, -2 means undefined and -1 means maximum)"}},
				{subField: subField{name: "", description: "cgroup v2 most restrictive 'cpu.max' in upper directories calculated as percentage (-3 means no cgroup v2 support, -2 means undefined and -1 means maximum)"}},
				{subField: subField{name: "contextSwitch", description: "number of voluntary context switches"}},
				{subField: subField{name: "inContextSwitch", description: "number of involuntary context switches"}},
			},
			subName: pidList.getProcessPid,
			scale: func(value float64, intetrval int64, scale float64) float64 {
				return 100 * value / (float64(intetrval) * scale)
			},
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
				{subField: subField{name: "", description: "PID"},
					isNote: true},
				{subField: subField{name: "", description: "name (between parenthesis or underscores for spaces)"},
					isNote: true},
				{subField: subField{name: "", description: "state"},
					isNote: true},
				{subField: subField{name: "", description: "page size for this machine (in bytes)"},
					isScale: true},
				{subField: subField{name: "virtual", description: "virtual  memory  size  (KiB)"}},
				{subField: subField{name: "resident", description: "resident memory  size  (KiB)"}},
				{subField: subField{name: "shared", description: "shared  text  memory size (KiB)"}},
				{subField: subField{name: "virtualGrowth", description: "virtual memory growth (KiB)"}},
				{subField: subField{name: "residentGrowth", description: "resident memory growth (KiB)"}},
				{subField: subField{name: "minorFaults", description: "number of minor page faults"}},
				{subField: subField{name: "majorFaults", description: "number of major page faults"}},
				{subField: subField{name: "virtualLibrary", description: "virtual library exec size (KiB)"}},
				{subField: subField{name: "virtualData", description: "virtual data size (KiB)"}},
				{subField: subField{name: "virtualStach", description: "virtual stack size (KiB)"}},
				{subField: subField{name: "swap", description: "swap space used (KiB)"}},
				{subField: subField{name: "", description: "TGID (group number of related tasks/threads)"}},
				{subField: subField{name: "", description: "is_process (y/n)"},
					isProcess: true},
				{subField: subField{name: "", description: "proportional set size (KiB) if in 'R' option is specified"}},
				{subField: subField{name: "locked", description: "virtually locked memory space (KiB)"}},
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
				{subField: subField{name: "", description: "is_process (y/n)"},
					isProcess: true},
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
				{subField: subField{name: "", description: "is_process (y/n)"},
					isProcess: true},
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

func (obj *dataDescription) isProcess(data dataEntry) bool {

	length := min(len(data.points), len(obj.fields))
	for i := 0; i < length; i++ {
		if obj.fields[i].isProcess {
			switch string(data.points[i]) {
			case "y":
				return true
			case "n":
				return false
			default:
				panic("undefined")
			}
		}
	}

	return true
}

///////////////////////////////////////////////////////////////////////////////

func (obj *dataDescription) getCounters(data dataEntry) (count []keyValue, err error) {

	var scale float64 = 1

	length := min(len(data.points), len(obj.fields))
	count = make([]keyValue, 0, length)

	for i := 0; i < length; i++ {
		var value float64
		var field = obj.fields[i]

		if field.isSubName || field.isNote {
			continue
		}
		if field.name == "" && !field.isScale {
			continue
		}

		if value, err = bytesToFloat64(data.points[i]); err != nil {
			return nil, err
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

		if field.name != "" {
			count = append(count, keyValue{field.name, value})
		}
	}
	for _, field := range obj.counts {
		var value float64

		if value, err = field.counting(data); err != nil {
			return nil, err
		}

		count = append(count, keyValue{field.name, value})
	}

	return count, err
}

func (obj *dataDescription) getNotes(data dataEntry) (note []keyNote, err error) {

	length := min(len(data.points), len(obj.fields))
	note = make([]keyNote, 0, length)

	for i := 0; i < length; i++ {
		var field = obj.fields[i]

		if !field.isNote {
			continue
		}

		if field.name != "" {
			note = append(note, keyNote{field.name, data.points[i]})
		}
	}

	return note, nil
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
	state := string(entry.points[2])
	startTime := string(entry.points[8])

	if oldStartTime, ok := obj.data[pid]; ok {
		if state == "E" {
			startTime = oldStartTime
		}
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
	panic("no pid: " + pid)
	// obj.data[pid] = ""

	return fmt.Sprintf("%s_0", pid)
}
