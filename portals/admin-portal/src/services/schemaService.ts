import type { SchemaSubmission, ApprovedSchema, ApprovedSchemaApiResponse, SchemaSubmissionApiResponse } from '../types/schema';
import { URLBuilder } from '../utils';

export class SchemaService {
  static async addReviewToSchemaSubmission(submissionId: string, review: string, status: "approved" | "rejected"): Promise<SchemaSubmission> {
    const baseUrl = window.configs.VITE_API_URL || import.meta.env.VITE_BASE_PATH || '';
    try {
      const response = await fetch(`${baseUrl}/schema-submissions/${submissionId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          review,
          status
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to add review to schema submission! status: ${response.status}`);
      }

      const result: SchemaSubmission = await response.json();
      return result;
    } catch (error) {
      throw new Error(`Failed to add review to schema submission: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  static async getApprovedSchemas(): Promise<ApprovedSchema[]> {
    const baseUrl = window.configs.VITE_API_URL || import.meta.env.VITE_BASE_PATH || '';
    try {
      const response = await fetch(`${baseUrl}/schemas`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch schemas! status: ${response.status}`);
      }

      const result: ApprovedSchemaApiResponse = await response.json();
      
      // Handle API response structure {count: number, items: Array | null}
      if (result && typeof result === 'object' && 'items' in result) {
        return Array.isArray(result.items) ? result.items : [];
      }
      
      // Fallback for direct array response
      console.log('Result is not in expected format, returning empty array.');
      return Array.isArray(result) ? result : [];
    } catch (error) {
      throw new Error(`Failed to get approved schemas: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  static async getSchemaSubmissions(): Promise<SchemaSubmission[]> {
    const baseUrl = window.configs.VITE_API_URL || import.meta.env.VITE_BASE_PATH || '';
    try {
      const url = URLBuilder.from(baseUrl)
        .path('/schema-submissions')
        .param('status', 'pending')
        .param('status', 'rejected')
        .build();
      const response = await fetch(url, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
       
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch schema submissions! status: ${response.status}`);
      }

      const result: SchemaSubmissionApiResponse = await response.json();

      // Handle API response structure {count: number, items: Array | null}
      if (result && typeof result === 'object' && 'items' in result) {
        return Array.isArray(result.items) ? result.items : [];
      }

      // Fallback for direct array response
      console.log('Result is not in expected format, returning empty array.');
      return Array.isArray(result) ? result : [];
    } catch (error) {
      throw new Error(`Failed to get schema submissions: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}