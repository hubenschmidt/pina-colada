# Local LLM Hardware Comparison

## Hardware Specs

| Spec | M2 Max (32GB) | RTX 3090 | RTX 5090 |
|------|---------------|----------|----------|
| Memory | 32GB unified | 24GB GDDR6X | 32GB GDDR7 |
| Bandwidth | ~400 GB/s | ~936 GB/s | ~1.5TB/s |
| Power | ~30W | ~350W | ~450W |

## Model Compatibility

| Model | M2 Max 32GB | RTX 3090 | RTX 5090 |
|-------|-------------|----------|----------|
| 7-8B Q8 | ✅ | ✅ | ✅ |
| 13-14B Q8 | ✅ tight | ✅ | ✅ |
| 32-34B Q4 | ✅ | ✅ | ✅ |
| 32-34B Q8 | ❌ | ❌ | ⚠️ tight |
| 70B Q3 | ❌ | ❌ | ✅ |
| 70B Q4 | ❌ | ❌ | ⚠️ barely |
| Mixtral 8x7B Q4 | ❌ | ⚠️ tight | ✅ |

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

## Notes

- Larger model at Q4 typically beats smaller model at Q8
- Quality loss appears first in: complex reasoning, math, long-context coherence
- OpenAI models (GPT-4, etc.) are proprietary — not runnable locally
