export interface ContactInput {
  id?: number;
  individual_id?: number;
  first_name: string;
  last_name: string;
  email?: string;
  phone?: string;
  title?: string;
  is_primary?: boolean;
}

export interface CreatedJob {
  id: string;
  account: string;
  account_type: 'Organization' | 'Individual' | null;
  job_title: string;
  date: string;
  status: 'Lead' | 'Applied' | 'Interviewing' | 'Rejected' | 'Offer' | 'Accepted' | 'Do Not Apply';
  job_url: string | null;
  description: string | null;
  resume: string | null;
  salary_range: string | null;  // Display label from SalaryRange
  salary_range_id: number | null;  // FK to SalaryRange
  industry: string[] | null;
  source: 'manual' | 'agent';
  lead_status_id: string | null;
  contacts?: ContactInput[];
  created_at: string;
  updated_at: string;
}

// ==============================================
// Prospect Research Data Model Types
// ==============================================

// Lookup Tables
export interface RevenueRange {
  id: number;
  label: string;
  min_value: number | null;
  max_value: number | null;
  display_order: number;
}

export interface Technology {
  id: number;
  name: string;
  category: string;
  vendor: string | null;
}

// Organization Extensions
export interface OrganizationFirmographics {
  revenue_range_id: number | null;
  revenue_range?: RevenueRange;
  founding_year: number | null;
  headquarters_city: string | null;
  headquarters_state: string | null;
  headquarters_country: string | null;
  company_type: 'private' | 'public' | 'nonprofit' | 'government' | null;
  linkedin_url: string | null;
  crunchbase_url: string | null;
}

export interface OrganizationTechnology {
  organization_id: number;
  technology_id: number;
  technology?: Technology;
  detected_at: string;
  source: string | null;
  confidence: number | null;
}

export interface FundingRound {
  id: number;
  organization_id: number;
  round_type: string;
  amount: number | null;
  announced_date: string | null;
  lead_investor: string | null;
  source_url: string | null;
  created_at: string;
}

export interface CompanySignal {
  id: number;
  organization_id: number;
  signal_type: 'hiring' | 'expansion' | 'product_launch' | 'partnership' | 'leadership_change' | 'news';
  headline: string;
  description: string | null;
  signal_date: string | null;
  source: string | null;
  source_url: string | null;
  sentiment: 'positive' | 'neutral' | 'negative' | null;
  relevance_score: number | null;
  created_at: string;
}

// Individual Extensions
export interface IndividualContactIntelligence {
  twitter_url: string | null;
  github_url: string | null;
  bio: string | null;
  seniority_level: 'C-Level' | 'VP' | 'Director' | 'Manager' | 'IC' | null;
  department: string | null;
  is_decision_maker: boolean;
  reports_to_id: number | null;
  reports_to?: Individual;
}

// Data Provenance
export interface DataProvenance {
  id: number;
  entity_type: 'Organization' | 'Individual';
  entity_id: number;
  field_name: string;
  source: string;
  source_url: string | null;
  confidence: number | null;
  verified_at: string;
  verified_by: number | null;
  raw_value: Record<string, unknown> | null;
  created_at: string;
}

// Extended Organization Type
export interface Organization {
  id: number;
  account_id: number | null;
  name: string;
  website: string | null;
  phone: string | null;
  employee_count: number | null;
  employee_count_range_id: number | null;
  funding_stage_id: number | null;
  description: string | null;
  notes: string | null;
  created_at: string;
  updated_at: string;
}

export interface OrganizationWithResearch extends Organization, OrganizationFirmographics {
  technologies?: OrganizationTechnology[];
  funding_rounds?: FundingRound[];
  signals?: CompanySignal[];
}

// Extended Individual Type
export interface Individual {
  id: number;
  account_id: number | null;
  first_name: string;
  last_name: string;
  email: string | null;
  phone: string | null;
  linkedin_url: string | null;
  title: string | null;
  description: string | null;
  created_at: string;
  updated_at: string;
}

export interface IndividualWithResearch extends Individual, IndividualContactIntelligence {
  direct_reports?: Individual[];
}
