package gpu

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// GPUInfo represents detailed information about a GPU
type GPUInfo struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Vendor       string  `json:"vendor"`
	Model        string  `json:"model"`
	VRAMTotal    uint64  `json:"vram_total_mb"`
	VRAMFree     uint64  `json:"vram_free_mb"`
	VRAMUsed     uint64  `json:"vram_used_mb"`
	PowerDraw    uint32  `json:"power_draw_w"`
	PowerLimit   uint32  `json:"power_limit_w"`
	Temperature  uint8   `json:"temperature_c"`
	Utilization  uint8   `json:"utilization_percent"`
	DriverVersion string `json:"driver_version"`
	CUDAVersion  string  `json:"cuda_version,omitempty"`
	ComputeCapability string `json:"compute_capability,omitempty"`
	PCIBusID     string  `json:"pci_bus_id"`
	IsAvailable  bool    `json:"is_available"`
}

// GPUMetrics represents real-time GPU metrics
type GPUMetrics struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	Utilization  uint8     `json:"utilization_percent"`
	VRAMUsed     uint64    `json:"vram_used_mb"`
	VRAMTotal    uint64    `json:"vram_total_mb"`
	PowerDraw    uint32    `json:"power_draw_w"`
	Temperature  uint8     `json:"temperature_c"`
	FanSpeed     uint8     `json:"fan_speed_percent"`
	ClockCore    uint32    `json:"clock_core_mhz"`
	ClockMemory  uint32    `json:"clock_memory_mhz"`
}

// Detector handles GPU detection and monitoring
type Detector struct {
	logger *zap.Logger
	gpus   []GPUInfo
}

// NewDetector creates a new GPU detector
func NewDetector(logger *zap.Logger) *Detector {
	return &Detector{
		logger: logger,
		gpus:   make([]GPUInfo, 0),
	}
}

// DetectGPUs detects all available GPUs on the system
func (d *Detector) DetectGPUs(ctx context.Context) ([]GPUInfo, error) {
	d.logger.Info("Starting GPU detection")

	var allGPUs []GPUInfo

	// Detect NVIDIA GPUs
	nvidiaGPUs, err := d.detectNVIDIAGPUs(ctx)
	if err != nil {
		d.logger.Warn("Failed to detect NVIDIA GPUs", zap.Error(err))
	} else {
		allGPUs = append(allGPUs, nvidiaGPUs...)
	}

	// Detect AMD GPUs
	amdGPUs, err := d.detectAMDGPUs(ctx)
	if err != nil {
		d.logger.Warn("Failed to detect AMD GPUs", zap.Error(err))
	} else {
		allGPUs = append(allGPUs, amdGPUs...)
	}

	// Detect Apple Silicon GPUs
	if runtime.GOOS == "darwin" {
		appleGPUs, err := d.detectAppleGPUs(ctx)
		if err != nil {
			d.logger.Warn("Failed to detect Apple GPUs", zap.Error(err))
		} else {
			allGPUs = append(allGPUs, appleGPUs...)
		}
	}

	// Detect Intel GPUs
	intelGPUs, err := d.detectIntelGPUs(ctx)
	if err != nil {
		d.logger.Warn("Failed to detect Intel GPUs", zap.Error(err))
	} else {
		allGPUs = append(allGPUs, intelGPUs...)
	}

	d.gpus = allGPUs
	d.logger.Info("GPU detection completed", zap.Int("gpu_count", len(allGPUs)))

	return allGPUs, nil
}

// detectNVIDIAGPUs detects NVIDIA GPUs using nvidia-smi
func (d *Detector) detectNVIDIAGPUs(ctx context.Context) ([]GPUInfo, error) {
	// Check if nvidia-smi is available
	if !d.isCommandAvailable("nvidia-smi") {
		return nil, fmt.Errorf("nvidia-smi not found")
	}

	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=index,name,memory.total,memory.free,memory.used,power.draw,power.limit,temperature.gpu,utilization.gpu,driver_version,pci.bus_id", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run nvidia-smi: %w", err)
	}

	var gpus []GPUInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, ", ")
		if len(fields) < 11 {
			continue
		}

		gpu := GPUInfo{
			Vendor:      "NVIDIA",
			IsAvailable: true,
		}

		// Parse fields
		if id := strings.TrimSpace(fields[0]); id != "" {
			gpu.ID = fmt.Sprintf("nvidia-%s", id)
		}

		gpu.Name = strings.TrimSpace(fields[1])
		gpu.Model = gpu.Name

		if vramTotal, err := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 64); err == nil {
			gpu.VRAMTotal = vramTotal
		}

		if vramFree, err := strconv.ParseUint(strings.TrimSpace(fields[3]), 10, 64); err == nil {
			gpu.VRAMFree = vramFree
		}

		if vramUsed, err := strconv.ParseUint(strings.TrimSpace(fields[4]), 10, 64); err == nil {
			gpu.VRAMUsed = vramUsed
		}

		if powerDraw, err := strconv.ParseFloat(strings.TrimSpace(fields[5]), 32); err == nil {
			gpu.PowerDraw = uint32(powerDraw)
		}

		if powerLimit, err := strconv.ParseFloat(strings.TrimSpace(fields[6]), 32); err == nil {
			gpu.PowerLimit = uint32(powerLimit)
		}

		if temp, err := strconv.ParseUint(strings.TrimSpace(fields[7]), 10, 8); err == nil {
			gpu.Temperature = uint8(temp)
		}

		if util, err := strconv.ParseUint(strings.TrimSpace(fields[8]), 10, 8); err == nil {
			gpu.Utilization = uint8(util)
		}

		gpu.DriverVersion = strings.TrimSpace(fields[9])
		gpu.PCIBusID = strings.TrimSpace(fields[10])

		// Get CUDA version
		if cudaVersion, err := d.getCUDAVersion(ctx); err == nil {
			gpu.CUDAVersion = cudaVersion
		}

		// Get compute capability
		if computeCap, err := d.getComputeCapability(ctx, gpu.ID); err == nil {
			gpu.ComputeCapability = computeCap
		}

		gpus = append(gpus, gpu)
	}

	return gpus, nil
}

