// internal/monitoring/sysmon.go
package monitoring

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SystemStats struct {
	CollectedAt       time.Time     `json:"collected_at"`
	ProcessUptimeSecs float64       `json:"process_uptime_secs"`
	Stale             bool          `json:"stale"` // true if collector goroutine appears stuck
	CPU               CPUStat       `json:"cpu"`
	Memory            MemStat       `json:"memory"`
	Process           ProcessStat   `json:"process"`
	GoRuntime         GoRuntimeStat `json:"go_runtime"`
}

type CPUStat struct {
	UsedPercent float64 `json:"used_percent"`
	NumCPU      int     `json:"num_cpu"`
}

type MemStat struct {
	TotalBytes  uint64  `json:"total_bytes"`
	UsedBytes   uint64  `json:"used_bytes"`
	FreeBytes   uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

type ProcessStat struct {
	PID      int    `json:"pid"`
	RSSBytes uint64 `json:"rss_bytes"`
	Threads  int    `json:"threads"`
}

type GoRuntimeStat struct {
	Goroutines  int     `json:"goroutines"`
	HeapAllocMB float64 `json:"heap_alloc_mb"`
	GCCycles    uint32  `json:"gc_cycles"`
}

type rawCPU struct {
	user, nice, system, idle, iowait, irq, softirq uint64
	total                                          uint64
}

type SysMonitor struct {
	mu         sync.RWMutex
	latest     SystemStats
	interval   time.Duration
	staleAfter time.Duration // flag as stale after this long without an update
	startedAt  time.Time
	prevCPU    rawCPU
}

func NewSysMonitor(interval time.Duration) *SysMonitor {
	return &SysMonitor{
		interval:   interval,
		staleAfter: interval * 3, // e.g. 6s if interval is 2s
		startedAt:  time.Now(),
	}
}

// Run starts the collection loop — call as `go mon.Run(ctx)`
func (s *SysMonitor) Run(ctx interface{ Done() <-chan struct{} }) {
	// Warm up CPU baseline before first real read
	s.prevCPU, _ = readRawCPU()
	time.Sleep(s.interval)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snap := s.collect()
			s.mu.Lock()
			s.latest = snap
			s.mu.Unlock()
		}
	}
}

// GetStats returns the latest snapshot. The Stale field is computed at call
// time so it reflects how long ago the collector last wrote.
func (s *SysMonitor) GetStats() SystemStats {
	s.mu.RLock()
	snap := s.latest
	s.mu.RUnlock()

	snap.Stale = time.Since(snap.CollectedAt) > s.staleAfter
	return snap
}

// IsStale is a lightweight check — useful for health-check endpoints.
func (s *SysMonitor) IsStale() bool {
	s.mu.RLock()
	t := s.latest.CollectedAt
	s.mu.RUnlock()
	return time.Since(t) > s.staleAfter
}

func (s *SysMonitor) collect() SystemStats {
	snap := SystemStats{
		CollectedAt:       time.Now(),
		ProcessUptimeSecs: time.Since(s.startedAt).Seconds(),
	}

	if curr, err := readRawCPU(); err == nil {
		snap.CPU = CPUStat{
			UsedPercent: cpuPercent(s.prevCPU, curr),
			NumCPU:      runtime.NumCPU(),
		}
		s.prevCPU = curr
	}

	if m, err := readMem(); err == nil {
		snap.Memory = m
	}
	if ps, err := readProc(); err == nil {
		snap.Process = ps
	}

	snap.GoRuntime = readGoRT()

	return snap
}

// ── /proc readers ─────────────────────────────────────────────────────────────

func readRawCPU() (rawCPU, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return rawCPU{}, err
	}
	defer f.Close()

	var c rawCPU
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		parse := func(i int) uint64 {
			if i >= len(fields) {
				return 0
			}
			v, _ := strconv.ParseUint(fields[i], 10, 64)
			return v
		}
		c.user, c.nice, c.system, c.idle = parse(1), parse(2), parse(3), parse(4)
		c.iowait, c.irq, c.softirq = parse(5), parse(6), parse(7)
		c.total = c.user + c.nice + c.system + c.idle + c.iowait + c.irq + c.softirq
		return c, nil
	}
	return c, fmt.Errorf("cpu line not found")
}

func cpuPercent(prev, curr rawCPU) float64 {
	td := float64(curr.total - prev.total)
	id := float64(curr.idle - prev.idle)
	if td == 0 {
		return 0
	}
	return (1 - id/td) * 100
}

func readMem() (MemStat, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemStat{}, err
	}
	defer f.Close()

	var m MemStat
	avail := uint64(0)
	kv := map[string]*uint64{
		"MemTotal":     &m.TotalBytes,
		"MemAvailable": &avail,
		"MemFree":      &m.FreeBytes,
	}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		parts := strings.Fields(sc.Text())
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		if ptr, ok := kv[key]; ok {
			v, _ := strconv.ParseUint(parts[1], 10, 64)
			*ptr = v * 1024
		}
	}

	m.UsedBytes = m.TotalBytes - avail
	if m.TotalBytes > 0 {
		m.UsedPercent = float64(m.UsedBytes) / float64(m.TotalBytes) * 100
	}
	return m, nil
}

func readProc() (ProcessStat, error) {
	pid := os.Getpid()
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return ProcessStat{PID: pid}, err
	}
	defer f.Close()

	ps := ProcessStat{PID: pid}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		parts := strings.Fields(sc.Text())
		if len(parts) < 2 {
			continue
		}
		v, _ := strconv.ParseUint(parts[1], 10, 64)
		switch strings.TrimSuffix(parts[0], ":") {
		case "VmRSS":
			ps.RSSBytes = v * 1024
		case "Threads":
			ps.Threads = int(v)
		}
	}
	return ps, nil
}

func readGoRT() GoRuntimeStat {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return GoRuntimeStat{
		Goroutines:  runtime.NumGoroutine(),
		HeapAllocMB: float64(ms.HeapAlloc) / 1024 / 1024,
		GCCycles:    ms.NumGC,
	}
}
