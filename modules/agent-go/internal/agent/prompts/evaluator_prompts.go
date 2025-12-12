package prompts

// Base evaluation rules - shared across all evaluators
const EvalBase = `Evaluate the assistant's response. Return structured JSON:
- feedback: brief assessment (1-2 sentences)
- success_criteria_met: true if score>=60 and task addressed
- user_input_needed: true ONLY if user must provide NEW information (e.g., missing criteria, ambiguous request). FALSE if agent can improve by trying harder or rephrasing.
- score: 0-100

IMPORTANT: user_input_needed should almost always be FALSE. Only set TRUE when the request itself is unclear or requires info the user hasn't provided. If agent just returned too few results or wrong format, set FALSE - agent can retry.

Be LENIENT: pass if main question was answered. Focus on what's provided, not what's missing.`

// CareerEvaluatorPrompt for job search and career-related tasks
const CareerEvaluatorPrompt = EvalBase + `

CAREER-SPECIFIC CHECKS:
- Job search: Must include actual job URLs (not just company homepages)
- Must include CRM record info if requested (name, id, email)
- Must list documents if requested (with IDs)
- Should return the number of results requested (some variation OK if hard to find)
- URLs should be company career pages or ATS links, avoid major job boards (LinkedIn, Indeed)
- Cover letters: tailored to job, professional tone

STRICT FAILURES (score < 50):
- Missing CRM data when requested
- Returns nothing when results were requested
- Includes major job boards: linkedin.com, indeed.com, glassdoor.com, ziprecruiter.com, motionrecruitment.com
- NOTE: ATS platforms (lever.co, greenhouse.io, ashbyhq.com) and startup aggregators (ycombinator.com/jobs, wellfound.com, builtinnyc.com) are OK

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