// detectAMDGPUs detects AMD GPUs using rocm-smi
func (d *Detector) detectAMDGPUs(ctx context.Context) ([]GPUInfo, error) {
	// Check if rocm-smi is available
	if !d.isCommandAvailable("rocm-smi") {
		return nil, fmt.Errorf("rocm-smi not found")
	}

	cmd := exec.CommandContext(ctx, "rocm-smi", "--showid", "--showproductname", "--showmeminfo", "vram", "--showpower", "--showtemp", "--showuse")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run rocm-smi: %w", err)
	}

	var gpus []GPUInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	// Parse rocm-smi output (format varies, this is a simplified parser)
	for _, line := range lines {
		if strings.Contains(line, "GPU") && strings.Contains(line, "card") {
			gpu := GPUInfo{
				Vendor:      "AMD",
				IsAvailable: true,
			}

			// Extract GPU ID
			re := regexp.MustCompile(`card(\d+)`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				gpu.ID = fmt.Sprintf("amd-%s", matches[1])
			}

			// This is a simplified implementation
			// In production, you would need more sophisticated parsing
			gpu.Name = "AMD GPU"
			gpu.Model = "AMD GPU"

			gpus = append(gpus, gpu)
		}
	}

	return gpus, nil
}

// detectAppleGPUs detects Apple Silicon GPUs
func (d *Detector) detectAppleGPUs(ctx context.Context) ([]GPUInfo, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("Apple GPU detection only supported on macOS")
	}

	// Use system_profiler to get GPU information
	cmd := exec.CommandContext(ctx, "system_profiler", "SPDisplaysDataType", "-json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run system_profiler: %w", err)
	}

	var profilerData map[string]interface{}
	if err := json.Unmarshal(output, &profilerData); err != nil {
		return nil, fmt.Errorf("failed to parse system_profiler output: %w", err)
	}

	var gpus []GPUInfo

	// Parse the system profiler data
	if displays, ok := profilerData["SPDisplaysDataType"].([]interface{}); ok {
		for i, display := range displays {
			if displayMap, ok := display.(map[string]interface{}); ok {
				gpu := GPUInfo{
					ID:          fmt.Sprintf("apple-%d", i),
					Vendor:      "Apple",
					IsAvailable: true,
				}

				if name, ok := displayMap["_name"].(string); ok {
					gpu.Name = name
					gpu.Model = name
				}

				// Apple Silicon GPUs share system memory
				// We'll estimate based on the chip type
				if strings.Contains(gpu.Name, "M1") {
					if strings.Contains(gpu.Name, "Ultra") {
						gpu.VRAMTotal = 128 * 1024 // 128GB unified memory
					} else if strings.Contains(gpu.Name, "Max") {
						gpu.VRAMTotal = 64 * 1024 // 64GB unified memory
					} else {
						gpu.VRAMTotal = 16 * 1024 // 16GB unified memory
					}
				} else if strings.Contains(gpu.Name, "M2") {
					if strings.Contains(gpu.Name, "Ultra") {
						gpu.VRAMTotal = 192 * 1024 // 192GB unified memory
					} else if strings.Contains(gpu.Name, "Max") {
						gpu.VRAMTotal = 96 * 1024 // 96GB unified memory
					} else {
						gpu.VRAMTotal = 24 * 1024 // 24GB unified memory
					}
				} else if strings.Contains(gpu.Name, "M3") {
					if strings.Contains(gpu.Name, "Ultra") {
						gpu.VRAMTotal = 256 * 1024 // 256GB unified memory
					} else if strings.Contains(gpu.Name, "Max") {
						gpu.VRAMTotal = 128 * 1024 // 128GB unified memory
					} else {
						gpu.VRAMTotal = 32 * 1024 // 32GB unified memory
					}
				}

				gpu.VRAMFree = gpu.VRAMTotal // Simplified - would need actual memory usage
				gpu.VRAMUsed = 0

				gpus = append(gpus, gpu)
			}
		}
	}

	return gpus, nil
}

