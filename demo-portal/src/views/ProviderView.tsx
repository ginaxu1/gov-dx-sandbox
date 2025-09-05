import React from 'react';
import ProviderSubmissionForm from '../components/ProviderSubmissionForm';

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
    return (
        <div className="space-y-4">
            <h2 className="text-2xl font-semibold">Data Provider</h2>
            <p className="text-gray-600">Register as a new provider and submit your schema for approval.</p>
            {!showProviderForm ? (
                <div className="space-y-4 p-6 bg-blue-50 rounded-lg border border-blue-200">
                    <h3 className="text-xl font-medium">1. Register as a Provider</h3>
                    <button
                        onClick={() => setShowProviderForm(true)}
                        className={`w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors bg-blue-600 hover:bg-blue-700`}
                    >
                        Register Provider
                    </button>
                    <div className="text-sm text-gray-500 mt-2">
                        {providerState.submissionId ? `Submission ID: ${providerState.submissionId}. Now go to Admin view to approve.` : ''}
                    </div>
                </div>
            ) : (
                <ProviderSubmissionForm logApiCall={logApiCall} onSuccess={handleProviderSubmitSuccess} />
            )}
            <div className="space-y-4 p-6 bg-green-50 rounded-lg border border-green-200">
                <h3 className="text-xl font-medium">2. Submit Schema (Requires Admin Approval First)</h3>
                <button
                    onClick={handleSubmitSchema}
                    disabled={!providerState.providerId || providerState.isSchemaSubmitted}
                    className={`w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors ${
                        !providerState.providerId || providerState.isSchemaSubmitted ? 'bg-gray-400 cursor-not-allowed' : 'bg-green-600 hover:bg-green-700'
                    }`}
                >
                    Submit Schema
                </button>
                <div className="text-sm text-gray-500 mt-2">
                    {providerState.isSchemaSubmitted ? 'Schema submitted for approval. Go to Admin view to approve.' : ''}
                </div>
            </div>
        </div>
    );
}
