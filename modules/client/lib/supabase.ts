import { SupabaseClient } from '@supabase/supabase-js'
import { Database } from './database.types'

const getSupabaseClient = (): SupabaseClient<Database> | null => {
  // Never create Supabase client - always use agent's REST API
  // The agent's repository layer handles Supabase vs local Postgres based on environment
  return null
}

// Supabase client is always null - client should use agent's REST API instead
export const supabase: SupabaseClient<Database> | null = getSupabaseClient()

// Re-export types for convenience
export type AppliedJob = Database['public']['Tables']['Job']['Row']
export type AppliedJobInsert = Database['public']['Tables']['Job']['Insert']
export type AppliedJobUpdate = Database['public']['Tables']['Job']['Update']
