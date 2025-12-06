"""
Evaluator prompts - low-token optimized.
"""

# Shared base: score>=60 passes, be lenient
EVAL_BASE = """Evaluate response. Output JSON via tool:
- feedback: brief assessment
- success_criteria_met: true if score>=60 and task addressed
- user_input_needed: true if user must respond
- score: 0-100

Be LENIENT: pass if main question answered. Focus on what's provided, not missing."""


def build_career_evaluator_prompt(_ctx: str = "") -> str:
    """Career evaluator - job search and cover letters."""
    return f"""{EVAL_BASE}

CAREER CHECKS:
- Job search: external postings are VALID (not user's resume). Direct URLs only.
- Cover letters: tailored to job, professional tone, proper format."""


def build_general_evaluator_prompt(_ctx: str = "") -> str:
    """General evaluator."""
    return f"""{EVAL_BASE}

CHECK: accurate, complete, clear, relevant."""


def build_crm_evaluator_prompt(_ctx: str = "") -> str:
    """CRM evaluator - data queries and proposals."""
    return f"""{EVAL_BASE}

CRM CHECKS: data accuracy, addresses query, relevant entities, proposals well-formatted.
Do NOT evaluate against resume criteria."""
