import { useState } from 'react';
import ProviderSubmissionForm from '../components/ProviderRegistrationForm';
import ProviderSchemaJsonForm from '../components/ProviderSchemaJsonForm';

interface ProviderViewProps {
    showProviderForm: boolean;
    setShowProviderForm: (show: boolean) => void;
    logApiCall: (message: string, status: number | string, response: object) => void;
    handleProviderSubmitSuccess: (submissionId: string) => void;
    providerState: {
        providerId: string | null;
        isSchemaSubmitted: boolean;
        submissionId: string | null;
    };
    handleSubmitSchema: () => Promise<void>;
}

export default function ProviderView({
    showProviderForm,
    setShowProviderForm,
    logApiCall,
    handleProviderSubmitSuccess,
    providerState,
    handleSubmitSchema,
}: ProviderViewProps) {
    const [showSchemaForm, setShowSchemaForm] = useState(false);
    
    const handleSchemaSubmitSuccess = () => {
        handleSubmitSchema();
        setShowSchemaForm(false);
    }
    
    const handleSubmitSchemaButtonClick = () => {
        if (!providerState.isSchemaSubmitted) {
            setShowSchemaForm(true);
        }
    };
    
    return (
        <div className="space-y-4">
            <h2 className="text-2xl font-semibold">Data Provider</h2>
            <p className="text-gray-600">Register as a provider and submit your schema for approval</p>
            {!showProviderForm ? (
                <div className="space-y-4 p-6 bg-blue-50 rounded-lg border border-blue-200">
                    <h3 className="text-xl font-medium">Register as a Provider</h3>
                    {providerState.submissionId && (
                        <div className="text-center p-6 bg-green-100 text-green-800 rounded-lg">
                            <h3 className="text-xl font-semibold">Application Submitted!</h3>
                            <p>Please wait for the Admin to approve</p>
                            <p className="mt-2 text-sm">Submission ID: {providerState.submissionId}</p>
                        </div>
                    )}
                    <button
                        onClick={() => setShowProviderForm(true)}
                        className={`w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors bg-blue-600 hover:bg-blue-700`}
                    >
                        Register Provider
                    </button>
                </div>
            ) : (
                <ProviderSubmissionForm logApiCall={logApiCall} onSuccess={handleProviderSubmitSuccess} />
            )}
            {!showSchemaForm ? (
                <div className="space-y-4 p-6 bg-green-50 rounded-lg border border-green-200">
                    <h3 className="text-xl font-medium">Submit Schema (Requires Admin Approval)</h3>
                    <button
                        onClick={handleSubmitSchemaButtonClick}
                        className={`w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors ${
                            'bg-green-600 hover:bg-green-700'
                        }`}
                    >
                        Submit Schema
                    </button>
                    <div className="text-sm text-gray-500 mt-2">
                        {!providerState.providerId 
                            ? 'Please register as a provider and get admin approval first' 
                            : providerState.isSchemaSubmitted 
                                ? 'Submission received! Please wait for the Admin to approve' 
                                : ''
                        }
                    </div>
                </div>
            ) : (
                <ProviderSchemaJsonForm 
                    providerId={providerState.providerId}
                    logApiCall={logApiCall}
                    onSuccess={handleSchemaSubmitSuccess}
                />
            )}
        </div>
    );
}
