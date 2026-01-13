# RX 7900 XT GPU Optimization Plan for Local LLM Inference

## Overview

Optimize the RX 7900 XT eGPU setup for maximum LLM inference performance on the NucBox K8.

## Current Setup Status

- Hardware: RX 7900 XT (20GB VRAM, ~800 GB/s bandwidth) via OcuLink
- Software: ROCm 7.1.1, Ollama working
- Baseline VRAM: ~11% used by desktop compositor (3 displays)

---

## 1. GPU Power Profile Optimization

### What it does
Forces the GPU to run at maximum clocks instead of dynamically scaling. By default, the GPU enters low-power states when idle, which can cause latency spikes when inference starts.

### Implementation
```bash
# Check current setting
cat /sys/class/drm/card1/device/power_dpm_force_performance_level

# Options: auto, low, high, manual, profile_standard, profile_peak
# Set to high for consistent performance
echo "high" | sudo tee /sys/class/drm/card1/device/power_dpm_force_performance_level
```

### Make persistent (survives reboot)
Create `/etc/systemd/system/gpu-performance.service`:
```ini
[Unit]
Description=Set GPU to high performance mode
After=multi-user.target

[Service]
Type=oneshot
ExecStart=/bin/bash -c 'echo high > /sys/class/drm/card1/device/power_dpm_force_performance_level'
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
```

Then enable:
```bash
sudo systemctl enable gpu-performance.service
```

### Trade-off
- Higher idle power consumption (~50W vs ~10W)
- More heat when idle
- Faster response when inference starts

---

## 2. ROCm Environment Variables

### HSA_OVERRIDE_GFX_VERSION
Forces ROCm to treat the GPU as a specific architecture. Sometimes needed for compatibility with newer RDNA 3 GPUs.

```bash
export HSA_OVERRIDE_GFX_VERSION=11.0.0
```
- `11.0.0` = gfx1100 (Navi 31 / RX 7900 series)
- Usually not needed if rocminfo already shows the GPU, but can fix edge cases

### GPU_MAX_HW_QUEUES
Controls how many hardware queues are used for GPU work submission. More queues can improve throughput for parallel workloads.

```bash
export GPU_MAX_HW_QUEUES=8
```
- Default is typically 4
- Higher values may help with concurrent operations
- Diminishing returns above 8

### PYTORCH_ROCM_ARCH
Tells PyTorch which GPU architecture to target for JIT compilation.

```bash
export PYTORCH_ROCM_ARCH=gfx1100
```
- Ensures optimal code generation for your GPU
- Only matters for PyTorch workloads (not Ollama/llama.cpp)

### Add to ~/.zshrc
```bash
# ROCm optimizations for RX 7900 XT
export HSA_OVERRIDE_GFX_VERSION=11.0.0
export GPU_MAX_HW_QUEUES=8
export PYTORCH_ROCM_ARCH=gfx1100
```

---

## 3. Ollama Configuration

### Config file: ~/.ollama/config.json
```json
{
  "num_gpu": 1,
  "num_thread": 8,
  "gpu_memory_utilization": 0.9
}
```

### Environment variables for Ollama
```bash
# Context window size (tokens) - larger = more VRAM
export OLLAMA_NUM_CTX=8192

# Number of layers to offload to GPU (99 = all)
export OLLAMA_NUM_GPU=99

# Keep model in memory (seconds) - 0 = unload immediately
export OLLAMA_KEEP_ALIVE=300
```

### Add to ~/.zshrc for persistent config
```bash
# Ollama optimizations
export OLLAMA_NUM_CTX=8192
export OLLAMA_NUM_GPU=99
export OLLAMA_KEEP_ALIVE=300
```

---

## 4. Memory and System Optimizations

### Increase GPU memory allocation priority
```bash
# Check current
cat /sys/module/amdgpu/parameters/vm_fragment_size

# Larger fragments = less fragmentation for large allocations
# Default is 9 (512KB), try 12 (4MB) for LLM workloads
echo 12 | sudo tee /sys/module/amdgpu/parameters/vm_fragment_size
```

### Disable GPU reset on timeout (prevents interruptions)
Add to `/etc/modprobe.d/amdgpu.conf`:
```
options amdgpu gpu_recovery=0
```

### Increase locked memory limit
Add to `/etc/security/limits.conf`:
```
*    soft    memlock    unlimited
*    hard    memlock    unlimited
```

---

## 5. Monitoring and Verification

### Watch GPU stats during inference
```bash
watch -n 1 rocm-smi
```

### Check if GPU is being used
```bash
radeontop -b 03
```

### Verify Ollama is using GPU
```bash
# While model is running
rocm-smi --showmemuse
```

### Monitor tokens/s

**Ollama verbose mode (easiest):**
```bash
ollama run qwen2.5-coder:7b --verbose
# Shows eval rate after each response
```

**API with timing stats:**
```bash
curl http://localhost:11434/api/generate -d '{
  "model": "qwen2.5-coder:7b",
  "prompt": "Write hello world in Python",
  "stream": false
}' | jq '.eval_duration, .eval_count'
```

**oterm — Terminal UI for Ollama:**
```bash
pipx install oterm
oterm
```

**Open WebUI — Web dashboard:**
```bash
docker run -d -p 3000:8080 \
  --add-host=host.docker.internal:host-gateway \
  -v open-webui:/app/backend/data \
  --name open-webui \
  ghcr.io/open-webui/open-webui:main
```
Access at `http://localhost:3000`

---

## Files to Modify

| File | Changes |
|------|---------|
| `~/.zshrc` | Add ROCm + Ollama environment variables |
| `/etc/systemd/system/gpu-performance.service` | Create for persistent high-perf mode |
| `~/.ollama/config.json` | Create Ollama tuning config |
| `/etc/modprobe.d/amdgpu.conf` | Optional GPU driver params |
| `/etc/security/limits.conf` | Increase memlock limits |

---

## Verification Steps

1. Reboot after changes
2. Run `rocminfo | grep -i marketing` — should show RX 7900 XT
3. Run `cat /sys/class/drm/card1/device/power_dpm_force_performance_level` — should show `high`
4. Run `ollama run qwen2.5-coder:7b` and verify VRAM usage with `rocm-smi`
5. Compare inference speed before/after with a standard prompt

---

## Expected Impact

| Optimization | Impact | Effort |
|--------------|--------|--------|
| Power profile = high | Medium (faster first-token) | Low |
| HSA_OVERRIDE_GFX_VERSION | Low (compatibility fix) | Low |
| OLLAMA_NUM_CTX | Low (more context = more VRAM) | Low |
| memlock limits | Low (prevents memory errors) | Low |
| GPU fragment size | Low (reduces fragmentation) | Medium |

---

## Document Updates

After implementation, update these files in `spec/todo/`:
- `gpu-docker-llm-setup.ticket.md` — add optimization section
- `coding-model-comparison.md` — add optimized performance numbers
