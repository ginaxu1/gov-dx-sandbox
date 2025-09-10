// components/FieldConfiguration.tsx
import type { GraphQLField } from '../types/graphql';
import type { FieldConfiguration as FieldConfig } from '../types/graphql';
import { SchemaService } from '../services/schemaService';

interface FieldConfigurationProps {
  typeName: string;
  field: GraphQLField;
  configuration: FieldConfig;
  onChange: (typeName: string, fieldName: string, config: FieldConfig) => void;
}

export const FieldConfiguration: React.FC<FieldConfigurationProps> = ({
  typeName,
  field,
  configuration,
  onChange
}) => {
  const handleSourceChange = (source: 'authoritative' | 'fallback' | 'other') => {
    onChange(typeName, field.name, { ...configuration, source });
  };

  const handleIsOwnerChange = (isOwner: true | false) => {
    onChange(typeName, field.name, { ...configuration, isOwner });
  };

  const handleDescriptionChange = (description: string) => {
    onChange(typeName, field.name, { ...configuration, description });
  };

  return (
    <div className="border-l-4 border-blue-200 pl-4 mb-4">
      <div className="flex items-start gap-4">
        <div className="flex-1">
          <div className="font-medium text-gray-800 mb-1">
            <span className="text-blue-600">{field.name}</span>
            <span className="text-gray-500 ml-2">: {SchemaService.getTypeString(field.type)}</span>
          </div>
          
          {field.description && (
            <p className="text-sm text-gray-600 mb-2">{field.description}</p>
          )}

          {field.args && field.args.length > 0 && (
            <div className="text-xs text-gray-500 mb-2">
              Args: {field.args.map(arg => `${arg.name}: ${SchemaService.getTypeString(arg.type)}`).join(', ')}
            </div>
          )}
        </div>

        <div className="flex-1 space-y-3">
          {/* Source Configuration */}
          {(!configuration.isQueryType && !configuration.isUserDefinedTypeField) && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Source <span className="text-red-500">*</span>
            </label>
            <div className="flex space-x-4">
              {(['authoritative', 'fallback', 'other'] as const).map((option) => (
                <label key={option} className="flex items-center">
                  <input
                    type="radio"
                    name={`source-${typeName}-${field.name}`}
                    value={option}
                    checked={configuration.source === option}
                    onChange={() => handleSourceChange(option)}
                    className="mr-1 text-blue-600 focus:ring-blue-500"
                    required
                  />
                  <span className="text-sm capitalize">{option}</span>
                </label>
              ))}
            </div>
          </div>
          )}

          {/* Is Owner Configuration */}
          {(!configuration.isQueryType && !configuration.isUserDefinedTypeField) && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                <span>Is Owner</span>
                <input
                    type="checkbox"
                  name={`isOwner-${typeName}-${field.name}`}
                  checked={configuration.isOwner === true}
                  onChange={(e) => handleIsOwnerChange(e.target.checked)}
                  className="ml-1 text-blue-600 focus:ring-blue-500"
              />
            </label>
          </div>
        )}

          {/* Description */}
          <div>
            <label 
              htmlFor={`desc-${typeName}-${field.name}`}
              className="block text-sm font-medium text-gray-700 mb-1"
            >
              {
                (configuration.isUserDefinedTypeField && !configuration.isQueryType) && (
                  <span className='text-blue-500'>
                    This is an user-defined type field. You need to specify the source and isOwner for all the fields inside this user-defined type.<br />
                  </span>
                )
              }
              Description
              {(configuration.isQueryType || configuration.isUserDefinedTypeField) && <span className="text-red-500 ml-1">*</span>}
            </label>
            <textarea
              id={`desc-${typeName}-${field.name}`}
              value={configuration.description}
              onChange={(e) => handleDescriptionChange(e.target.value)}
              placeholder="Describe this field's purpose and data source..."
              rows={2}
              className="w-full text-sm px-2 py-1 border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
              required={configuration.isQueryType || configuration.isUserDefinedTypeField}
            />
            </div>
        </div>
      </div>
    </div>
  );
};