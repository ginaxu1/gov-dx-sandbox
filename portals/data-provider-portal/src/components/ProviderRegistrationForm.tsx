import React, { useState } from 'react';
import { registerProvider } from '../services/apiService';

const ProviderRegistrationForm: React.FC = () => {
    const [providerName, setProviderName] = useState('');
    const [contactEmail, setContactEmail] = useState('');
    const [phoneNumber, setPhoneNumber] = useState('');
    const [providerType, setProviderType] = useState<'government' | 'board' | 'business'>('government');

    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [successMessage, setSuccessMessage] = useState<string | null>(null);

    const handleSubmit = async (event: React.FormEvent) => {
        event.preventDefault();
        setIsLoading(true);
        setError(null);
        setSuccessMessage(null);

        try {
            const result = await registerProvider({
                providerName,
                contactEmail,
                phoneNumber,
                providerType,
            });
            setSuccessMessage(`Successfully registered provider! Your new Provider ID is: ${result.providerId}`);
            // Clear the form
            setProviderName('');
            setContactEmail('');
            setPhoneNumber('');
        } catch (err: any) {
            setError(err.message);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div>
            <h2>Register as a Data Provider</h2>
            <form onSubmit={handleSubmit}>
                {/* Form inputs for each field */}
                <div>
                    <label>Provider Name:</label>
                    <input type="text" value={providerName} onChange={(e) => setProviderName(e.target.value)} required />
                </div>
                <div>
                    <label>Contact Email:</label>
                    <input type="email" value={contactEmail} onChange={(e) => setContactEmail(e.target.value)} required />
                </div>
                <div>
                    <label>Phone Number:</label>
                    <input type="tel" value={phoneNumber} onChange={(e) => setPhoneNumber(e.target.value)} required />
                </div>
                <div>
                    <label>Provider Type:</label>
                    <select value={providerType} onChange={(e) => setProviderType(e.target.value as any)}>
                        <option value="government">Government</option>
                        <option value="board">Board</option>
                        <option value="business">Business</option>
                    </select>
                </div>

                <button type="submit" disabled={isLoading}>
                    {isLoading ? 'Submitting...' : 'Register'}
                </button>
            </form>

            {/* Display success or error messages */}
            {successMessage && <div style={{ color: 'green' }}>{successMessage}</div>}
            {error && <div style={{ color: 'red' }}>Error: {error}</div>}
        </div>
    );
};

export default ProviderRegistrationForm;
