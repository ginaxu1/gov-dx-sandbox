// services/applicationService.ts
import type { 
  ApplicationRegistration, 
  ApplicationSubmission, 
  ApprovedApplication
} from '../types/applications';

export class ApplicationService {
  
  static async getApprovedApplications(consumerId: string): Promise<ApprovedApplication[]> {
    const baseUrl = window.configs.apiUrl || import.meta.env.VITE_BASE_PATH || '';
    try {
      const response = await fetch(`${baseUrl}/consumers/${consumerId}/applications`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch approved applications! status: ${response.status}`);
      }

      const result = await response.json();
      
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

  static async getApplicationSubmissions(consumerId: string): Promise<ApplicationSubmission[]> {
    const baseUrl = window.configs.apiUrl || import.meta.env.VITE_BASE_PATH || '';
    try {
      const response = await fetch(`${baseUrl}/consumers/${consumerId}/application-submissions`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch application submissions! status: ${response.status}`);
      }

      const result = await response.json();
      
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

  static async registerApplication(consumerId: string, registration: ApplicationRegistration): Promise<void> {
    const baseUrl = window.configs.apiUrl || import.meta.env.VITE_BASE_PATH || '';
    console.log('Registering application at:', `${baseUrl}/consumers/${consumerId}/application-submissions`);
    try {
      const response = await fetch(`${baseUrl}/consumers/${consumerId}/application-submissions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(registration),
      });

      if (!response.ok) {
        throw new Error(`Application registration failed! status: ${response.status}`);
      }
    } catch (error) {
      throw new Error(`Failed to register application: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  static async updateApplication(consumerId: string, applicationId: string, updates: Partial<ApplicationRegistration>): Promise<void> {
    const baseUrl = window.configs.apiUrl || import.meta.env.VITE_BASE_PATH || '';
    try {
      const response = await fetch(`${baseUrl}/consumers/${consumerId}/applications/${applicationId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updates),
      });

      if (!response.ok) {
        throw new Error(`Application update failed! status: ${response.status}`);
      }
    } catch (error) {
      throw new Error(`Failed to update application: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  static async deleteApplication(consumerId: string, applicationId: string): Promise<void> {
    const baseUrl = window.configs.apiUrl || import.meta.env.VITE_BASE_PATH || '';
    try {
      const response = await fetch(`${baseUrl}/consumers/${consumerId}/applications/${applicationId}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Application deletion failed! status: ${response.status}`);
      }
    } catch (error) {
      throw new Error(`Failed to delete application: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  // Helper method to get application display name
  static getApplicationDisplayName(application: ApprovedApplication | ApplicationSubmission): string {
    return application.name || 'Unnamed Application';
  }

  // Helper method to get application status for display
  static getApplicationStatusInfo(application: ApprovedApplication | ApplicationSubmission): {
    status: string;
    colorClass: string;
    icon: 'active' | 'deprecated' | 'pending' | 'approved' | 'rejected';
  } {
    if ('version' in application) {
      // ApprovedApplication
      return {
        status: application.version,
        colorClass: application.version === 'Active' ? 'text-green-600' : 'text-orange-600',
        icon: application.version === 'Active' ? 'active' : 'deprecated'
      };
    } else {
      // ApplicationSubmission
      const statusMap = {
        'pending': { colorClass: 'text-yellow-600', icon: 'pending' as const },
        'approved': { colorClass: 'text-green-600', icon: 'approved' as const },
        'rejected': { colorClass: 'text-red-600', icon: 'rejected' as const }
      };
      
      const statusInfo = statusMap[application.status] || statusMap.pending;
      return {
        status: application.status.charAt(0).toUpperCase() + application.status.slice(1),
        colorClass: statusInfo.colorClass,
        icon: statusInfo.icon
      };
    }
  }
}