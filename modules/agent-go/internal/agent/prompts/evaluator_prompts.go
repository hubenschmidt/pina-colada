package prompts

// Base evaluation rules - shared across all evaluators
const EvalBase = `Evaluate the assistant's response. Return structured JSON:
- feedback: brief assessment (1-2 sentences)
- success_criteria_met: true if score>=60 and task addressed
- user_input_needed: true if user must respond before continuing
- score: 0-100

Be LENIENT: pass if main question was answered. Focus on what's provided, not what's missing.`

// CareerEvaluatorPrompt for job search and career-related tasks
const CareerEvaluatorPrompt = EvalBase + `

CAREER-SPECIFIC CHECKS:
- Job search: Must include actual job URLs (not just company homepages)
- Must include CRM record info if requested (name, id, email)
- Must list documents if requested (with IDs)
- Must return EXACTLY the number of results requested (e.g., 5 results = 5 URLs)
- URLs must be direct company career pages, NOT job boards (LinkedIn, Indeed, BuiltIn, etc.)
- Cover letters: tailored to job, professional tone

STRICT FAILURES (score < 50):
- Missing CRM data when requested
- Wrong number of results (asked for 5, returned 10)
- Includes job board URLs (linkedin.com, indeed.com, builtinnyc.com, lever.co, greenhouse.io, etc.)

CRITICAL: If user asked to look up CRM record AND search jobs, response MUST include BOTH.`

// CRMEvaluatorPrompt for CRM data queries
const CRMEvaluatorPrompt = EvalBase + `

CRM-SPECIFIC CHECKS:
- Data accuracy: IDs and names should match what was requested
- Addresses the query: includes requested fields
- Lists relevant entities
- Proposals well-formatted

Do NOT evaluate against job search or resume criteria.`

// GeneralEvaluatorPrompt for general queries
const GeneralEvaluatorPrompt = EvalBase + `

GENERAL CHECKS:
- Accurate information
- Complete answer
- Clear and relevant response`
