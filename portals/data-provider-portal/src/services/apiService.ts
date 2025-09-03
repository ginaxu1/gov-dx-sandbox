// Define a type for the provider registration payload
// This should match the expected request body of your POST /providers endpoint
interface ProviderRegistrationPayload {
    providerName: string;
    contactEmail: string;
    phoneNumber: string;
    providerType: 'government' | 'board' | 'business';
}

/**
 * Registers a new Data Provider by sending a POST request to the BFF (api-server)
 * @param providerData The data for the new provider
 * @returns A promise that resolves with the server's response data
 */
export const registerProvider = async (providerData: ProviderRegistrationPayload) => {
    const response = await fetch('/api/providers', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(providerData),
    });

    // Handle non-successful responses
    if (!response.ok) {
        const errorData = await response.json();
        // Throw an error with the message from the server
        throw new Error(errorData.error || 'Failed to register provider');
    }

    // Return the successful response data
    return response.json();
};

// TODO: add other API functions in next PR...
// export const submitProviderSchema = async (schemaData) => { ... }
