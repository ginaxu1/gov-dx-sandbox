import { useState } from 'react';
import ProviderView from './views/ProviderView';
import ConsumerView from './views/ConsumerView';
import AdminView from './views/AdminView';

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
  const [submissionIdInput, setSubmissionIdInput] = useState<string>('');
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
      } else {
        throw new Error(result.message || `HTTP ${response.status}: ${response.statusText}`);
      }
    } catch (error: any) {
      logApiCall('POST /provider-submissions/review', 'Error', { message: error.message });
      throw error; // Re-throw to let the notification wrapper handle it
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
      
      if (!response.ok) {
        throw new Error(result.message || `HTTP ${response.status}: ${response.statusText}`);
      }
    } catch (error: any) {
      logApiCall('POST /provider-schemas/review', 'Error', { message: error.message });
      throw error; // Re-throw to let the notification wrapper handle it
    }
  };

  const handleSubmitApp = async (payload: { appId: string, requiredFields: object }) => {
    logApiCall('Calling POST /applications...', 'Pending', payload);
    try {
      const response = await fetch(`${API_BASE_URL}/applications`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      const result = await response.json();
      logApiCall('POST /applications', response.status, result);

      if (response.ok) {
        setConsumerState({ appId: result.data.appId, isSubmitted: true });
      } else {
        throw new Error(result.message || `HTTP ${response.status}: ${response.statusText}`);
      }
    } catch (error: any) {
      logApiCall('POST /applications', 'Error', { message: error.message });
      throw error; // Re-throw to let the form handle it
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
      
      if (!response.ok) {
        throw new Error(result.message || `HTTP ${response.status}: ${response.statusText}`);
      }
    } catch (error: any) {
      logApiCall('POST /applications/review', 'Error', { message: error.message });
      throw error; // Re-throw to let the notification wrapper handle it
    }
  };

  const renderRoleView = () => {
    switch (role) {
      case 'provider':
        return (
          <ProviderView 
            showProviderForm={showProviderForm} 
            setShowProviderForm={setShowProviderForm}
            logApiCall={logApiCall}
            handleProviderSubmitSuccess={handleProviderSubmitSuccess}
            providerState={providerState}
            handleSubmitSchema={handleSubmitSchema}
          />
        );
      case 'consumer':
        return (
          <ConsumerView 
            logApiCall={logApiCall}
            consumerState={consumerState}
            handleSubmitApp={handleSubmitApp}
          />
        );
      case 'admin':
        return (
          <AdminView 
            submissionIdInput={submissionIdInput}
            setSubmissionIdInput={setSubmissionIdInput}
            handleApproveProvider={handleApproveProvider}
            providerState={providerState}
            handleApproveSchema={handleApproveSchema}
            consumerState={consumerState}
            handleApproveApp={handleApproveApp}
          />
        );
      default:
        return null;
    }
  };

  return (
    <div className="w-full max-w-4xl bg-white rounded-lg shadow-xl p-8 space-y-8">
      <header className="flex justify-between items-center pb-4 border-b-2 border-gray-200">
        <h1 className="text-3xl font-bold text-gray-800">OpenDIF Portal</h1>
        <div className="flex items-center space-x-2">
          <label htmlFor="role-selector" className="text-sm font-medium text-gray-600">Role:</label>
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
