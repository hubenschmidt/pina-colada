import { createClient, SupabaseClient } from '@supabase/supabase-js'
import { Database } from './database.types'

let supabaseInstance: SupabaseClient<Database> | null = null

const getSupabaseClient = (): SupabaseClient<Database> | null => {
  // Don't create Supabase client in development (use local Postgres via API routes)
  if (process.env.NODE_ENV === "development") {
    return null
  }

  if (supabaseInstance) {
    return supabaseInstance
  }

  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL || ''
  const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY || ''

  if (!supabaseUrl || !supabaseAnonKey) {
    console.warn('Supabase environment variables not configured')
    return null
  }

  supabaseInstance = createClient<Database>(supabaseUrl, supabaseAnonKey)
  return supabaseInstance
}

// Only create Supabase client in production
export const supabase: SupabaseClient<Database> | null = getSupabaseClient()

// Re-export types for convenience
export type AppliedJob = Database['public']['Tables']['applied_jobs']['Row']
export type AppliedJobInsert = Database['public']['Tables']['applied_jobs']['Insert']
export type AppliedJobUpdate = Database['public']['Tables']['applied_jobs']['Update']
