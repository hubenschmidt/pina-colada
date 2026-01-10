# Local LLM Hardware Comparison

## Hardware Specs

| Spec | M2 Max (32GB) | RTX 3090 | RTX 5090 | AI PRO R9700 |
|------|---------------|----------|----------|--------------|
| Memory | 32GB unified | 24GB GDDR6X | 32GB GDDR7 | 32GB GDDR6 |
| Bandwidth | ~400 GB/s | ~936 GB/s | ~1.5TB/s | 640 GB/s |
| Power | ~30W | ~350W | ~450W | 300W |

## Model Compatibility

| Model | M2 Max 32GB | RTX 3090 | RTX 5090 | AI PRO R9700 |
|-------|-------------|----------|----------|--------------|
| 7-8B Q8 | ✅ | ✅ | ✅ | ✅ |
| 13-14B Q8 | ✅ tight | ✅ | ✅ | ✅ |
| 32-34B Q4 | ✅ | ✅ | ✅ | ✅ |
| 32-34B Q8 | ❌ | ❌ | ⚠️ tight | ⚠️ tight |
| 70B Q3 | ❌ | ❌ | ✅ | ✅ |
| 70B Q4 | ❌ | ❌ | ⚠️ barely | ❌ |
| Mixtral 8x7B Q4 | ❌ | ⚠️ tight | ✅ | ✅ |

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

## eGPU Setup: AI PRO R9700 + OcuLink

**Host**: GMKTec NucBox K8 (or similar mini PC with OcuLink port)

| Component | Est. Cost | Notes |
|-----------|-----------|-------|
| AMD Radeon AI PRO R9700 | $1,299 | 32GB GDDR6 |
| OcuLink eGPU enclosure | ~$60-100 | ADT-Link R43SG or similar open-frame |
| OcuLink cable (SFF-8611) | ~$20 | Keep <50cm for signal integrity |
| ATX PSU 750W+ | ~$80-120 | Needs 12V-2x6 connector |
| PSU jumper/bridge | ~$5 | To power on standalone PSU |
| GPU support bracket | ~$15 | Optional, prevents sag |
| **Total** | **~$1,500-1,560** | Excluding host mini PC |

**Bandwidth note**: OcuLink PCIe 4.0 x4 = ~8 GB/s. Adequate for LLM inference (model stays in VRAM), but initial model load will be slower than native PCIe x16.

## Notes

- Larger model at Q4 typically beats smaller model at Q8
- Quality loss appears first in: complex reasoning, math, long-context coherence
- OpenAI models (GPT-4, etc.) are proprietary — not runnable locally
