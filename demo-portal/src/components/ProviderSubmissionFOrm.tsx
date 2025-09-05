import { useState, FormEvent } from 'react';
import { submitProviderSubmission } from '../services/api.service';
import { ProviderType } from '../../../api-server/src/types';
import { LoadingSpinner, ErrorMessage } from '../utils/form-helpers';

interface ProviderSubmissionFormProps {
    logApiCall: (message: string, status: number | string, response: object) => void;
    onSuccess: (submissionId: string) => void;
}

const SuccessMessage = ({ submissionId }: { submissionId: string }) => (
    <div className="text-center p-6 bg-green-100 text-green-800 rounded-lg">
        <h3 className="text-xl font-semibold">Registration Submitted!</h3>
        <p>Your submission ID is: <code className="font-mono bg-gray-200 px-1 py-0.5 rounded text-sm">{submissionId}</code></p>
        <p className="mt-2">Go to the **Admin** view to approve it.</p>
    </div>
);

export default function ProviderSubmissionForm({ logApiCall, onSuccess }: ProviderSubmissionFormProps) {
    const [providerName, setProviderName] = useState('');
    const [contactEmail, setContactEmail] = useState('');
    const [phoneNumber, setPhoneNumber] = useState('');
    const [providerType, setProviderType] = useState<ProviderType>('business');
    const [status, setStatus] = useState<'idle' | 'submitting' | 'success' | 'error'>('idle');
    const [error, setError] = useState<string | null>(null);
    const [notification, setNotification] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

    const handleSubmit = async (event: FormEvent) => {
        event.preventDefault();
        setError(null);
        setNotification(null);
        setStatus('submitting');
        
        const payload = { providerName, contactEmail, phoneNumber, providerType };
        logApiCall('Calling POST /provider-submissions...', 'Pending', payload);

        try {
            const result = await submitProviderSubmission(payload);
            setStatus('success');
            logApiCall('POST /provider-submissions', 202, result);
            setNotification({ message: 'Registration submitted successfully!', type: 'success' });
            onSuccess(result.submissionId);
            setProviderName('');
            setContactEmail('');
            setPhoneNumber('');
            setProviderType('business');
        } catch (err: any) {
            setError(err.message);
            setStatus('error');
            logApiCall('POST /provider-submissions', 'Error', { message: err.message });
            setNotification({ message: 'An error occurred during submission. Please check the log.', type: 'error' });
        }
    };

    if (status === 'submitting') {
        return <LoadingSpinner text="Submitting registration..." />;
    }

    if (status === 'success') {
        return <SuccessMessage submissionId={providerName} />;
    }

    return (
        <div className="space-y-4 p-6 bg-blue-50 rounded-lg border border-blue-200">
            <h3 className="text-xl font-medium">1. Register as a Provider</h3>
            <form onSubmit={handleSubmit} className="space-y-4">
                {notification && (
                    <div className={`p-4 rounded-md ${notification.type === 'success' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                        {notification.message}
                    </div>
                )}
                <div>
                    <label htmlFor="providerName" className="block text-sm font-medium text-gray-700">Provider Name</label>
                    <input
                        type="text"
                        id="providerName"
                        value={providerName}
                        onChange={(e) => setProviderName(e.target.value)}
                        placeholder="e.g., Department of Registration"
                        required
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2"
                    />
                </div>
                <div>
                    <label htmlFor="contactEmail" className="block text-sm font-medium text-gray-700">Contact Email</label>
                    <input
                        type="email"
                        id="contactEmail"
                        value={contactEmail}
                        onChange={(e) => setContactEmail(e.target.value)}
                        placeholder="e.g., contact@example.com"
                        required
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2"
                    />
                </div>
                <div>
                    <label htmlFor="phoneNumber" className="block text-sm font-medium text-gray-700">Phone Number</label>
                    <input
                        type="tel"
                        id="phoneNumber"
                        value={phoneNumber}
                        onChange={(e) => setPhoneNumber(e.target.value)}
                        placeholder="e.g., 555-123-4567"
                        required
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2"
                    />
                </div>
                <div>
                    <label htmlFor="providerType" className="block text-sm font-medium text-gray-700">Provider Type</label>
                    <select
                        id="providerType"
                        value={providerType}
                        onChange={(e) => setProviderType(e.target.value as ProviderType)}
                        required
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2"
                    >
                        <option value="business">Business</option>
                        <option value="government">Government</option>
                        <option value="board">Board</option>
                    </select>
                </div>
                {status === 'error' && <ErrorMessage message={error!} />}
                <button
                    type="submit"
                    disabled={status === 'submitting'}
                    className={`w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors ${
                        status === 'submitting' ? 'bg-gray-400 cursor-not-allowed' : 'bg-blue-600 hover:bg-blue-700'
                    }`}
                >
                    {status === 'submitting' ? 'Submitting...' : 'Submit Registration'}
                </button>
            </form>
        </div>
    );
}
