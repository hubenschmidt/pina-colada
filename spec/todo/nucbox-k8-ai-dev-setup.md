# NucBox K8 Plus AI Development Setup Plan

## Hardware Specs
- **CPU:** AMD Ryzen 7 8845HS (8c/16t, Zen 4)
- **RAM:** 64GB DDR5
- **iGPU:** Radeon 780M (RDNA 3, 12 CUs)
- **NPU:** Ryzen AI (16 TOPS)
- **Optional eGPU:** RTX 5070 via Thunderbolt 4

---

## Phase 1: Base OS & Environment

### OS Choice: Ubuntu 24.04 LTS (Native Install)

**Why native Linux over Windows + WSL2:**

| Factor | Windows + WSL2 | Ubuntu Native |
|--------|----------------|---------------|
| RAM available for models | ~50-55GB | ~60-62GB |
| Docker overhead | Double virtualization | Native containers |
| eGPU support | Experimental, problematic | Clean TB4 authorization |
| GPU drivers | Passthrough complexity | Direct install |
| Disk I/O | Slower (VHD layer) | Native speed |

**Bottom line:** For AI development with local models, native Linux maximizes your 64GB RAM and simplifies eGPU setup.

**If you need Windows occasionally:**
- Dual boot (recommended), or
- Run Windows in KVM/QEMU VM for specific tasks

### Initial Setup
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install essentials
sudo apt install -y build-essential git curl wget htop nvtop btop

# Install Python env management
curl -LsSf https://astral.sh/uv/install.sh | sh
```

---

## Phase 2: CPU/RAM Optimized Inference (No eGPU)

### llama.cpp (Primary - CPU inference)
```bash
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp
make -j16 LLAMA_NATIVE=1
```

### Ollama (Easy model management)
```bash
curl -fsSL https://ollama.com/install.sh | sh
ollama pull llama3.2:7b
ollama pull codellama:13b
ollama pull nomic-embed-text
```

### Recommended Models for 64GB RAM (CPU)
| Model | Size | Use Case | Speed |
|-------|------|----------|-------|
| Llama 3.2 7B Q4 | ~4GB | General/code | ~20 tok/s |
| CodeLlama 13B Q4 | ~8GB | Code completion | ~12 tok/s |
| Llama 3.1 70B Q2 | ~26GB | Complex reasoning | ~3 tok/s |
| Mixtral 8x7B Q4 | ~26GB | High quality | ~8 tok/s |

---

## Phase 3: AMD iGPU Acceleration (ROCm)

### Install ROCm for Radeon 780M
```bash
# Add ROCm repo
wget https://repo.radeon.com/rocm/rocm.gpg.key -O - | sudo apt-key add -
echo 'deb [arch=amd64] https://repo.radeon.com/rocm/apt/6.0 ubuntu main' | sudo tee /etc/apt/sources.list.d/rocm.list

sudo apt update
sudo apt install -y rocm-hip-runtime rocm-dev

# Verify
rocminfo
```

### llama.cpp with ROCm
```bash
make clean
make -j16 LLAMA_HIPBLAS=1
```

**Note:** 780M has limited VRAM (shared with system RAM), expect modest gains (~1.5-2x vs pure CPU).

---

## Phase 4: NPU Setup (Ryzen AI)

### Install Ryzen AI SDK
```bash
# Download from AMD site (requires registration)
# https://www.amd.com/en/products/software/ryzen-ai-software.html

# Install ONNX Runtime with DirectML/XDNA
pip install onnxruntime-directml
```

### Use Cases for NPU
- Small model inference (up to ~3B params)
- Embeddings generation
- Whisper transcription
- Background inference while GPU handles main tasks

---

## Phase 5: eGPU Setup

### GPU Selection Guide

**RTX 50 Series Comparison:**

| GPU | VRAM | MSRP | Max Model in VRAM |
|-----|------|------|-------------------|
| RTX 5070 | 12GB | ~$549 | 7B Q8, 13B Q4 |
| RTX 5070 Ti | 16GB | ~$749 | 13B Q8, 30B Q4 |
| RTX 5080 | 16GB | ~$999 | 13B Q8, 30B Q4 |
| RTX 5090 | 32GB | ~$1,999 | 70B Q4 fully in VRAM |

**The VRAM Problem:** 12GB is limiting for AI inference. When model exceeds VRAM, it offloads to system RAM and slows significantly.

| Model | VRAM Needed (Q4) | RTX 5070 (12GB) |
|-------|------------------|-----------------|
| 7B | ~4-5GB | ✓ Full speed |
| 13B | ~8-9GB | ✓ Full speed |
| 30B | ~18GB | ✗ Offload to RAM (slower) |
| 70B | ~35GB | ✗ Mostly on CPU (slow) |

**Used GPU Alternative (Better Value):**

| Option | VRAM | Used Price | Notes |
|--------|------|------------|-------|
| RTX 3090 | 24GB | ~$700-900 | Best value/VRAM ratio |
| RTX 4090 | 24GB | ~$1,400-1,600 | Faster, more efficient |

**Recommendations:**
1. **Budget pick:** Used RTX 3090 (~$800) - 24GB handles 30B fully, 70B with minimal offload
2. **New card:** RTX 5070 Ti (~$749) - if 16GB is enough for your workflow
3. **Skip:** RTX 5070 at 12GB - you'll hit VRAM limits quickly

### Prerequisites
- Thunderbolt 4 eGPU enclosure (Razer Core X, Sonnet Breakaway)
- 650W+ PSU in enclosure
- GPU of choice (see selection guide above)

### NVIDIA Driver Setup
```bash
# Add NVIDIA repo
sudo add-apt-repository ppa:graphics-drivers/ppa
sudo apt update

