# GPU Docker LLM Setup

## Summary

Set up local LLM inference using RX 7900 XT eGPU with Docker and ROCm.

## Hardware

| Component | Model | Status |
|-----------|-------|--------|
| Host | GMKTec NucBox K8 (Ryzen 7 8845HS) | Existing |
| GPU | XFX AMD Radeon RX 7900 XT (20GB) | Purchased |
| eGPU Dock | MINISFORUM DEG1 (OcuLink) | Purchased |
| PSU | MSI MAG A750GL 750W | Purchased |
| Cables | 2x DisplayPort, OcuLink (included) | Purchased |

## Software Requirements

- [ ] Linux kernel 6.2+ (RDNA 3 support)
- [ ] ROCm 6.0+ drivers
- [ ] AMD Container Toolkit
- [ ] Docker with GPU passthrough configured

## Installation Steps

### 1. Install ROCm drivers

```bash
# Add ROCm repo (Ubuntu)
wget https://repo.radeon.com/amdgpu-install/latest/ubuntu/jammy/amdgpu-install_6.0.60000-1_all.deb
sudo apt install ./amdgpu-install_6.0.60000-1_all.deb
sudo amdgpu-install --usecase=rocm

# Add user to groups
sudo usermod -aG render,video $USER
```

### 2. Verify GPU detection

```bash
rocminfo | grep -E "Name:|Marketing"
# Should show: gfx1100 (RDNA 3)
```

### 3. Install AMD Container Toolkit

```bash
# Install container toolkit
sudo apt install amd-container-toolkit

# Configure Docker
sudo amd-container-toolkit setup --runtime=docker
sudo systemctl restart docker
```

### 4. Test GPU in Docker

```bash
docker run --rm --device /dev/kfd --device /dev/dri rocm/rocm-terminal rocminfo
```

## Usage Examples

### Ollama (easiest)

```bash
docker run -d \
  --name ollama \
  --device /dev/kfd \
  --device /dev/dri \
  -v ollama:/root/.ollama \
  -p 11434:11434 \
  ollama/ollama:rocm

# Pull and run a model
docker exec ollama ollama pull qwen2.5-coder:32b-instruct-q4_K_M
docker exec ollama ollama run qwen2.5-coder:32b-instruct-q4_K_M
```

### vLLM (OpenAI-compatible API)

```bash
docker run -d \
  --name vllm \
  --device /dev/kfd \
  --device /dev/dri \
  -p 8000:8000 \
  -v ~/.cache/huggingface:/root/.cache/huggingface \
  vllm/vllm-openai:rocm \
  --model Qwen/Qwen2.5-Coder-32B-Instruct-AWQ \
  --max-model-len 8192
```

### llama.cpp (manual)

```bash
docker run -it \
  --device /dev/kfd \
  --device /dev/dri \
  -v ~/models:/models \
  rocm/pytorch:latest \
  bash

# Inside container: build llama.cpp with ROCm
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp
make LLAMA_HIPBLAS=1
./main -m /models/qwen2.5-coder-32b-q4_K_M.gguf -p "Write a function"
```

## Expected Performance

| Model | Quantization | VRAM | Est. tokens/s |
|-------|--------------|------|---------------|
| Qwen 2.5 Coder 32B | Q4_K_M | ~17GB | ~23 |
| DeepSeek Coder 33B | Q4_K_M | ~17GB | ~23 |
| Llama 3.1 8B | Q8 | ~8GB | ~50+ |

## Troubleshooting

### GPU not detected in container
- Verify host can see GPU: `rocminfo`
- Check permissions: user in `render` and `video` groups
- Ensure `/dev/kfd` and `/dev/dri` exist

### Out of memory
- Reduce model size or quantization
- Lower `--max-model-len` in vLLM
- Check `rocm-smi` for VRAM usage

### Slow performance
- Verify GPU is being used (not CPU fallback)
- Check `HSA_OVERRIDE_GFX_VERSION` if needed for RDNA 3

## References

- [ROCm Documentation](https://rocm.docs.amd.com)
- [AMD Container Toolkit](https://rocm.blogs.amd.com/software-tools-optimization/amd-container-toolkit/README.html)
- [Ollama ROCm](https://ollama.com/blog/amd-preview)
- [vLLM ROCm](https://docs.vllm.ai/en/latest/getting_started/amd-installation.html)
