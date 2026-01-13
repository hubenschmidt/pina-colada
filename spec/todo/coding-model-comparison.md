# Open Source Coding Models Comparison

## Hardware: RX 7900 XT (20GB VRAM, ~800 GB/s bandwidth)

## 2025 Benchmark Comparison

| Model | Size | HumanEval | LiveCodeBench | Tool Calling | Agency | Notes |
|-------|------|-----------|---------------|--------------|--------|-------|
| MiniMax-M2 | 456B MoE | 85-87% | 70-75% | 80-85% | 80-85% | SOTA open-source, 1M context, too large for 20GB |
| DeepSeek-Coder-V2 Lite | 16B | 90.2% | 43.4% | 65-70% | 70-75% | Excellent coder, fits 20GB ✅ |
| Qwen3-Coder | 14-30B | 89.3% | 70.7% | 75-80% | 75-80% | Native tool calling, 128K context ✅ |
| Gemma2 | 9B | 82-86% | 60-70% | 75-80% | 72-78% | Fastest, good function calling ✅ |
| GLM-Z1-9B | 9B | 75-80% | 65-70% | 55-60% | 60-65% | Good reasoning, prompt-only agency |
| Codestral | 22B | 71-87% | 37.9% | 65-70% | 65-70% | Solid multi-language |
| CodeLlama | 13-34B | 53-68% | 30-40% | <50% | <50% | Outdated (2023), skip |

## Fits on 7900 XT (20GB)

| Model | Quant | VRAM | Est. tok/s | Best For |
|-------|-------|------|------------|----------|
| DeepSeek-Coder-V2 Lite 16B | Q8 | ~16GB | ~25 | Best HumanEval score |
| Qwen3-Coder 14B | Q8 | ~14GB | ~28 | Tool calling + agency |
| Gemma2 9B | Q8 | ~9GB | ~45 | Speed + function calling |
| GLM-Z1-9B | Q8 | ~9GB | ~45 | Reasoning + planning |

## Recommended for Your Setup

### Best Coder (fits, good speed)
```bash
ollama pull deepseek-coder-v2:16b
```
90.2% HumanEval — highest coding accuracy that fits.

### Best for Agentic/Tool Use
```bash
ollama pull qwen3:14b
```
Native tool calling, 128K context, strong agency.

### Fastest (still good)
```bash
ollama pull gemma2:9b
```
~45 tok/s, good function calling, reliable structured output.

## Ollama Model Tags

| Model | Ollama Tag |
|-------|------------|
| Qwen 2.5 Coder 32B Q4 | `qwen2.5-coder:32b` |
| Qwen 2.5 Coder 14B | `qwen2.5-coder:14b` |
| Qwen 2.5 Coder 7B | `qwen2.5-coder:7b` |
| DeepSeek Coder V2 16B | `deepseek-coder-v2:16b` |
| DeepSeek Coder 33B | `deepseek-coder:33b` |
| CodeLlama 34B | `codellama:34b` |
| StarCoder2 15B | `starcoder2:15b` |

## Task-Based Recommendations

| Task | Recommended Model | Why |
|------|-------------------|-----|
| Autocomplete / inline suggestions | 7B Q8 | Speed matters most |
| Single function generation | 14B Q8 | Good balance |
| Multi-file refactoring | 32B Q4 | Quality matters, wait is OK |
| Code review / explanation | 14B Q8 | Good reasoning, fast enough |
| Debugging complex issues | 32B Q4 | Needs best reasoning |

## Benchmark Notes

- HumanEval scores are approximate and vary by quantization
- Real-world coding ability often differs from benchmarks
- Qwen 2.5 Coder family currently leads on most coding benchmarks
- DeepSeek Coder V2 excels at long-context tasks

## Try This

Test the 14B model — best balance of speed and quality:
```bash
ollama pull qwen2.5-coder:14b
ollama run qwen2.5-coder:14b
```

Expected: ~28 tok/s on 7900 XT (2x faster than 32B, 90% of quality)
