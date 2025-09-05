// components/SchemaExplorer.tsx
import React, { useState } from 'react';
import type { IntrospectionResult, FieldConfiguration as FieldConfig } from '../types/graphql';
import { FieldConfiguration } from './FieldConfiguration';

interface SchemaExplorerProps {
  schema: IntrospectionResult;
  configurations: Record<string, Record<string, FieldConfig>>;
  onConfigurationChange: (typeName: string, fieldName: string, config: FieldConfig) => void;
  onSubmit: () => void;
  loading: boolean;
}

export const SchemaExplorer: React.FC<SchemaExplorerProps> = ({
  schema,
  configurations,
  onConfigurationChange,
  onSubmit,
  loading
}) => {
  const [expandedTypes, setExpandedTypes] = useState<Set<string>>(new Set(['Query']));

  const toggleTypeExpansion = (typeName: string) => {
    const newExpanded = new Set(expandedTypes);
    if (newExpanded.has(typeName)) {
      newExpanded.delete(typeName);
    } else {
      newExpanded.add(typeName);
    }
    setExpandedTypes(newExpanded);
  };

  const getUserDefinedTypes = () => {
    return schema.data.__schema.types.filter(type => 
      !type.name.startsWith('__') && // Remove introspection types
      type.kind === 'OBJECT' &&
      type.fields &&
      type.fields.length > 0
    );
  };

  const isFormValid = () => {
    const userTypes = getUserDefinedTypes();
    
    for (const type of userTypes) {
      if (!type.fields) continue;
      
      for (const field of type.fields) {
        const config = configurations[type.name]?.[field.name];
        if (!config || !config.source || config.isOwner === null) {
          return false;
        }
      }
    }
    return true;
  };

  const getFieldCount = () => {
    return getUserDefinedTypes().reduce((total, type) => 
      total + (type.fields?.length || 0), 0
    );
  };

  const getConfiguredCount = () => {
    let configured = 0;
    getUserDefinedTypes().forEach(type => {
      if (!type.fields) return;
      type.fields.forEach(field => {
        const config = configurations[type.name]?.[field.name];
        if (config?.source) {
          configured++;
        }
      });
    });
    return configured;
  };

  const userTypes = getUserDefinedTypes();
  const totalFields = getFieldCount();
  const configuredFields = getConfiguredCount();

  return (
    <div className="bg-white p-6 rounded-lg shadow-md">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-800">Configure Schema Fields</h2>
        <div className="text-sm text-gray-600">
          Progress: {configuredFields}/{totalFields} fields configured
        </div>
      </div>

      <div className="mb-4 p-3 bg-blue-50 rounded-lg">
        <p className="text-sm text-blue-800">
          Configure each field by selecting its source type, ownership status, and providing a description. 
          All fields marked with <span className="text-red-500 font-semibold">*</span> are required.
        </p>
      </div>

      <div className="space-y-4 max-h-96 overflow-y-auto border border-gray-200 rounded-lg p-4">
        {userTypes.map((type) => (
          <div key={type.name} className="border-b border-gray-100 pb-4 last:border-b-0">
            <button
              type="button"
              onClick={() => toggleTypeExpansion(type.name)}
              className="w-full flex items-center justify-between p-2 text-left bg-gray-50 hover:bg-gray-100 rounded-md transition-colors"
            >
              <span className="font-semibold text-lg text-gray-800">
                {type.name}
                {type.fields && (
                  <span className="text-sm font-normal text-gray-500 ml-2">
                    ({type.fields.length} fields)
                  </span>
                )}
              </span>
              <span className="text-gray-400">
                {expandedTypes.has(type.name) ? '−' : '+'}
              </span>
            </button>

            {type.description && expandedTypes.has(type.name) && (
              <p className="mt-2 text-sm text-gray-600 italic px-2">
                {type.description}
              </p>
            )}

            {expandedTypes.has(type.name) && type.fields && (
              <div className="mt-4 space-y-3 pl-4">
                {type.fields.map((field) => {
                  const config = configurations[type.name]?.[field.name] || {
                    source: '' as any,
                    isowner: '' as any,
                    description: ''
                  };

                  return (
                    <FieldConfiguration
                      key={field.name}
                      typeName={type.name}
                      field={field}
                      configuration={config}
                      onChange={onConfigurationChange}
                    />
                  );
                })}
              </div>
            )}
          </div>
        ))}
      </div>

      <div className="mt-6 flex items-center justify-between">
        <div className="text-sm text-gray-600">
          {configuredFields < totalFields && (
            <span className="text-amber-600">
              Please configure all {totalFields - configuredFields} remaining fields to continue
            </span>
          )}
          {configuredFields === totalFields && totalFields > 0 && (
            <span className="text-green-600">
              All fields configured! Ready to submit registration.
            </span>
          )}
        </div>

        <button
          type="button"
          onClick={onSubmit}
          disabled={!isFormValid() || loading}
          className="bg-green-500 text-white py-2 px-6 rounded-md hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {loading ? 'Submitting...' : 'Submit Registration'}
        </button>
      </div>
    </div>
  );
};