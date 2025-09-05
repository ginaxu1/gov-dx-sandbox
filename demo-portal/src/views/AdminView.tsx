import React from 'react';

interface AdminViewProps {
    submissionIdInput: string;
    setSubmissionIdInput: (id: string) => void;
    handleApproveProvider: () => Promise<void>;
    providerState: {
        providerId: string | null;
        isSchemaSubmitted: boolean;
        submissionId: string | null;
    };
    handleApproveSchema: () => Promise<void>;
    consumerState: {
        isSubmitted: boolean;
        appId: string | null;
    };
    handleApproveApp: () => Promise<void>;
}

export default function AdminView({
    submissionIdInput,
    setSubmissionIdInput,
    handleApproveProvider,
    providerState,
    handleApproveSchema,
    consumerState,
    handleApproveApp,
}: AdminViewProps) {
    return (
        <div className="space-y-4">
            <h2 className="text-2xl font-semibold">Admin</h2>
            <p className="text-gray-600">Approve provider registrations, schemas, and consumer applications.</p>
            <div className="space-y-4 p-6 bg-red-50 rounded-lg border border-red-200">
                <h3 className="text-xl font-medium">1. Review Provider Submissions</h3>
                <div className="space-y-2">
                    <label htmlFor="submissionId" className="block text-sm font-medium text-gray-700">Submission ID:</label>
                    <input
                        id="submissionId"
                        type="text"
                        value={submissionIdInput}
                        onChange={(e) => setSubmissionIdInput(e.target.value)}
                        placeholder="Enter Submission ID"
                        className="block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                </div>
                <button
                    onClick={handleApproveProvider}
                    disabled={!submissionIdInput || providerState.providerId !== null}
                    className={`w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors ${
                        !submissionIdInput || providerState.providerId !== null ? 'bg-gray-400 cursor-not-allowed' : 'bg-red-600 hover:bg-red-700'
                    }`}
                >
                    Approve Last Provider
                </button>
                <div className="text-sm text-gray-500 mt-2">
                    {providerState.providerId ? `Provider ID: ${providerState.providerId}.` : ''}
                </div>
            </div>
            <div className="space-y-4 p-6 bg-purple-50 rounded-lg border border-purple-200">
                <h3 className="text-xl font-medium">2. Review Schema Submissions</h3>
                <button
                    onClick={handleApproveSchema}
                    disabled={!providerState.isSchemaSubmitted}
                    className={`w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors ${
                        !providerState.isSchemaSubmitted ? 'bg-gray-400 cursor-not-allowed' : 'bg-purple-600 hover:bg-purple-700'
                    }`}
                >
                    Approve Last Schema
                </button>
                <div className="text-sm text-gray-500 mt-2"></div>
            </div>
            <div className="space-y-4 p-6 bg-yellow-50 rounded-lg border border-yellow-200">
                <h3 className="text-xl font-medium">3. Review Consumer Applications</h3>
                <button
                    onClick={handleApproveApp}
                    disabled={!consumerState.isSubmitted}
                    className={`w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors ${
                        !consumerState.isSubmitted ? 'bg-gray-400 cursor-not-allowed' : 'bg-yellow-600 hover:bg-yellow-700'
                    }`}
                >
                    Approve Last Application
                </button>
                <div className="text-sm text-gray-500 mt-2"></div>
            </div>
        </div>
    );
}
