import { AvailableFieldsResponse } from '../components/ConsumerRegistrationForm';

const API_BASE_URL = 'http://localhost:3000';

interface ApplicationPayload {
  appId: string;
  requiredFields: Record<string, object>;
}

export const fetchAvailableFields = async (): Promise<AvailableFieldsResponse> => {
  const response = await fetch(`${API_BASE_URL}/available-fields`);
  if (!response.ok) {
    throw new Error(`Failed to fetch available fields: ${response.statusText}`);
  }
  const result = await response.json();
  if (result.status !== 'success') {
    throw new Error(result.message || 'An error occurred while fetching fields.');
  }
  return result.data;
};

export const submitApplication = async (payload: ApplicationPayload): Promise<void> => {
  const response = await fetch(`${API_BASE_URL}/applications`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const errorResult = await response.json();
    throw new Error(errorResult.message || 'Failed to submit application.');
  }
};
