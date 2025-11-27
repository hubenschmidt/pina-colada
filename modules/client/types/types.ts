export interface ContactInput {
  individual_id?: number;
  first_name: string;
  last_name: string;
  email?: string;
  phone?: string;
  title?: string;
}

export interface CreatedJob {
  id: string;
  account: string;
  account_type: 'Organization' | 'Individual' | null;
  job_title: string;
  date: string;
  status: 'Lead' | 'Applied' | 'Interviewing' | 'Rejected' | 'Offer' | 'Accepted' | 'Do Not Apply';
  job_url: string | null;
  notes: string | null;
  resume: string | null;
  salary_range: string | null;  // Display label from RevenueRange
  revenue_range_id: number | null;  // FK to RevenueRange
  industry: string | null;
  source: 'manual' | 'agent';
  lead_status_id: string | null;
  contacts?: ContactInput[];
  created_at: string;
  updated_at: string;
}
