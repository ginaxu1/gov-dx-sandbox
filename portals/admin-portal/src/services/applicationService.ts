// services/applicationService.ts
import type { 
  ApplicationRegistration, 
  ApplicationSubmission, 
  ApprovedApplication,
  PendingApplicationApiResponse,
  ApprovedApplicationApiResponse
} from '../types/applications';

export class ApplicationService {
  
  static async registerApplication(consumerId: string, applicationData: ApplicationRegistration): Promise<void> {
    const baseUrl = window.configs.apiUrl || import.meta.env.VITE_BASE_PATH || '';
    try {
      const response = await fetch(`${baseUrl}/application-submissions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...applicationData,
          consumerId
        }),
      });

      if (!response.ok) {
        let errorMessage = `Application registration failed with status: ${response.status}`;
        
        try {
          // Try to get error details from response
          const errorData = await response.json();
          if (errorData.message) {
            errorMessage += ` - ${errorData.message}`;
          } else if (errorData.error) {
            errorMessage += ` - ${errorData.error}`;
          } else if (typeof errorData === 'string') {
            errorMessage += ` - ${errorData}`;
          }
        } catch (jsonError) {
          // If we can't parse the error response, use the status text
          errorMessage += ` - ${response.statusText || 'Unknown error'}`;
        }
        
        throw new Error(errorMessage);
      }
    } catch (error) {
      // Re-throw network errors or already formatted errors
      if (error instanceof TypeError && error.message.includes('fetch')) {
        throw new Error('Network error: Unable to connect to the server. Please check your connection and try again.');
      }
      throw error;
    }
  }

  static async addReviewToApplicationSubmission(submissionId: string, review: string, status: string): Promise<ApplicationSubmission> {
    const baseUrl = window.configs.apiUrl || import.meta.env.VITE_BASE_PATH || '';
    try {
      const response = await fetch(`${baseUrl}/application-submissions/${submissionId}`, {
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
        throw new Error(`Failed to add review to application submission! status: ${response.status}`);
      }

      const result: ApplicationSubmission = await response.json();
      return result;
    } catch (error) {
      throw new Error(`Failed to add review to application submission: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  static async getApprovedApplications(): Promise<ApprovedApplication[]> {
    const baseUrl = window.configs.apiUrl || import.meta.env.VITE_BASE_PATH || '';
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
    const baseUrl = window.configs.apiUrl || import.meta.env.VITE_BASE_PATH || '';
    try {
      const url: URL = new URL(`${baseUrl}/application-submissions`);
      url.searchParams.append('status', 'pending');
      url.searchParams.append('status', 'rejected');
      const response = await fetch(url.toString(), {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch application submissions! status: ${response.status}`);
      }

      const result: PendingApplicationApiResponse = await response.json();
      
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

  static async updateApplicationSubmission(submissionId: string, registration: ApplicationRegistration): Promise<void> {
    const baseUrl = window.configs.apiUrl || import.meta.env.VITE_BASE_PATH || '';
    console.log('Updating application at:', `${baseUrl}/application-submissions/${submissionId}`);
    try {
      const response = await fetch(`${baseUrl}/application-submissions/${submissionId}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(registration),
      });

      if (!response.ok) {
        let errorMessage = `Application update failed with status: ${response.status}`;
        
        try {
          // Try to get error details from response
          const errorData = await response.json();
          if (errorData.message) {
            errorMessage += ` - ${errorData.message}`;
          } else if (errorData.error) {
            errorMessage += ` - ${errorData.error}`;
          } else if (typeof errorData === 'string') {
            errorMessage += ` - ${errorData}`;
          }
        } catch (jsonError) {
          // If we can't parse the error response, use the status text
          errorMessage += ` - ${response.statusText || 'Unknown error'}`;
        }
        
        throw new Error(errorMessage);
      }
    } catch (error) {
      // Re-throw network errors or already formatted errors
      if (error instanceof TypeError && error.message.includes('fetch')) {
        throw new Error('Network error: Unable to connect to the server. Please check your connection and try again.');
      }
      throw error;
    }
  }

  static async updateApplication(applicationId: string, registration: ApplicationRegistration): Promise<void> {
    const baseUrl = window.configs.apiUrl || import.meta.env.VITE_BASE_PATH || '';
    console.log('Updating application at:', `${baseUrl}/applications/${applicationId}`);
    try {
      const response = await fetch(`${baseUrl}/applications/${applicationId}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(registration),
      });

      if (!response.ok) {
        let errorMessage = `Application update failed with status: ${response.status}`;

        try {
          // Try to get error details from response
          const errorData = await response.json();
          if (errorData.message) {
            errorMessage += ` - ${errorData.message}`;
          } else if (errorData.error) {
            errorMessage += ` - ${errorData.error}`;
          } else if (typeof errorData === 'string') {
            errorMessage += ` - ${errorData}`;
          }
        } catch (jsonError) {
          // If we can't parse the error response, use the status text
          errorMessage += ` - ${response.statusText || 'Unknown error'}`;
        }

        throw new Error(errorMessage);
      }
    } catch (error) {
      // Re-throw network errors or already formatted errors
      if (error instanceof TypeError && error.message.includes('fetch')) {
        throw new Error('Network error: Unable to connect to the server. Please check your connection and try again.');
      }
      throw error;
    }
  }
}