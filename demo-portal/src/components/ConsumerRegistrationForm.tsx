import { useState, FormEvent, useEffect } from 'react';
import { submitApplication, fetchAvailableFields } from '../services/api.service';
import { LoadingSpinner, ErrorMessage } from '../utils/form-helpers';

interface ConsumerRegistrationFormProps {
    logApiCall: (message: string, status: number | string, response: object) => void;
}

// Type Definitions for fetched data
type FieldData = {
    displayName: string;
    types: Record<string, { fields: string[] }>;
};
type AvailableFieldsResponse = Record<string, FieldData>;


const SuccessMessage = ({ appId }: { appId: string }) => (
    <div className="text-center p-6 bg-green-100 text-green-800 rounded-lg">
        <h3 className="text-xl font-semibold">Application Submitted!</h3>
        <p>Your application for "{appId}" has been sent for review.</p>
    </div>
);

// Main Form Component

export default function ConsumerRegistrationForm({ logApiCall }: ConsumerRegistrationFormProps) {
    const [appId, setAppId] = useState('');
    const [availableFields, setAvailableFields] = useState<AvailableFieldsResponse | null>(null);
    const [selectedFields, setSelectedFields] = useState<Record<string, boolean>>({});
    const [status, setStatus] = useState<'loading' | 'idle' | 'submitting' | 'success' | 'error'>('loading');
    const [error, setError] = useState<string | null>(null);
    const [notification, setNotification] = useState<{ message: string; type: 'success' | 'error' } | null>(null);


    const loadFields = async () => {
        setStatus('loading');
        setError(null);
        setNotification(null);
        logApiCall('Calling GET /available-fields...', 'Pending', {});
        try {
            const fields = await fetchAvailableFields();
            setAvailableFields(fields);
            setStatus('idle');
            logApiCall('GET /available-fields', 200, { fields });
        } catch (err: any) {
            setError('Could not load available data fields. Please try again later.');
            setStatus('error');
            logApiCall('GET /available-fields', 'Error', { message: err.message });
        }
    };

    // Fetch the available fields when the component first mounts
    useEffect(() => {
        loadFields();
    }, []);

    const handleCheckboxChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const { name, checked } = event.target;
        setSelectedFields(prev => ({ ...prev, [name]: checked }));
    };

    const handleSubmit = async (event: FormEvent) => {
        event.preventDefault();
        setError(null);
        setNotification(null);
        setStatus('submitting');

        const requiredFields = Object.keys(selectedFields).reduce((acc, key) => {
            if (selectedFields[key]) {
                acc[key] = {}; // The backend expects an empty object for each selected field
            }
            return acc;
        }, {} as Record<string, object>);

        if (Object.keys(requiredFields).length === 0) {
            setError('You must select at least one data field to request.');
            setStatus('idle'); // Revert to idle to allow re-submission
            setNotification({ message: 'You must select at least one data field to request.', type: 'error' });
            return;
        }

        const payload = { appId, requiredFields };
        logApiCall('Calling POST /applications...', 'Pending', payload);

        try {
            await submitApplication(payload);
            setStatus('success');
            setNotification({ message: 'Application submitted successfully!', type: 'success' });
            setAppId('');
            setSelectedFields({});
            logApiCall('POST /applications', 201, { appId, requiredFields, status: 'pending' });
        } catch (err: any) {
            setError(err.message);
            setStatus('error');
            logApiCall('POST /applications', 'Error', { message: err.message });
            setNotification({ message: 'An error occurred during submission. Please check the log.', type: 'error' });
        }
    };

    // Render different UI based on the current status

    if (status === 'loading') {
        return <LoadingSpinner text="Loading available data fields..." />;
    }

    if (status === 'error' && !availableFields) {
        return <ErrorMessage message={error!} onRetry={loadFields} />;
    }

    if (status === 'success') {
        return <SuccessMessage appId={appId} />;
    }

    return (
        <form onSubmit={handleSubmit} className="space-y-6 w-full text-left">
            <div className="space-y-4 p-6 bg-indigo-50 rounded-lg border border-indigo-200">
                <h3 className="text-xl font-medium">1. Submit Data Application</h3>
                {notification && (
                    <div className={`p-4 rounded-md ${notification.type === 'success' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                        {notification.message}
                    </div>
                )}
                <div className="space-y-2">
                    <label htmlFor="appId" className="block text-sm font-medium text-gray-700">Application ID</label>
                    <input
                        id="appId"
                        type="text"
                        value={appId}
                        onChange={(e) => setAppId(e.target.value)}
                        placeholder="e.g., passport-renewal-app"
                        required
                        className="block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                </div>

                <div className="space-y-4">
                    <label className="block text-sm font-medium text-gray-700">Select Required Fields</label>
                    <div className="p-4 border border-gray-300 rounded-md max-h-80 overflow-y-auto space-y-4 bg-gray-50">
                        {availableFields && Object.entries(availableFields).map(([providerId, providerData]) => (
                            <div key={providerId}>
                                <h4 className="font-semibold text-blue-600">{providerData.displayName}</h4>
                                <div className="pl-4 mt-2 space-y-2 border-l-2 border-gray-200">
                                    {Object.entries(providerData.types).map(([typeName, typeData]) => (
                                        <div key={typeName}>
                                            <p className="text-sm font-medium text-gray-500">{typeName}</p>
                                            <div className="pl-4 mt-1 space-y-1">
                                                {typeData.fields.map(fieldName => {
                                                    const fullFieldName = `${providerId}.${typeName}.${fieldName}`;
                                                    return (
                                                        <label key={fullFieldName} className="flex items-center space-x-2 font-normal text-gray-800">
                                                            <input
                                                                type="checkbox"
                                                                name={fullFieldName}
                                                                checked={!!selectedFields[fullFieldName]}
                                                                onChange={handleCheckboxChange}
                                                                className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                                                            />
                                                            <span>{fieldName}</span>
                                                        </label>
                                                    );
                                                })}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        ))}
                    </div>
                </div>

                <button
                    type="submit"
                    disabled={status === 'submitting'}
                    className="w-full flex justify-center py-3 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
                >
                    {status === 'submitting' ? <LoadingSpinner text="Submitting..." /> : 'Submit for Review'}
                </button>
            </div>
        </form>
    );
}
