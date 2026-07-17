package collector

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"

	"github.com/z9wen/VPSentinel/protocol"
)

// Collector 隔离“采集指标”与“发送指标”。测试其他组件时可以替换成 fake。
type Collector interface {
	Collect(context.Context) (protocol.MetricSample, error)
}

type HostCollector struct {
	diskPath string
}

func NewHostCollector(diskPath string) *HostCollector {
	return &HostCollector{diskPath: diskPath}
}

func (c *HostCollector) Collect(ctx context.Context) (protocol.MetricSample, error) {
	sample := protocol.MetricSample{RecordedAt: time.Now().UTC()}
	var collectionErrors []error

	recordFailure := func(name string, err error) {
		sample.FailedCollectors = append(sample.FailedCollectors, name)
		collectionErrors = append(collectionErrors, fmt.Errorf("collect %s: %w", name, err))
	}

	// 200ms 是 CPU 采样窗口。它是 context-aware 的，程序退出时不会无故阻塞。
	cpuPercent, err := safeCall(func() ([]float64, error) {
		return cpu.PercentWithContext(ctx, 200*time.Millisecond, false)
	})
	if err != nil {
		recordFailure("cpu", err)
	} else if len(cpuPercent) == 0 {
		recordFailure("cpu", errors.New("no values returned"))
	} else {
		sample.CPUPercent = pointer(cpuPercent[0])
	}

	memory, err := safeCall(func() (*mem.VirtualMemoryStat, error) {
		return mem.VirtualMemoryWithContext(ctx)
	})
	if err != nil {
		recordFailure("memory", err)
	} else {
		sample.MemoryUsedBytes = pointer(memory.Used)
		sample.MemoryTotalBytes = pointer(memory.Total)
	}

	diskUsage, err := safeCall(func() (*disk.UsageStat, error) {
		return disk.UsageWithContext(ctx, c.diskPath)
	})
	if err != nil {
		recordFailure("disk", fmt.Errorf("path %q: %w", c.diskPath, err))
	} else {
		sample.DiskUsedBytes = pointer(diskUsage.Used)
		sample.DiskTotalBytes = pointer(diskUsage.Total)
	}

	loadAverage, err := safeCall(func() (*load.AvgStat, error) {
		return load.AvgWithContext(ctx)
	})
	if err != nil {
		recordFailure("load", err)
	} else {
		sample.Load1 = pointer(loadAverage.Load1)
		sample.Load5 = pointer(loadAverage.Load5)
		sample.Load15 = pointer(loadAverage.Load15)
	}

	hostInfo, err := safeCall(func() (*host.InfoStat, error) {
		return host.InfoWithContext(ctx)
	})
	if err != nil {
		recordFailure("host", err)
	} else {
		sample.UptimeSeconds = pointer(hostInfo.Uptime)
		sample.OSName = hostInfo.OS
		sample.KernelVersion = hostInfo.KernelVersion
	}

	interfaces, err := safeCall(func() ([]net.IOCountersStat, error) {
		return net.IOCountersWithContext(ctx, true)
	})
	if err != nil {
		recordFailure("network", err)
	} else {
		var rxBytes, txBytes uint64
		for _, networkInterface := range interfaces {
			rxBytes += networkInterface.BytesRecv
			txBytes += networkInterface.BytesSent
		}
		sample.NetworkRXBytes = pointer(rxBytes)
		sample.NetworkTXBytes = pointer(txBytes)
	}

	return sample, errors.Join(collectionErrors...)
}

func pointer[T any](value T) *T {
	return &value
}

// safeCall prevents a bug in one platform-specific third-party collector from
// terminating the whole monitoring agent.
func safeCall[T any](call func() (T, error)) (value T, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("collector panic: %v", recovered)
		}
	}()
	return call()
}
