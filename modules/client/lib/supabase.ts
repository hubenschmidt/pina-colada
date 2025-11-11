import { createClient } from '@supabase/supabase-js'

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL || ''
const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY || ''

if (!supabaseUrl || !supabaseAnonKey) {
  console.warn('Supabase environment variables not configured')
}

export const supabase = createClient(supabaseUrl, supabaseAnonKey)

export type AppliedJob = {
  id: string
  company: string
  job_title: string
  application_date: string
  status: 'applied' | 'interviewing' | 'rejected' | 'offer' | 'accepted'
  job_url?: string
  location?: string
  salary_range?: string
  notes?: string
  source: 'manual' | 'agent'
  created_at: string
  updated_at: string
}
