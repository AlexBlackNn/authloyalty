package agentmetricsservice

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/AlexBlackNn/metrics/internal/config/configagent"
	"github.com/AlexBlackNn/metrics/internal/domain/models"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type MonitorService struct {
	log     *slog.Logger
	cfg     *configagent.Config
	Metrics map[string]models.MetricInteraction
	mutex   sync.RWMutex
}

func New(
	log *slog.Logger,
	cfg *configagent.Config,
) *MonitorService {
	return &MonitorService{
		Metrics: make(map[string]models.MetricInteraction),
		log:     log,
		cfg:     cfg,
	}
}

// Collect starts collecting runtime metrics.
func (ms *MonitorService) Collect(ctx context.Context) {
	log := ms.log.With(
		slog.String("info", "SERVICE LAYER: agentmetricservice.Start"),
	)

	var rtm runtime.MemStats
	ms.Metrics["PollCount"] = &models.Metric[uint64]{Type: configagent.MetricTypeCounter, Value: uint64(0), Name: "PollCount"}
	for {
		select {
		case <-ctx.Done():
			log.Info("stop signal received")
			return
		case <-time.After(time.Duration(ms.cfg.PollInterval) * time.Second):
			log.Info("starts Collect metric pooling")
			// Read full mem stats
			runtime.ReadMemStats(&rtm)
			ms.mutex.Lock()
			ms.Metrics["Alloc"] = &models.Metric[uint64]{Type: configagent.MetricTypeGauge, Value: rtm.Alloc, Name: "Alloc"}
			ms.mutex.Unlock()
			log.Info("metric pooling finished")
		}
	}
}

// CollectAddition Collect starts collecting gopsutil metrics.
func (ms *MonitorService) CollectAddition(ctx context.Context) {
	log := ms.log.With(
		slog.String("info", "SERVICE LAYER: agentmetricservice.Start"),
	)

	virtMem, err := mem.VirtualMemory()
	if err != nil {
		log.Error(err.Error())
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("stop signal received")
			return
		case <-time.After(time.Duration(ms.cfg.PollInterval) * time.Second):
			ms.mutex.Lock()
			log.Info("starts CollectAddingMetrics metrics pooling")

			utilCPU, err := ms.calculateUtilization()
			if err != nil {
				log.Error(err.Error())
				continue
			}
			ms.Metrics["CPUutilization1"] = &models.Metric[float64]{Type: configagent.MetricTypeGauge, Value: utilCPU, Name: "CPUutilization1"}
			ms.Metrics["TotalMemory"] = &models.Metric[uint64]{Type: configagent.MetricTypeGauge, Value: virtMem.Total, Name: "TotalMemory"}
			ms.Metrics["FreeMemory"] = &models.Metric[uint64]{Type: configagent.MetricTypeGauge, Value: virtMem.Available, Name: "FreeMemory"}
			log.Info("metric pooling finished")
			ms.mutex.Unlock()
		}
	}
}

// GetMetrics return collected metrics as thread safe map.
func (ms *MonitorService) GetMetrics() map[string]models.MetricInteraction {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	return ms.Metrics
}

func (ms *MonitorService) calculateUtilization() (float64, error) {
	// get available cpu
	numCPUs := runtime.NumCPU()

	// Get cpu loading statistic
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}

	// calculate average CPU loading
	totalPercent := 0.0
	for _, percent := range cpuPercent {
		totalPercent += percent
	}
	return totalPercent / float64(numCPUs), nil
}