# Install latest driver (550+ for RTX 50 series)
sudo apt install -y nvidia-driver-550

# Install CUDA toolkit
sudo apt install -y nvidia-cuda-toolkit

# Verify
nvidia-smi
```

### Authorize Thunderbolt Device
```bash
# Install bolt for TB management
sudo apt install -y bolt

# Authorize eGPU
boltctl list
boltctl authorize <device-uuid>
```

### llama.cpp with CUDA
```bash
make clean
make -j16 LLAMA_CUDA=1
```

### Expected Performance by GPU

**RTX 3090 / RTX 5070 Ti+ (24GB / 16GB VRAM):**
| Model | VRAM Used | Speed |
|-------|-----------|-------|
| Llama 3.2 7B Q4 | ~5GB | ~100-120 tok/s |
| CodeLlama 13B Q4 | ~8GB | ~70-90 tok/s |
| Llama 3.1 30B Q4 | ~18GB | ~40 tok/s (3090 only) |
| Llama 3.1 70B Q4 | GPU+CPU offload | ~20-25 tok/s |

**RTX 5070 (12GB VRAM):**
| Model | VRAM Used | Speed |
|-------|-----------|-------|
| Llama 3.2 7B Q4 | ~5GB | ~100 tok/s |
| CodeLlama 13B Q4 | ~8GB | ~70 tok/s |
| Llama 3.1 70B Q4 | Heavy offload | ~10-15 tok/s |

---

## Phase 6: Development Environment

### VS Code / Cursor Setup
```bash
# Install Cursor (AI-native editor)
curl -fsSL https://cursor.sh/install.sh | sh

# Or VS Code with Continue extension
sudo snap install code --classic
```

### Local AI API Server
```bash
# Option 1: Ollama (already installed)
# Runs on http://localhost:11434

# Option 2: LocalAI (OpenAI-compatible)
docker run -p 8080:8080 --gpus all localai/localai:latest

# Option 3: vLLM (production-grade)
pip install vllm
python -m vllm.entrypoints.openai.api_server --model meta-llama/Llama-3.2-7B
```

### IDE Integration
- Configure Cursor/Continue to use `http://localhost:11434`
- Set up code completion with CodeLlama
- Configure embedding model for RAG

---

## Phase 7: Monitoring & Optimization

### System Monitoring
```bash
# CPU/RAM
htop
btop

# GPU (NVIDIA)
nvtop
nvidia-smi -l 1

# GPU (AMD)
radeontop
```

### Performance Tuning
```bash
# Enable hugepages for large models
echo 'vm.nr_hugepages=1024' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Set CPU governor to performance
sudo cpupower frequency-set -g performance
```

---

## Decision Tree: Do You Need the eGPU?

```
Start with CPU-only (Ollama + 7B-13B models)
          │
          ▼
    Is inference speed acceptable?
          │
    ┌─────┴─────┐
   YES          NO
    │            │
    ▼            ▼
  Done      Do you need >13B models?
                 │
           ┌─────┴─────┐
          YES          NO
           │            │
           ▼            ▼
    Consider RTX 5090   RTX 5070 is fine
    (24GB VRAM) or      Add eGPU enclosure
    wait for 5080
```

---

## RAM Upgrade Options

**Form Factor:** DDR5-5600 SO-DIMM (NOT UDIMM - desktop RAM won't fit)

**Source:** [Crucial Compatible Upgrades](https://www.crucial.com/compatible-upgrade-for/gmktec/nucbox-k8-plus)

### Single Modules
| Part Number | Capacity | Price |
|-------------|----------|-------|
| CT8G56C46S5 | 8GB | $60.99 |
| CT16G56C46S5 | 16GB | $120.99 |
| CT24G56C46S5 | 24GB | $156.99 |
| CT32G56C46S5 | 32GB | $234.99 |
| CT48G56C46S5 | 48GB | $304.99 |

### Kits (Matched Pairs - Recommended)
| Part Number | Capacity | Price |
|-------------|----------|-------|
| CT2K8G56C46S5 | 16GB (8GBx2) | $126.99 |
| CT2K16G56C46S5 | 32GB (16GBx2) | $250.99 |
| CT2K24G56C46S5 | 48GB (24GBx2) | $325.99 |
| CT2K32G56C46S5 | 64GB (32GBx2) | $488.99 |
| CT2K48G56C46S5 | 96GB (48GBx2) | $634.99 |

**Specs:** DDR5-5600 • CL46 • Non-ECC • SO-DIMM • 262-pin • 1.1V

### Upgrade Recommendations
| Current | Upgrade To | Cost | Benefit |
|---------|------------|------|---------|
| 64GB | 96GB | ~$635 | Run 70B Q5/Q6 models |
| 64GB | 128GB* | ~$800-1000 | Run 70B Q8, multiple models |

*128GB requires 2x 64GB modules (user-tested compatible, not officially supported)

---

## Estimated Costs

| Component | Cost |
|-----------|------|
| NucBox K8 Plus 64GB | ~$700 |
| RAM Upgrade to 96GB | ~$635 |
| eGPU Enclosure | ~$250-400 |
| RTX 5070 | ~$550 (expected) |
| **Total (with eGPU)** | **~$1,500-1,650** |
| **Total (CPU-only)** | **~$700** |
| **Total (96GB + eGPU)** | **~$2,135-2,285** |

---

## Next Steps

1. [ ] Receive NucBox, install Ubuntu 24.04
2. [ ] Set up base environment (Phase 1-2)
3. [ ] Test inference speeds with target models
4. [ ] Evaluate if eGPU is needed based on real usage
5. [ ] If yes, order enclosure + RTX 5070 when available
