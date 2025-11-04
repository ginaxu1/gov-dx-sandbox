import type {
    ApplicationSubmission,
    ApplicationSubmissionApiResponse,
    ApprovedApplication,
    ApprovedApplicationApiResponse
} from '../types/applications';
import { URLBuilder } from '../utils';

export class ApplicationService {
  static async addReviewToApplicationSubmission(submissionId: string, review: string, status: "approved" | "rejected"): Promise<ApplicationSubmission> {
    const baseUrl = window.configs.VITE_API_URL || import.meta.env.VITE_BASE_PATH || '';
    try {
      const response = await fetch(`${baseUrl}/application-submissions/${submissionId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ review, status }),
      });

      if (!response.ok) {
        throw new Error(`Failed to add review to application submission! status: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      throw new Error(`Failed to add review to application submission: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  static async getApprovedApplications(): Promise<ApprovedApplication[]> {
    const baseUrl = window.configs.VITE_API_URL || import.meta.env.VITE_BASE_PATH || '';
    try {
      const response = await fetch(`${baseUrl}/applications`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch approved applications! status: ${response.status}`);
      }

      const result: ApprovedApplicationApiResponse = await response.json();
      
      // Handle API response structure {count: number, items: Array | null}
      if (result && typeof result === 'object' && 'items' in result) {
        return Array.isArray(result.items) ? result.items : [];
      }
      
      // Fallback for direct array response
      return Array.isArray(result) ? result : [];
    } catch (error) {
      throw new Error(`Failed to get approved applications: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  static async getApplicationSubmissions(): Promise<ApplicationSubmission[]> {
    const baseUrl = window.configs.VITE_API_URL || import.meta.env.VITE_BASE_PATH || '';
    try {
      const url = URLBuilder.from(baseUrl)
        .path('/application-submissions')
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
        throw new Error(`Failed to fetch application submissions! status: ${response.status}`);
      }

      const result: ApplicationSubmissionApiResponse = await response.json();
      
      // Handle API response structure {count: number, items: Array | null}
      if (result && typeof result === 'object' && 'items' in result) {
        return Array.isArray(result.items) ? result.items : [];
      }
      
      // Fallback for direct array response
      return Array.isArray(result) ? result : [];
    } catch (error) {
      throw new Error(`Failed to get application submissions: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}