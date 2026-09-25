package model

import "time"

type SystemInfo struct {
	Hostname         string    `json:"hostname" yaml:"hostname"`
	OS               string    `json:"os" yaml:"os"`
	Version          string    `json:"version" yaml:"version"`
	Architecture     string    `json:"architecture" yaml:"architecture"`
	UptimeSeconds    uint64    `json:"uptime_seconds" yaml:"uptime_seconds"`
	CPUCount         int       `json:"cpu_count" yaml:"cpu_count"`
	CPUUsagePct      float64   `json:"cpu_usage_pct" yaml:"cpu_usage_pct"`
	MemoryTotalBytes uint64    `json:"memory_total_bytes" yaml:"memory_total_bytes"`
	MemoryUsedBytes  uint64    `json:"memory_used_bytes" yaml:"memory_used_bytes"`
	StorageTotal     uint64    `json:"storage_total_bytes" yaml:"storage_total_bytes"`
	StorageUsed      uint64    `json:"storage_used_bytes" yaml:"storage_used_bytes"`
	SerialNumber     string    `json:"serial_number" yaml:"serial_number"`
	Time             time.Time `json:"time" yaml:"time"`
}

type MetricSample struct {
	MetricName string            `json:"metric_name" yaml:"metric_name"`
	Timestamp  time.Time         `json:"timestamp" yaml:"timestamp"`
	Value      float64           `json:"value" yaml:"value"`
	Labels     map[string]string `json:"labels" yaml:"labels"`
}
