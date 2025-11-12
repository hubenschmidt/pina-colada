import { createJob, fetchLeadStatuses } from '../api/jobs'

/**
 * Pattern to match job leads in agent responses:
 * Company - Job Title - https://url
 * Also handles numbered lists: 1. Company - Job Title - https://url
 */
const JOB_LEAD_PATTERN = /(?:^\d+\.\s*)?(.+?)\s*-\s*(.+?)\s*-\s*(https?:\/\/[^\s]+)/gm

/**
 * Extract job leads from message text
 */
export const extractJobLeads = (messageText: string): Array<{
  company: string
  job_title: string
  job_url: string
}> => {
  const leads: Array<{ company: string; job_title: string; job_url: string }> = []
  const matches = Array.from(messageText.matchAll(JOB_LEAD_PATTERN))

  for (const match of matches) {
    leads.push({
      company: match[1].trim(),
      job_title: match[2].trim(),
      job_url: match[3].trim(),
    })
  }

  return leads
}

/**
 * Save extracted job leads to the database
 * @returns Number of leads saved
 */
export const saveJobLeads = async (
  leads: Array<{ company: string; job_title: string; job_url: string }>
): Promise<number> => {
  if (leads.length === 0) return 0

  // Get the "Qualifying" lead status ID
  const leadStatuses = await fetchLeadStatuses()
  const qualifyingStatus = leadStatuses.find(s => s.name === 'Qualifying')

  if (!qualifyingStatus) {
    console.error('Qualifying lead status not found')
    return 0
  }

  let savedCount = 0

  for (const lead of leads) {
    try {
      await createJob({
        company: lead.company,
        job_title: lead.job_title,
        job_url: lead.job_url,
        status: 'lead',
        source: 'agent',
        lead_status_id: qualifyingStatus.id,
      })
      savedCount++
    } catch (error) {
      console.error('Failed to save job lead:', lead, error)
    }
  }

  return savedCount
}

/**
 * Extract and save job leads from a message
 * @returns Number of leads saved
 */
export const extractAndSaveJobLeads = async (messageText: string): Promise<number> => {
  const leads = extractJobLeads(messageText)
  return await saveJobLeads(leads)
}
