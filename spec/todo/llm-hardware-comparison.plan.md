# Local LLM Hardware Comparison

## Hardware Specs

| Spec | M2 Max (32GB) | RTX 3090 | RTX 5090 | AI PRO R9700 | RX 7900 XT |
|------|---------------|----------|----------|--------------|------------|
| Memory | 32GB unified | 24GB GDDR6X | 32GB GDDR7 | 32GB GDDR6 | 20GB GDDR6 |
| Bandwidth | ~400 GB/s | ~936 GB/s | ~1.5TB/s | 640 GB/s | ~800 GB/s |
| Power | ~30W | ~350W | ~450W | 300W | ~300W |

## Model Compatibility

| Model | M2 Max 32GB | RTX 3090 | RTX 5090 | AI PRO R9700 | RX 7900 XT |
|-------|-------------|----------|----------|--------------|------------|
| 7-8B Q8 | ✅ | ✅ | ✅ | ✅ | ✅ |
| 13-14B Q8 | ✅ tight | ✅ | ✅ | ✅ | ✅ |
| 32-34B Q4 | ✅ | ✅ | ✅ | ✅ | ⚠️ tight |
| 32-34B Q8 | ❌ | ❌ | ⚠️ tight | ⚠️ tight | ❌ |
| 70B Q3 | ❌ | ❌ | ✅ | ✅ | ❌ |
| 70B Q4 | ❌ | ❌ | ⚠️ barely | ❌ | ❌ |
| Mixtral 8x7B Q4 | ❌ | ⚠️ tight | ✅ | ✅ | ❌ |

## Quantization Reference

| Level | Memory/1B params | Quality |
|-------|------------------|---------|
| Q8 | ~1GB | near-lossless |
| Q6_K | ~0.75GB | minimal loss |
| Q4_K_M | ~0.5GB | slight loss |
| Q3_K_M | ~0.4GB | noticeable loss |
| Q2_K | ~0.3GB | significant loss |

## Recommendations

- **M2 Max 32GB**: Best for silent operation, power efficiency. Speed ~2x slower than 3090.
- **RTX 3090**: Best value for 7B-34B models. Sweet spot: Qwen 2.5 32B Q4.
- **RTX 5090**: Opens 70B territory. Sweet spot: Llama 3.1 70B Q3_K_M.
- **AI PRO R9700**: 32GB matches RTX 5090 capacity at lower bandwidth. Sweet spot: Llama 3.1 70B Q3_K_M or Qwen 2.5 32B Q8. ROCm support improving but still behind CUDA.
- **RX 7900 XT**: Best value AMD option. 20GB caps at 32B Q4. Higher bandwidth than R9700. Mature ROCm support (RDNA 3). Sweet spot: Qwen 2.5 32B Q4 or DeepSeek Coder 33B Q4.

## eGPU Setup: OcuLink

**Host**: GMKTec NucBox K8 (or similar mini PC with OcuLink port)

| Component | AI PRO R9700 | RX 7900 XT | Notes |
|-----------|--------------|------------|-------|
| GPU | $1,299 | ~$650-750 | 32GB vs 20GB |
| OcuLink eGPU enclosure | ~$60-100 | ~$60-100 | ADT-Link R43SG or similar |
| OcuLink cable (SFF-8611) | ~$20 | ~$20 | Keep <50cm |
| ATX PSU 750W+ | ~$80-120 | ~$80-120 | 12V-2x6 vs 2× 8-pin |
| PSU jumper/bridge | ~$5 | ~$5 | To power on standalone PSU |
| GPU support bracket | ~$15 | ~$15 | Optional, prevents sag |
| **Total** | **~$1,500-1,560** | **~$850-1,010** | Excluding host mini PC |

**Bandwidth note**: OcuLink PCIe 4.0 x4 = ~8 GB/s. Adequate for LLM inference (model stays in VRAM), but initial model load will be slower than native PCIe x16.

**Trade-off**: RX 7900 XT saves ~$500-600 but loses 12GB VRAM. Same enclosure/cables work for both — upgrade path available.

## Shopping List: RX 7900 XT eGPU Setup

| Item | Qty | Est. Cost | Notes |
|------|-----|-----------|-------|
| XFX AMD Radeon RX 7900 XT | 1 | ~$650-750 | Triple fan, 20GB GDDR6 |
| OcuLink eGPU enclosure | 1 | ~$60-100 | ADT-Link R43SG or similar, verify GPU length ≥330mm |
| OcuLink cable (SFF-8611 to SFF-8611) | 1 | ~$20 | Keep ≤50cm for signal integrity |
| ATX PSU 750W+ | 1 | ~$80-120 | Must have 2× 8-pin PCIe connectors |
| PSU jumper/bridge | 1 | ~$5 | 24-pin ATX jumper to power on standalone PSU |
| GPU support bracket | 1 | ~$15 | Optional but recommended for triple-fan card |
| DisplayPort cables | 2 | ~$15-30 | DP 1.4 or higher, length as needed |
| **Total** | | **~$865-1,040** | |

**Before leaving, verify:**
- [ ] NucBox K8 has OcuLink port (SFF-8612)
- [ ] PSU has 2× 8-pin PCIe cables (not 12V-2x6)
- [ ] Enclosure supports GPU length ≥330mm (XFX triple-fan is long)
- [ ] DisplayPort cable length suits your desk setup

**Software setup (later):**
- Linux kernel 6.2+ (RDNA 3 support)
- ROCm 6.0+ or AMDGPU drivers
- PyTorch ROCm build: `pip install torch --index-url https://download.pytorch.org/whl/rocm6.0`

## Notes

- Larger model at Q4 typically beats smaller model at Q8
- Quality loss appears first in: complex reasoning, math, long-context coherence
- OpenAI models (GPT-4, etc.) are proprietary — not runnable locally
