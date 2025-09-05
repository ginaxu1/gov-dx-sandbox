import { useState } from 'react';
import ProviderSubmissionForm from './components/ProviderSubmissionForm';
import ConsumerRegistrationForm from './components/ConsumerRegistrationForm';

const API_BASE_URL = 'http://localhost:3000';

interface ProviderState {
  submissionId: string | null;
  providerId: string | null;
  isRegistered: boolean;
  isSchemaSubmitted: boolean;
}

interface ConsumerState {
  appId: string | null;
  isSubmitted: boolean;
}

function App() {
  const [role, setRole] = useState<'provider' | 'consumer' | 'admin'>('provider');
  const [log, setLog] = useState<string[]>([]);
  const [showProviderForm, setShowProviderForm] = useState(false);
  const [submissionIdInput, setSubmissionIdInput] = useState<string>(''); // New state for input field
  const [providerState, setProviderState] = useState<ProviderState>({
    submissionId: null,
    providerId: null,
    isRegistered: false,
    isSchemaSubmitted: false,
  });
  const [consumerState, setConsumerState] = useState<ConsumerState>({
    appId: null,
    isSubmitted: false,
  });

  const logApiCall = (message: string, status: number | string, response: object) => {
    const timestamp = new Date().toLocaleTimeString();
    setLog((prevLog) => [
      ...prevLog,
      `[${timestamp}] ${message}\nStatus: ${status}\nResponse: ${JSON.stringify(response, null, 2)}`,
    ]);
  };

  const handleProviderSubmitSuccess = (submissionId: string) => {
    setProviderState(prev => ({ ...prev, submissionId }));
    setShowProviderForm(false);
  };
  
  const handleApproveProvider = async () => {
    const idToApprove = submissionIdInput || providerState.submissionId;
    if (!idToApprove) return;
    logApiCall(`Calling POST /provider-submissions/${idToApprove}/review...`, 'Pending', {});
    try {
      const response = await fetch(`${API_BASE_URL}/provider-submissions/${idToApprove}/review`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ decision: 'approve' }),
      });
      const result = await response.json();
      logApiCall(`POST /provider-submissions/${idToApprove}/review`, response.status, result);

      if (response.ok) {
        setProviderState((prev) => ({
          ...prev,
          providerId: result.data.providerId,
        }));
      }
    } catch (error: any) {
      logApiCall('POST /provider-submissions/review', 'Error', { message: error.message });
    }
  };

  const handleSubmitSchema = async () => {
    if (!providerState.providerId) return;
    logApiCall(`Calling POST /providers/${providerState.providerId}/schemas...`, 'Pending', {});
    try {
      const response = await fetch(`${API_BASE_URL}/providers/${providerState.providerId}/schemas`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          fieldConfigurations: {
            PersonData: {
              nic: { source: 'authoritative', isOwner: true, description: 'National ID Card number.' },
            },
          },
        }),
      });
      const result = await response.json();
      logApiCall(`POST /providers/${providerState.providerId}/schemas`, response.status, result);

      if (response.ok) {
        setProviderState((prev) => ({ ...prev, isSchemaSubmitted: true }));
      }
    } catch (error: any) {
      logApiCall('POST /providers/schemas', 'Error', { message: error.message });
    }
  };

  const handleApproveSchema = async () => {
    if (!providerState.providerId) return;
    logApiCall(`Calling POST /provider-schemas/${providerState.providerId}/review...`, 'Pending', {});
    try {
      const response = await fetch(`${API_BASE_URL}/provider-schemas/${providerState.providerId}/review`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ decision: 'approve' }),
      });
      const result = await response.json();
      logApiCall(`POST /provider-schemas/${providerState.providerId}/review`, response.status, result);
    } catch (error: any) {
    logApiCall('POST /provider-schemas/review', 'Error', { message: error.message });
    }
  };

  const handleSubmitApp = async () => {
    logApiCall('Calling POST /applications...', 'Pending', {});
    try {
      const response = await fetch(`${API_BASE_URL}/applications`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          appId: `app_${Math.floor(Math.random() * 1000)}`,
          requiredFields: { 'drp.PersonData.nic': {} },
        }),
      });
      const result = await response.json();
      logApiCall('POST /applications', response.status, result);

      if (response.ok) {
        setConsumerState({ appId: result.data.appId, isSubmitted: true });
      }
    } catch (error: any) {
      logApiCall('POST /applications', 'Error', { message: error.message });
    }
  };

  const handleApproveApp = async () => {
    if (!consumerState.appId) return;
    logApiCall(`Calling POST /applications/${consumerState.appId}/review...`, 'Pending', {});
    try {
      const response = await fetch(`${API_BASE_URL}/applications/${consumerState.appId}/review`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ decision: 'approve' }),
      });
      const result = await response.json();
      logApiCall(`POST /applications/${consumerState.appId}/review`, response.status, result);
    } catch (error: any) {
      logApiCall('POST /applications/review', 'Error', { message: error.message });
    }
  };

  const renderRoleView = () => {
    switch (role) {
      case 'provider':
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
      case 'consumer':
        return (
          <div className="space-y-4">
            <h2 className="text-2xl font-semibold">Data Consumer</h2>
            <p className="text-gray-600">Submit an application to access data from an approved provider.</p>
            <ConsumerRegistrationForm logApiCall={logApiCall} />
          </div>
        );
      case 'admin':
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
      default:
        return null;
    }
  };

  return (
    <div className="w-full max-w-4xl bg-white rounded-lg shadow-xl p-8 space-y-8">
      <header className="flex justify-between items-center pb-4 border-b-2 border-gray-200">
        <h1 className="text-3xl font-bold text-gray-800">API Server Demo</h1>
        <div className="flex items-center space-x-2">
          <label htmlFor="role-selector" className="text-sm font-medium text-gray-600">Current Role:</label>
          <select
            id="role-selector"
            value={role}
            onChange={(e) => setRole(e.target.value as 'provider' | 'consumer' | 'admin')}
            className="rounded-md border border-gray-300 p-2 text-sm bg-gray-50 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="provider">Data Provider</option>
            <option value="consumer">Data Consumer</option>
            <option value="admin">Admin</option>
          </select>
        </div>
      </header>

      <main className="space-y-6">
        {renderRoleView()}
      </main>

      <div className="mt-8 pt-4 border-t-2 border-gray-200">
        <h2 className="text-2xl font-semibold">API Call Log</h2>
        <pre className="mt-4 p-4 bg-gray-800 text-green-400 rounded-md overflow-x-auto text-sm whitespace-pre-wrap">
          {log.join('\n\n')}
        </pre>
      </div>
    </div>
  );
}

export default App;
