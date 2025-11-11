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

  // Debug logging
  console.log('Supabase config check:', {
    hasUrl: !!supabaseUrl,
    urlLength: supabaseUrl.length,
    hasKey: !!supabaseAnonKey,
    keyLength: supabaseAnonKey.length,
    nodeEnv: process.env.NODE_ENV
  })

  if (!supabaseUrl || !supabaseAnonKey) {
    console.warn('Supabase environment variables not configured', {
      url: supabaseUrl ? 'SET' : 'MISSING',
      key: supabaseAnonKey ? 'SET' : 'MISSING'
    })
    return null
  }

  // Clean up URL if it has duplicate .supabase.co
  const cleanedUrl = supabaseUrl.replace(/\.supabase\.co\.supabase\.co$/, '.supabase.co')

  supabaseInstance = createClient<Database>(cleanedUrl, supabaseAnonKey)
  return supabaseInstance
}

// Only create Supabase client in production
export const supabase: SupabaseClient<Database> | null = getSupabaseClient()

// Re-export types for convenience
export type AppliedJob = Database['public']['Tables']['Job']['Row']
export type AppliedJobInsert = Database['public']['Tables']['Job']['Insert']
export type AppliedJobUpdate = Database['public']['Tables']['Job']['Update']
