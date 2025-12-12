export interface ConsentRecord {
  consent_id: string;
  owner_id: string;
  owner_name: string;
  owner_email: string;
  data_consumer: string;
  status: 'pending' | 'approved' | 'rejected' | 'expired' | 'revoked';
  type?: string;
  created_at: string;
  updated_at: string;
  expires_at: string;
  fields: string[];
  session_id: string;
  redirect_url: string;
  app_display_name: string;
}