import { useState } from 'react';

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
        appId: string | null;
        isSubmitted: boolean;
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
    const [notification, setNotification] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

    const handleApproveProviderWithNotification = async () => {
        try {
            await handleApproveProvider();
            setNotification({ message: 'Provider submission approved', type: 'success' });
        } catch (error: any) {
            setNotification({ message: `An error occurred: ${error.message}`, type: 'error' });
        }
    };

    const handleApproveSchemaWithNotification = async () => {
        try {
            await handleApproveSchema();
            setNotification({ message: 'Provider schema approved', type: 'success' });
        } catch (error: any) {
            setNotification({ message: `An error occurred: ${error.message}`, type: 'error' });
        }
    };

    const handleApproveAppWithNotification = async () => {
        try {
            await handleApproveApp();
            setNotification({ message: 'Consumer application approved', type: 'success' });
        } catch (error: any) {
            setNotification({ message: `An error occurred: ${error.message}`, type: 'error' });
        }
    };

    return (
        <div className="space-y-4">
            <h2 className="text-2xl font-semibold">Admin</h2>
            <p className="text-gray-600">Approve provider registrations, schemas, and consumer applications</p>

            {notification && (
                <div className={`p-4 rounded-md ${notification.type === 'success' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                    {notification.message}
                </div>
            )}
            
            <div className="space-y-4 p-6 bg-red-50 rounded-lg border border-red-200">
                <h3 className="text-xl font-medium">Review Provider Submissions</h3>
                <div className="space-y-2">
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
                    onClick={handleApproveProviderWithNotification}
                    disabled={!submissionIdInput || providerState.providerId !== null}
                    className={`w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors ${
                        !submissionIdInput || providerState.providerId !== null ? 'bg-gray-400 cursor-not-allowed' : 'bg-red-600'
                    }`}
                >
                    Approve Provider
                </button>
                <div className="text-sm text-gray-500 mt-2">
                    {providerState.providerId ? `Provider ID: ${providerState.providerId}` : ''}
                </div>
            </div>
            <div className="space-y-4 p-6 bg-purple-50 rounded-lg border border-purple-200">
                <h3 className="text-xl font-medium">Review Schema Submissions</h3>
                <button
                    onClick={handleApproveSchemaWithNotification}
                    disabled={!providerState.isSchemaSubmitted}
                    className={`w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors ${
                        !providerState.isSchemaSubmitted ? 'bg-gray-400 cursor-not-allowed' : 'bg-purple-600 hover:bg-purple-700'
                    }`}
                >
                    Approve Schema
                </button>
                <div className="text-sm text-gray-500 mt-2">
                    {providerState.providerId ? `Provider ID: ${providerState.providerId}` : ''}
                </div>
            </div>
            <div className="space-y-4 p-6 bg-yellow-50 rounded-lg border border-yellow-200">
                <h3 className="text-xl font-medium">Review Consumer Applications</h3>
                <button
                    onClick={handleApproveAppWithNotification}
                    disabled={!consumerState.appId}
                    className={`w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors ${
                        !consumerState.appId ? 'bg-gray-400 cursor-not-allowed' : 'bg-yellow-600 hover:bg-yellow-700'
                    }`}
                >
                    Approve Application
                </button>
                <div className="text-sm text-gray-500 mt-2">
                    {consumerState.appId ? `App ID: ${consumerState.appId}` : 'No application to approve'}
                </div>
            </div>
        </div>
    );
}
