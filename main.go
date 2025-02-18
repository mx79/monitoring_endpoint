package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type ServerStats struct {
	CPUUsage  float64 `json:"cpu_usage"`
	RAMUsage  float64 `json:"ram_usage"`
	DiskUsage float64 `json:"disk_usage"`
}

func getCPUUsage() float64 {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		log.Printf("Failed to read CPU load: %v", err)
		return 0.0
	}

	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		log.Printf("Unexpected format in /proc/loadavg")
		return 0.0
	}

	value, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		log.Printf("Failed to parse CPU load: %v", err)
		return 0.0
	}

	return value * 10
}

func getRAMUsage() float64 {
	memUsagePath := "/sys/fs/cgroup/memory.current"
	memLimitPath := "/sys/fs/cgroup/memory.max"

	memLimitBytes, err := os.ReadFile(memLimitPath)
	if err != nil {
		log.Printf("Failed to read memory limit: %v", err)
		return 0.0
	}

	memUsageBytes, err := os.ReadFile(memUsagePath)
	if err != nil {
		log.Printf("Failed to read memory usage: %v", err)
		return 0.0
	}

	memLimit, _ := strconv.ParseFloat(strings.TrimSpace(string(memLimitBytes)), 64)
	memUsage, _ := strconv.ParseFloat(strings.TrimSpace(string(memUsageBytes)), 64)

	if memLimit == 0 {
		return 0.0
	}
	return (memUsage / memLimit) * 100.0
}

func getDiskUsage() float64 {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		log.Printf("Failed to retrieve disk usage: %v", err)
		return 0.0
	}

	total := float64(stat.Blocks * uint64(stat.Bsize))
	used := float64((stat.Blocks - stat.Bfree) * uint64(stat.Bsize))

	return (used / total) * 100.0
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := ServerStats{
		CPUUsage:  getCPUUsage(),
		RAMUsage:  getRAMUsage(),
		DiskUsage: getDiskUsage(),
	}

	json.NewEncoder(w).Encode(stats)
}

func main() {
	http.HandleFunc("/stats", statsHandler)
	fmt.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
