// pages/RegistrationPage.tsx
import { useState } from 'react';
import type { IntrospectionResult, FieldConfiguration, SchemaRegistration, GraphQLType } from '../types/graphql';
import { SchemaInput } from '../components/SchemaInput';
import { SchemaExplorer } from '../components/SchemaExplorer';
import { SchemaService } from '../services/schemaService';

interface SchemaRegistrationPageProps {
  providerId: string;
  providerName: string;
}

export const SchemaRegistrationPage: React.FC<SchemaRegistrationPageProps> = ({
  providerId,
  providerName,
}) => {
  const [step, setStep] = useState<'input' | 'configure'>('input');
  const [schema, setSchema] = useState<IntrospectionResult | null>(null);
  const [configurations, setConfigurations] = useState<Record<string, Record<string, FieldConfiguration>>>({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [userDefinedTypes, setUserDefinedTypes] = useState<GraphQLType[]>([]);

  const handleSchemaLoaded = (loadedSchema: IntrospectionResult) => {
    setSchema(loadedSchema);
    setStep('configure');
    setError('');
    // Initialize configurations for all fields
    const initialConfigs: Record<string, Record<string, FieldConfiguration>> = {};

    const userDefinedTypes_ = SchemaService.getUserDefinedTypes(loadedSchema);
    setUserDefinedTypes(userDefinedTypes_);
    loadedSchema.data.__schema.types
      .filter(type => 
        !type.name.startsWith('__') && 
        type.kind === 'OBJECT' && 
        type.fields
      )
      .forEach(type => {
        initialConfigs[type.name] = {};
        if (type.name === "Query"){
          type.fields?.forEach(field => {
            initialConfigs[type.name][field.name] = {
              source: "",
              isOwner: null,
              description: field.description || '',
              isQueryType: true,
              isUserDefinedTypeField: false
            };
          });
        }
        else {
          type.fields?.forEach(field => {
            const isUserDefinedTypeField_ = userDefinedTypes_.map(t => t.name).includes(SchemaService.getTypeString(field.type));
            initialConfigs[type.name][field.name] = {
              source: '',
              isOwner: isUserDefinedTypeField_ ? (null): false,
              description: field.description || '',
              isQueryType: false,
              isUserDefinedTypeField: isUserDefinedTypeField_
            };   
          });
        }        
      });
    
    setConfigurations(initialConfigs);
  };

  const handleConfigurationChange = (
    typeName: string, 
    fieldName: string, 
    config: FieldConfiguration
  ) => {
    setConfigurations(prev => ({
      ...prev,
      [typeName]: {
        ...prev[typeName],
        [fieldName]: config
      }
    }));
  };

  const handleSubmitRegistration = async () => {
    console.log('Submitting registration...');
    if (!schema) {
      setError('Schema is required');
      return;
    }

    setLoading(true);
    setError('');
    setSuccess('');

    try {
      // Generate SDL with directives
      const sdl = await SchemaService.generateSDLWithDirectives(schema, configurations);
      console.log("Generated SDL with directives:");
      console.log(sdl);

      const registration: SchemaRegistration = {
        sdl
      };
      console.log('Registering schema:', registration);
      await SchemaService.registerSchema(providerId,registration);
      setSuccess('Schema registered successfully!');
      
      // Reset form after successful registration
      setTimeout(() => {
        setStep('input');
        setSchema(null);
        setConfigurations({});
        // setProviderId('');
        setSuccess('');
      }, 3000);
      
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  const handleBackToInput = () => {
    setStep('input');
    setSchema(null);
    setConfigurations({});
    setError('');
  };

  const getSchemaStats = () => {
    if (!schema) return null;
    
    const types = schema.data.__schema.types.filter(type => 
      !type.name.startsWith('__') && type.kind === 'OBJECT'
    );
    
    const totalFields = types.reduce((sum, type) => 
      sum + (type.fields?.length || 0), 0
    );

    return {
      types: types.length,
      fields: totalFields,
      queryType: schema.data.__schema.queryType?.name,
      mutationType: schema.data.__schema.mutationType?.name,
      subscriptionType: schema.data.__schema.subscriptionType?.name
    };
  };

  const stats = getSchemaStats();

  return (
    <div className="min-h-screen bg-gray-50 py-8 w-full">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            GraphQL Schema Registration
          </h1>
          <p className="text-gray-600">
            Register your GraphQL schema as a data provider
          </p>
        </div>

        {/* Progress Indicator */}
        <div className="mb-8">
          <div className="flex items-center justify-center space-x-4">
            <div className={`flex items-center space-x-2 ${step === 'input' ? 'text-blue-600' : 'text-gray-400'}`}>
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
                step === 'input' ? 'bg-blue-600 text-white' : 'bg-gray-200'
              }`}>
                1
              </div>
              <span className="font-medium">Schema Input</span>
            </div>
            
            <div className="w-16 h-px bg-gray-300"></div>
            
            <div className={`flex items-center space-x-2 ${step === 'configure' ? 'text-blue-600' : 'text-gray-400'}`}>
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
                step === 'configure' ? 'bg-blue-600 text-white' : 'bg-gray-200'
              }`}>
                2
              </div>
              <span className="font-medium">Configure Fields</span>
            </div>
          </div>
        </div>

        {/* Error Alert */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-md">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">Error</h3>
                <div className="mt-2 text-sm text-red-700">{error}</div>
              </div>
            </div>
          </div>
        )}

        {/* Success Alert */}
        {success && (
          <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-md">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-green-800">Success</h3>
                <div className="mt-2 text-sm text-green-700">{success}</div>
              </div>
            </div>
          </div>
        )}

        {/* Schema Stats */}
        {stats && step === 'configure' && (
          <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-md">
            <h3 className="text-sm font-medium text-blue-800 mb-2">Schema Information</h3>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div>
                <span className="font-medium text-blue-700">Types:</span>
                <span className="ml-2 text-blue-600">{stats.types}</span>
              </div>
              <div>
                <span className="font-medium text-blue-700">Fields:</span>
                <span className="ml-2 text-blue-600">{stats.fields}</span>
              </div>
              <div>
                <span className="font-medium text-blue-700">Query:</span>
                <span className="ml-2 text-blue-600">{stats.queryType || 'None'}</span>
              </div>
              <div>
                <span className="font-medium text-blue-700">Mutation:</span>
                <span className="ml-2 text-blue-600">{stats.mutationType || 'None'}</span>
              </div>
            </div>
          </div>
        )}

        {/* Step Content */}
        {step === 'input' && (
          <SchemaInput
            onSchemaLoaded={handleSchemaLoaded}
            onError={setError}
            loading={loading}
            setLoading={setLoading}
          />
        )}

        {step === 'configure' && schema && (
          <div className="space-y-6">
            {/* Provider ID Input */}
            <div className="bg-white p-6 rounded-lg shadow-md">
              <div>
                <label htmlFor="providerName" className="block text-sm font-medium text-gray-700 mb-2">
                  Provider Name <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  id="providerName"
                  value={providerName}
                  placeholder="Provider Name"
                  disabled
                  readOnly
                  className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-100 text-gray-600 cursor-not-allowed"
                />
              </div>
              <div className="mt-4">
                <label htmlFor="providerId" className="block text-sm font-medium text-gray-700 mb-2">
                  Provider ID <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  id="providerId"
                  value={providerId}
                  placeholder="Provider ID"
                  disabled
                  readOnly
                  className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-100 text-gray-600 cursor-not-allowed"
                />
              </div>

              <p className="mt-1 text-sm text-gray-500">
                This will be used to identify your data provider in the system
              </p>

            </div>

            {/* Back Button */}
            <div className="flex justify-start">
              <button
                type="button"
                onClick={handleBackToInput}
                className="text-blue-600 hover:text-blue-800 text-sm font-medium flex items-center"
              >
                <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
                Back to Schema Input
              </button>
            </div>

            {/* Schema Explorer */}
            <SchemaExplorer
              configurations={configurations}
              userDefinedTypes={userDefinedTypes}
              onConfigurationChange={handleConfigurationChange}
              onSubmit={handleSubmitRegistration}
              loading={loading}
            />
          </div>
        )}
      </div>
    </div>
  );
};