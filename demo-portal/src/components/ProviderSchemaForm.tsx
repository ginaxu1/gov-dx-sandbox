import { useState, FormEvent } from 'react';
import { FieldConfiguration } from '../../../api-server/src/types';
import { submitProviderSchema } from '../services/api.service';
import { LoadingSpinner, ErrorMessage } from '../utils/form-helpers';

interface ProviderSchemaFormProps {
  providerId: string;
  logApiCall: (message: string, status: number | string, response: object) => void;
  onSuccess: () => void;
}

const SchemaSubmittedMessage = () => (
    <div className="text-center p-6 bg-green-100 text-green-800 rounded-lg">
        <h3 className="text-xl font-semibold">Schema Submitted!</h3>
        <p>Your schema has been submitted for review.</p>
        <p className="mt-2">Go to the **Admin** view to approve it.</p>
    </div>
);

export default function ProviderSchemaForm({ providerId, logApiCall, onSuccess }: ProviderSchemaFormProps) {
  const [schemaFields, setSchemaFields] = useState<Record<string, Record<string, FieldConfiguration>>>({
    PersonData: {
      nic: { source: 'fallback', isOwner: false, description: '' },
    },
  });
  const [status, setStatus] = useState<'idle' | 'submitting' | 'success' | 'error'>('idle');
  const [error, setError] = useState<string | null>(null);

  const handleFieldChange = (
    typeKey: string,
    fieldKey: string,
    property: keyof FieldConfiguration,
    value: string
  ) => {
    setSchemaFields((prevFields) => ({
      ...prevFields,
      [typeKey]: {
        ...prevFields[typeKey],
        [fieldKey]: {
          ...prevFields[typeKey][fieldKey],
          [property]: value,
        },
      },
    }));
  };

  const handleAddField = (typeKey: string, newFieldKey: string) => {
    if (!newFieldKey) return;
    setSchemaFields((prevFields) => ({
      ...prevFields,
      [typeKey]: {
        ...prevFields[typeKey],
        [newFieldKey]: {
          source: 'fallback',
          isOwner: false,
          description: '',
        },
      },
    }));
  };

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setError(null);
    setStatus('submitting');
    
    const payload = { fieldConfigurations: schemaFields };
    logApiCall(`Calling POST /providers/${providerId}/schemas...`, 'Pending', payload);

    try {
      await submitProviderSchema(providerId, payload);
      setStatus('success');
      logApiCall(`POST /providers/${providerId}/schemas`, 201, { message: 'Schema submitted successfully' });
      onSuccess();
    } catch (err: any) {
      setError(err.message);
      setStatus('error');
      logApiCall(`POST /providers/${providerId}/schemas`, 'Error', { message: err.message });
    }
  };

  if (status === 'submitting') {
    return <LoadingSpinner text="Submitting schema..." />;
  }

  if (status === 'success') {
    return <SchemaSubmittedMessage />;
  }

  return (
    <div className="space-y-4 p-6 bg-green-50 rounded-lg border border-green-200">
      <h3 className="text-xl font-medium">2. Submit Schema</h3>
      <p className="text-gray-600">Provider ID: <code className="font-mono bg-gray-200 px-1 py-0.5 rounded text-sm">{providerId}</code></p>
      <form onSubmit={handleSubmit} className="space-y-4">
        {Object.entries(schemaFields).map(([typeKey, typeFields]) => (
          <div key={typeKey} className="p-4 border border-green-300 rounded-md space-y-4">
            <h4 className="font-semibold text-green-700">Type: {typeKey}</h4>
            {Object.entries(typeFields).map(([fieldKey, fieldConfig]) => (
              <div key={fieldKey} className="space-y-2 p-3 bg-white rounded-md shadow-sm">
                <h5 className="font-medium text-gray-800">Field: {fieldKey}</h5>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Source</label>
                    <select
                      value={fieldConfig.source}
                      onChange={(e) => handleFieldChange(typeKey, fieldKey, 'source', e.target.value)}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm sm:text-sm p-2"
                    >
                      <option value="authoritative">Authoritative</option>
                      <option value="fallback">Fallback</option>
                      <option value="other">Other</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Is Owner</label>
                    <select
                      value={fieldConfig.isOwner.toString()}
                      onChange={(e) => handleFieldChange(typeKey, fieldKey, 'isOwner', e.target.value === 'true' ? 'true' : 'false')}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm sm:text-sm p-2"
                    >
                      <option value="true">Yes</option>
                      <option value="false">No</option>
                    </select>
                  </div>
                  <div className="md:col-span-1">
                    <label className="block text-sm font-medium text-gray-700">Description</label>
                    <input
                      type="text"
                      value={fieldConfig.description}
                      onChange={(e) => handleFieldChange(typeKey, fieldKey, 'description', e.target.value)}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm sm:text-sm p-2"
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        ))}

        {status === 'error' && <ErrorMessage message={error!} />}

        <button
          type="submit"
          disabled={status === 'submitting'}
          className={`w-full px-4 py-2 text-white font-semibold rounded-md shadow-md transition-colors ${
            status === 'submitting' ? 'bg-gray-400 cursor-not-allowed' : 'bg-green-600 hover:bg-green-700'
          }`}
        >
          {status === 'submitting' ? 'Submitting...' : 'Submit Schema'}
        </button>
      </form>
    </div>
  );
}
