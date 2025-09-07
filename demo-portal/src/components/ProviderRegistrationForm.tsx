import { useState, FormEvent } from 'react';
import { submitProviderSubmission } from '../services/api.service';
import { ProviderType } from '../../../api-server/src/types';
import { LoadingSpinner } from '../utils/form-helpers';

interface ProviderRegistrationFormProps {
    logApiCall: (message: string, status: number | string, response: object) => void;
    onSuccess: (submissionId: string) => void;
}

const SuccessMessage = ({ submissionId }: { submissionId: string }) => (
    <div className="text-center p-6 bg-green-100 text-green-800 rounded-lg">
        <h3 className="text-xl font-semibold">Application Submitted!</h3>
        <p>Please wait for the Admin to approv</p>
        <p className="mt-2 text-sm">Submission ID: {submissionId}</p>
    </div>
);

export default function ProviderRegistrationForm({ logApiCall, onSuccess }: ProviderRegistrationFormProps) {
    const [providerName, setProviderName] = useState('');
    const [contactEmail, setContactEmail] = useState('');
    const [phoneNumber, setPhoneNumber] = useState('');
    const [providerType, setProviderType] = useState<ProviderType>('business');
    const [status, setStatus] = useState<'idle' | 'submitting' | 'success' | 'error'>('idle');
    const [error, setError] = useState<string | null>(null);
    const [notification, setNotification] = useState<{ message: string; type: 'success' | 'error' } | null>(null);
    const [submissionId, setSubmissionId] = useState<string>('');

    const handlePhoneNumberChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        // Only allow digits
        const digitsOnly = value.replace(/\D/g, '');
        setPhoneNumber(digitsOnly);
    };

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
            setSubmissionId(result.submissionId);
            logApiCall('POST /provider-submissions', 202, result);
            setNotification({ message: 'Registration submitted', type: 'success' });
            onSuccess(result.submissionId);
            setProviderName('');
            setContactEmail('');
            setPhoneNumber('');
            setProviderType('business');
        } catch (err: any) {
            setError(err.message);
            setStatus('error');
            logApiCall('POST /provider-submissions', 'Error', { message: err.message });
            setNotification({ message: err.message, type: 'error' });
        }
    };

    if (status === 'submitting') {
        return <LoadingSpinner text="Submitting registration..." />;
    }

    if (status === 'success') {
        return <SuccessMessage submissionId={submissionId} />;
    }

    const buttonClasses = `w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors ${
        status !== 'idle'
            ? 'bg-gray-400 cursor-not-allowed'
            : 'bg-blue-600 hover:bg-blue-700'
    }`;

    return (
        <div className="space-y-4 p-6 bg-blue-50 rounded-lg border border-blue-200">
            <h3 className="text-xl font-medium">Register as a Provider</h3>
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
                        placeholder="Department of Registration"
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
                        placeholder="contact@email.com"
                        required
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blu
                        e-500 focus:ring-blue-500 sm:text-sm p-2"
                    />
                </div>
                <div>
                    <label htmlFor="phoneNumber" className="block text-sm font-medium text-gray-700">Phone Number</label>
                    <input
                        type="tel"
                        id="phoneNumber"
                        value={phoneNumber}
                        onChange={handlePhoneNumberChange}
                        placeholder="77101010"
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
                <button
                    type="submit"
                    disabled={status !== 'idle'}
                    className={buttonClasses}
                >
                    {status !== 'idle' ? 'Submitting...' : 'Submit Registration'}
                </button>
            </form>
        </div>
    );
}
