import { ProviderSubmission } from '../../../api-server/src/types';

const API_BASE_URL = 'http://localhost:3000';

/**
 * Submits a new provider registration request
 * @param data the provider submission data
 * @returns the new submission's ID
 */
export async function submitProviderSubmission(data: Omit<ProviderSubmission, 'submissionId' | 'status' | 'createdAt'>): Promise<{ submissionId: string }> {
    const response = await fetch(`${API_BASE_URL}/provider-submissions`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    });

    if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to submit provider registration.');
    }

    const result = await response.json();
    return result.data;
}

/**
 * Fetches the list of available data fields from all approved providers
 */
export async function fetchAvailableFields() {
    const response = await fetch(`${API_BASE_URL}/available-fields`);
    if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to fetch available fields.');
    }
    const result = await response.json();
    return result.data;
}

/**
 * Submits a new consumer application
 * @param data the application data
 */
export async function submitApplication(data: { appId: string, requiredFields: object }) {
    const response = await fetch(`${API_BASE_URL}/applications`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    });

    if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to submit application.');
    }

    return response.json();
}

/**
 * Submits a new provider schema
 * @param providerId
 * @param payload the fieldConfigurations object 
 */
export async function submitProviderSchema(providerId: string, payload: any) {
    const response = await fetch(`${API_BASE_URL}/providers/${providerId}/schemas`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Failed to submit schema');
    }

    return response.json();
}
