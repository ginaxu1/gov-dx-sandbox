import { useState, FormEvent } from 'react';
import { submitProviderSchema } from '../services/api.service';
import { LoadingSpinner, ErrorMessage } from '../utils/form-helpers';

interface ProviderSchemaJsonFormProps {
  providerId: string | null;
  logApiCall: (message: string, status: number | string, response: object) => void;
  onSuccess: () => void;
}

const SchemaSubmittedMessage = () => (
    <div className="text-center p-6 bg-green-100 text-green-800 rounded-lg">
        <h3 className="text-xl font-semibold">Schema Submitted!</h3>
        <p>Your schema has been submitted for review</p>
        <p className="mt-2">Go to the **Admin** view to approve it</p>
    </div>
);

export default function ProviderSchemaJsonForm({ providerId, logApiCall, onSuccess }: ProviderSchemaJsonFormProps) {
  const [jsonPayload, setJsonPayload] = useState(`{
  "fieldConfigurations": {
    "PersonData": {
      "nic": {
        "source": "fallback",
        "isOwner": false,
        "description": "National ID Card number"
      },
      "fullName": {
        "source": "fallback",
        "isOwner": true,
        "description": "Full name of the person"
      },
      "otherNames": {
        "source": "authoritative",
        "isOwner": false,
        "description": "Other names or aliases"
      },
      "permanentAddress": {
        "source": "authoritative",
        "isOwner": true,
        "description": "Permanent residential address"
      }
    },
    "Query": {
      "person": {
        "source": "authoritative",
        "isOwner": true,
        "description": "Query to retrieve person data"
      }
    }
  }
}`);
  const [status, setStatus] = useState<'idle' | 'submitting' | 'success' | 'error'>('idle');
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setError(null);
    setStatus('submitting');
    
    if (!providerId) {
      setError('No provider ID available. Please register as a provider and get admin approval first');
      setStatus('error');
      return;
    }
    
    try {
      // Parse the JSON to validate it
      const parsedPayload = JSON.parse(jsonPayload);
      
      // Convert string "yes"/"no" values to boolean for isOwner fields
      const convertedPayload = convertIsOwnerToBoolean(parsedPayload);
      
      logApiCall(`Calling POST /providers/${providerId}/schemas...`, 'Pending', convertedPayload);

      await submitProviderSchema(providerId, convertedPayload);
      setStatus('success');
      logApiCall(`POST /providers/${providerId}/schemas`, 201, { message: 'Schema submitted successfully' });
      onSuccess();
    } catch (err: any) {
      setError(err.message);
      setStatus('error');
      logApiCall(`POST /providers/${providerId}/schemas`, 'Error', { message: err.message });
    }
  };

  // Helper function to convert "yes"/"no" strings to boolean values for isOwner fields
  const convertIsOwnerToBoolean = (payload: any): any => {
    if (payload && typeof payload === 'object') {
      if (payload.fieldConfigurations) {
        const converted = { ...payload };
        converted.fieldConfigurations = { ...payload.fieldConfigurations };
        
        Object.keys(converted.fieldConfigurations).forEach(typeKey => {
          converted.fieldConfigurations[typeKey] = { ...payload.fieldConfigurations[typeKey] };
          
          Object.keys(converted.fieldConfigurations[typeKey]).forEach(fieldKey => {
            const field = converted.fieldConfigurations[typeKey][fieldKey];
            if (field && typeof field.isOwner === 'string') {
              field.isOwner = field.isOwner === 'yes';
            }
          });
        });
        
        return converted;
      }
    }
    return payload;
  };

  if (status === 'submitting') {
    return <LoadingSpinner text="Submitting schema..." />;
  }

  if (status === 'success') {
    return <SchemaSubmittedMessage />;
  }

  return (
    <div className="space-y-4 p-6 bg-green-50 rounded-lg border border-green-200">
      <h3 className="text-xl font-medium">Submit Schema</h3>
      {providerId ? (
        <p className="text-gray-600">Provider ID: <code className="font-mono bg-gray-200 px-1 py-0.5 rounded text-sm">{providerId}</code></p>
      ) : (
        <div className="p-3 bg-yellow-100 border border-yellow-300 rounded-md">
          <p className="text-yellow-800 text-sm">
            <strong>Warning:</strong> No provider ID available. Please register as a provider and get admin approval first
          </p>
        </div>
      )}
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Schema JSON Payload
          </label>
          <textarea
            value={jsonPayload}
            onChange={(e) => setJsonPayload(e.target.value)}
            className="w-full h-96 p-3 border border-gray-300 rounded-md font-mono text-sm"
            placeholder="Enter your schema JSON payload here..."
          />
          <p className="text-xs text-gray-500 mt-1">
            Enter the complete JSON payload for the schema submission. The JSON will be validated before submission
          </p>
        </div>

        {status === 'error' && <ErrorMessage message={error!} />}

        <button
          type="submit"
          disabled={status !== 'idle' || !providerId}
          className={`w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors ${
            status !== 'idle' || !providerId ? 'bg-gray-400 cursor-not-allowed' : 'bg-green-600 hover:bg-green-700'
          }`}
        >
          {!providerId ? 'No Provider ID' : status !== 'idle' ? 'Submitting...' : 'Submit Schema'}
        </button>
      </form>
    </div>
  );
}
