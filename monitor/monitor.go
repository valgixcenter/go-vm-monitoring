package monitor

import (
	"log"
	"math"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// ProcessInfo holds basic info about a process
type ProcessInfo struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
	MemoryRSS     uint64  `json:"memory_rss"`
}

// SystemStats struct to hold all metric data
type SystemStats struct {
	CPUUsage    float64       `json:"cpu_usage"` // Percentage
	CPUModel    string        `json:"cpu_model"`
	CPUCores    int           `json:"cpu_cores"`
	CPUThreads  int           `json:"cpu_threads"`
	MemoryTotal uint64        `json:"memory_total"`
	MemoryType  string        `json:"memory_type"`
	MemoryUsed  uint64        `json:"memory_used"`
	MemoryUsage float64       `json:"memory_usage"` // Percentage
	DiskTotal   uint64        `json:"disk_total"`
	DiskUsed    uint64        `json:"disk_used"`
	DiskUsage   float64       `json:"disk_usage"` // Percentage
	DiskFstype  string        `json:"disk_fstype"`
	NetInRate   uint64        `json:"net_in_rate"`  // Bytes per second
	NetOutRate  uint64        `json:"net_out_rate"` // Bytes per second
	Processes   []ProcessInfo `json:"processes"`
}

var (
	lastNetStat   []net.IOCountersStat
	lastCheckTime time.Time
)

// GetStats collects current system metrics
func GetStats() (*SystemStats, error) {
	stats := &SystemStats{}
	now := time.Now()

	// CPU Usage
	percent, err := cpu.Percent(0, false)
	if err != nil {
		log.Printf("Error getting CPU percent: %v", err)
	} else if len(percent) > 0 {
		stats.CPUUsage = math.Round(percent[0]*100) / 100
	}

	// CPU Info (Model Name) - Only need to fetch occasionally really, but here every time is fine
	cInfos, err := cpu.Info()
	if err == nil && len(cInfos) > 0 {
		stats.CPUModel = cInfos[0].ModelName
	}

	// CPU Cores/Threads
	physical, _ := cpu.Counts(false)
	logical, _ := cpu.Counts(true)
	stats.CPUCores = physical
	stats.CPUThreads = logical

	// Memory Usage
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error getting memory stats: %v", err)
	} else {
		stats.MemoryTotal = v.Total
		stats.MemoryUsed = v.Used
		stats.MemoryUsage = math.Round(v.UsedPercent*100) / 100
	}

	// Memory Type (Best Effort via dmidecode)
	cmd := exec.Command("dmidecode", "-t", "17")
	out, err := cmd.Output()
	if err != nil {
		stats.MemoryType = "Type: Unknown (Root Req?)"
	} else {
		// Simple parse: look for "Type: DDR..."
		lines := strings.Split(string(out), "\n")
		foundType := ""
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "Type:") {
				val := strings.TrimSpace(strings.TrimPrefix(trimmed, "Type:"))
				if strings.Contains(val, "DDR") {
					foundType = val
					break
				}
				if foundType == "" && val != "Unknown" {
					foundType = val
				}
			}
		}
		if foundType != "" {
			stats.MemoryType = foundType
		} else {
			stats.MemoryType = "Unknown"
		}
	}

	// Disk Usage (Root path)
	d, err := disk.Usage("/")
	if err != nil {
		log.Printf("Error getting disk usage: %v", err)
	} else {
		stats.DiskTotal = d.Total
		stats.DiskUsed = d.Used
		stats.DiskUsage = math.Round(d.UsedPercent*100) / 100
	}

	// Disk Info (Fstype for root)
	partitions, err := disk.Partitions(false)
	if err == nil {
		for _, p := range partitions {
			if p.Mountpoint == "/" {
				stats.DiskFstype = p.Fstype
				break
			}
		}
	}

	// Network I/O Rate (System)
	currentNetStat, err := net.IOCounters(false)
	if err != nil {
		log.Printf("Error getting net stats: %v", err)
	} else if len(currentNetStat) > 0 {
		if !lastCheckTime.IsZero() && len(lastNetStat) > 0 {
			duration := now.Sub(lastCheckTime).Seconds()
			if duration > 0 {
				bytesRecvDiff := currentNetStat[0].BytesRecv - lastNetStat[0].BytesRecv
				bytesSentDiff := currentNetStat[0].BytesSent - lastNetStat[0].BytesSent

				stats.NetInRate = uint64(float64(bytesRecvDiff) / duration)
				stats.NetOutRate = uint64(float64(bytesSentDiff) / duration)
			}
		}
		lastNetStat = currentNetStat
		lastCheckTime = now
	}

	// Process List
	procs, err := process.Processes()
	if err != nil {
		log.Printf("Error getting processes: %v", err)
	} else {
		var processes []ProcessInfo
		for _, p := range procs {
			// Basic info might fail for some processes (permission denied)
			name, err := p.Name()
			if err != nil {
				continue
			}
			cpuP, _ := p.CPUPercent()
			memP, _ := p.MemoryPercent()
			memInfo, _ := p.MemoryInfo()
			var rss uint64
			if memInfo != nil {
				rss = memInfo.RSS
			}

			pInfo := ProcessInfo{
				PID:           p.Pid,
				Name:          name,
				CPUPercent:    math.Round(cpuP*100) / 100,
				MemoryPercent: float32(math.Round(float64(memP)*100) / 100),
				MemoryRSS:     rss,
			}

			// Network Stats per process removed due to missing method in gopsutil/v4

			processes = append(processes, pInfo)
		}

		// Sort by CPU usage descending by default
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].CPUPercent > processes[j].CPUPercent
		})

		// Limit to top 20 to avoid large payload
		if len(processes) > 20 {
			processes = processes[:20]
		}
		stats.Processes = processes
	}

	return stats, nil
}
