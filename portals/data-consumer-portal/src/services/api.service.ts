/**
 * @file Centralizes all API communication for the Data Consumer Portal
 */

interface ApplicationPayload {
  appId: string;
  requiredFields: object;
}

/**
 * Fetches the available data fields from all providers
 * This is used to build the field selection UI
 * @returns {Promise<any>} A promise that resolves with ONLY the data part of the API response
 */
export const fetchAvailableFields = async (): Promise<any> => {
  const response = await fetch('/api/available-fields');
  // First, get the full JSON response from the server
  const result = await response.json();

  // Check for network errors or if the API returned a 'status: "error"' envelope
  if (!response.ok || result.status === 'error') {
    // Throw an error with the message from the API, or a default one
    throw new Error(result.message || 'Failed to fetch available fields');
  }

  // **THE FIX:** Return only the .data property from the envelope
  return result.data;
};


/**
 * Submits a new data consumer application to the backend API
 */
export const submitApplication = async (applicationData: ApplicationPayload): Promise<any> => {
  const response = await fetch('/api/applications', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(applicationData),
  });

  const result = await response.json();
  if (!response.ok || result.status === 'error') {
    throw new Error(result.message || 'Failed to submit application');
  }
  return result.data;
};