// detectIntelGPUs detects Intel GPUs
func (d *Detector) detectIntelGPUs(ctx context.Context) ([]GPUInfo, error) {
	// Check if intel_gpu_top is available
	if !d.isCommandAvailable("intel_gpu_top") {
		return nil, fmt.Errorf("intel_gpu_top not found")
	}

	// This is a simplified implementation
	// Intel GPU detection would require more sophisticated tools
	var gpus []GPUInfo

	// For now, just detect if Intel GPU tools are available
	gpu := GPUInfo{
		ID:          "intel-0",
		Name:        "Intel GPU",
		Model:       "Intel GPU",
		Vendor:      "Intel",
		IsAvailable: true,
	}

	gpus = append(gpus, gpu)

	return gpus, nil
}

// GetMetrics gets real-time metrics for all GPUs
func (d *Detector) GetMetrics(ctx context.Context) ([]GPUMetrics, error) {
	var allMetrics []GPUMetrics

	// Get NVIDIA metrics
	nvidiaMetrics, err := d.getNVIDIAMetrics(ctx)
	if err != nil {
		d.logger.Debug("Failed to get NVIDIA metrics", zap.Error(err))
	} else {
		allMetrics = append(allMetrics, nvidiaMetrics...)
	}

	// Get AMD metrics
	amdMetrics, err := d.getAMDMetrics(ctx)
	if err != nil {
		d.logger.Debug("Failed to get AMD metrics", zap.Error(err))
	} else {
		allMetrics = append(allMetrics, amdMetrics...)
	}

	// Get Apple metrics
	if runtime.GOOS == "darwin" {
		appleMetrics, err := d.getAppleMetrics(ctx)
		if err != nil {
			d.logger.Debug("Failed to get Apple metrics", zap.Error(err))
		} else {
			allMetrics = append(allMetrics, appleMetrics...)
		}
	}

	return allMetrics, nil
}

// Helper methods

// isCommandAvailable checks if a command is available in PATH
func (d *Detector) isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// getCUDAVersion gets the CUDA version
func (d *Detector) getCUDAVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getComputeCapability gets the compute capability for an NVIDIA GPU
func (d *Detector) getComputeCapability(ctx context.Context, gpuID string) (string, error) {
	// This would require nvidia-ml-py or similar tools
	// For now, return a placeholder
	return "8.6", nil
}

// getNVIDIAMetrics gets real-time metrics for NVIDIA GPUs
func (d *Detector) getNVIDIAMetrics(ctx context.Context) ([]GPUMetrics, error) {
	if !d.isCommandAvailable("nvidia-smi") {
		return nil, fmt.Errorf("nvidia-smi not available")
	}

	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=index,utilization.gpu,memory.used,memory.total,power.draw,temperature.gpu,fan.speed,clocks.gr,clocks.mem", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var metrics []GPUMetrics
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, ", ")
		if len(fields) < 9 {
			continue
		}

		metric := GPUMetrics{
			Timestamp: time.Now().UTC(),
		}

		if id := strings.TrimSpace(fields[0]); id != "" {
			metric.ID = fmt.Sprintf("nvidia-%s", id)
		}

		if util, err := strconv.ParseUint(strings.TrimSpace(fields[1]), 10, 8); err == nil {
			metric.Utilization = uint8(util)
		}

		if vramUsed, err := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 64); err == nil {
			metric.VRAMUsed = vramUsed
		}

		if vramTotal, err := strconv.ParseUint(strings.TrimSpace(fields[3]), 10, 64); err == nil {
			metric.VRAMTotal = vramTotal
		}

		if powerDraw, err := strconv.ParseFloat(strings.TrimSpace(fields[4]), 32); err == nil {
			metric.PowerDraw = uint32(powerDraw)
		}

		if temp, err := strconv.ParseUint(strings.TrimSpace(fields[5]), 10, 8); err == nil {
			metric.Temperature = uint8(temp)
		}

		if fanSpeed, err := strconv.ParseUint(strings.TrimSpace(fields[6]), 10, 8); err == nil {
			metric.FanSpeed = uint8(fanSpeed)
		}

		if clockCore, err := strconv.ParseUint(strings.TrimSpace(fields[7]), 10, 32); err == nil {
			metric.ClockCore = uint32(clockCore)
		}

		if clockMemory, err := strconv.ParseUint(strings.TrimSpace(fields[8]), 10, 32); err == nil {
			metric.ClockMemory = uint32(clockMemory)
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// getAMDMetrics gets real-time metrics for AMD GPUs
func (d *Detector) getAMDMetrics(ctx context.Context) ([]GPUMetrics, error) {
	// Simplified implementation for AMD GPUs
	return []GPUMetrics{}, nil
}

// getAppleMetrics gets real-time metrics for Apple GPUs
func (d *Detector) getAppleMetrics(ctx context.Context) ([]GPUMetrics, error) {
	// Simplified implementation for Apple GPUs
	return []GPUMetrics{}, nil
}
