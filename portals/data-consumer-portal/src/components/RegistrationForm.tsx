import { useState, FormEvent, useEffect } from 'react';
import { submitApplication, fetchAvailableFields } from '../services/api.service';

// --- Type Definitions for fetched data ---
// These types should ideally live in a central types file.
type FieldData = {
    displayName: string;
    types: Record<string, { fields: string[] }>;
};
type AvailableFieldsResponse = Record<string, FieldData>;


// --- Helper Components for UI States ---

const LoadingSpinner = ({ text }: { text: string }) => (
    <div className="flex flex-col items-center justify-center p-8 text-center">
        <svg className="animate-spin h-8 w-8 text-primary mb-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <p className="text-muted-foreground">{text}</p>
    </div>
);

const ErrorMessage = ({ message, onRetry }: { message: string, onRetry?: () => void }) => (
    <div className="text-center p-6 bg-destructive/10 text-destructive rounded-lg">
        <h3 className="text-xl font-semibold">An Error Occurred</h3>
        <p>{message}</p>
        {onRetry && (
            <button onClick={onRetry} className="mt-4 px-4 py-2 bg-destructive text-destructive-foreground rounded-md hover:bg-destructive/90">
                Try Again
            </button>
        )}
    </div>
);

const SuccessMessage = ({ appId }: { appId: string }) => (
    <div className="text-center p-6 bg-green-100 text-green-800 rounded-lg">
        <h3 className="text-xl font-semibold">Application Submitted!</h3>
        <p>Your application for "{appId}" has been sent for review.</p>
    </div>
);

// --- Main Form Component ---

export default function ConsumerRegistrationForm() {
    const [appId, setAppId] = useState('');
    const [availableFields, setAvailableFields] = useState<AvailableFieldsResponse | null>(null);
    const [selectedFields, setSelectedFields] = useState<Record<string, boolean>>({});
    const [status, setStatus] = useState<'loading' | 'idle' | 'submitting' | 'success' | 'error'>('loading');
    const [error, setError] = useState<string | null>(null);

    const loadFields = async () => {
        setStatus('loading');
        setError(null);
        try {
            const fields = await fetchAvailableFields();
            setAvailableFields(fields);
            setStatus('idle');
        } catch (err: any) {
            setError('Could not load available data fields. Please try again later.');
            setStatus('error');
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
            return;
        }

        try {
            await submitApplication({ appId, requiredFields });
            setStatus('success');
        } catch (err: any) {
            setError(err.message);
            setStatus('error');
        }
    };

    // --- Render different UI based on the current status ---

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
            {/* Application ID Input */}
            <div className="space-y-2">
                <label htmlFor="appId" className="block text-sm font-medium text-foreground">Application ID</label>
                <input
                    id="appId"
                    type="text"
                    value={appId}
                    onChange={(e) => setAppId(e.target.value)}
                    placeholder="e.g., passport-renewal-app"
                    required
                    className="block w-full px-3 py-2 bg-background border border-border rounded-md shadow-sm placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                />
            </div>

            {/* Dynamic Checkbox Section */}
            <div className="space-y-4">
                <label className="block text-sm font-medium text-foreground">Select Required Fields</label>
                <div className="p-4 border border-border rounded-md max-h-80 overflow-y-auto space-y-4 bg-background">
                    {availableFields && Object.entries(availableFields).map(([providerId, providerData]) => (
                        <div key={providerId}>
                            <h4 className="font-semibold text-primary">{providerData.displayName}</h4>
                            <div className="pl-4 mt-2 space-y-2 border-l-2 border-border">
                                {Object.entries(providerData.types).map(([typeName, typeData]) => (
                                    <div key={typeName}>
                                        <p className="text-sm font-medium text-muted-foreground">{typeName}</p>
                                        <div className="pl-4 mt-1 space-y-1">
                                            {typeData.fields.map(fieldName => {
                                                const fullFieldName = `${providerId}.${typeName}.${fieldName}`;
                                                return (
                                                    <label key={fullFieldName} className="flex items-center space-x-2 font-normal">
                                                        <input
                                                            type="checkbox"
                                                            name={fullFieldName}
                                                            checked={!!selectedFields[fullFieldName]}
                                                            onChange={handleCheckboxChange}
                                                            className="h-4 w-4 rounded border-border text-primary focus:ring-ring"
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

            {status === 'error' && <ErrorMessage message={error!} />}

            <button
                type="submit"
                disabled={status === 'submitting'}
                className="w-full flex justify-center py-3 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-primary-foreground bg-primary hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-ring disabled:opacity-50"
            >
                {status === 'submitting' && <LoadingSpinner text="Submitting..." />}
                {status !== 'submitting' && 'Submit for Review'}
            </button>
        </form>
    );
}

