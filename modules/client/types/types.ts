export interface CreatedJob {
  id: string;
  company: string;
  job_title: string;
  date: string;
  status: 'lead' | 'applied' | 'interviewing' | 'rejected' | 'offer' | 'accepted' | 'do_not_apply';
  job_url: string | null;
  notes: string | null;
  resume: string | null;
  salary_range: string | null;
  source: 'manual' | 'agent';
  lead_status_id: string | null;
  created_at: string;
  updated_at: string;
}
