import React, { useState, useMemo } from 'react';
import { ChevronRight, ChevronDown, Info, HelpCircle } from 'lucide-react';
import { 
  buildSchema, 
  isObjectType,
  isListType,
  isNonNullType,
} from 'graphql';
import type { 
  GraphQLSchema, 
  GraphQLField, 
  GraphQLType,
  GraphQLArgument
} from 'graphql';
import type { SelectedField } from '../types/applications';

// Types for our schema representation
interface Field {
  name: string;
  type: string;
  isArray?: boolean;
  isRequired?: boolean;
  sourceInfo?: {
    providerKey: string;
    schemaId: string;
    providerField: string;
  };
  description?: string;
  args?: Argument[];
}

interface Argument {
  name: string;
  type: string;
  isRequired?: boolean;
}

interface Type {
  name: string;
  fields: Field[];
}

interface Schema {
  queries: Field[];
  types: { [key: string]: Type };
}

// Helper function to extract type info from GraphQL type
const getTypeInfo = (type: GraphQLType): { typeName: string; isArray: boolean; isRequired: boolean } => {
  let isRequired = false;
  let isArray = false;
  let currentType = type;

  // Handle NonNull wrapper
  if (isNonNullType(currentType)) {
    isRequired = true;
    currentType = currentType.ofType;
  }

  // Handle List wrapper
  if (isListType(currentType)) {
    isArray = true;
    currentType = currentType.ofType;
    
    // Handle NonNull inside List
    if (isNonNullType(currentType)) {
      currentType = currentType.ofType;
    }
  }

  return {
    typeName: (currentType as any).name || 'Unknown',
    isArray,
    isRequired
  };
};

// Helper function to extract source info from directives
const extractSourceInfo = (field: GraphQLField<any, any>) => {
  if (!field.astNode?.directives) return undefined;

  const sourceInfoDirective = field.astNode.directives.find(
    directive => directive.name.value === 'sourceInfo'
  );

  if (!sourceInfoDirective?.arguments) return undefined;

  const sourceInfo: { providerKey?: string; schemaId?: string; providerField?: string } = {};

  sourceInfoDirective.arguments.forEach(arg => {
    if (arg.value.kind === 'StringValue') {
      switch (arg.name.value) {
        case 'providerKey':
          sourceInfo.providerKey = arg.value.value;
          break;
        case 'schemaId':
          sourceInfo.schemaId = arg.value.value;
          break;
        case 'providerField':
          sourceInfo.providerField = arg.value.value;
          break;
      }
    }
  });

  if (sourceInfo.providerKey && sourceInfo.schemaId && sourceInfo.providerField) {
    return sourceInfo as Required<typeof sourceInfo>;
  }

  return undefined;
};

// Helper function to convert GraphQL arguments to our format
const convertArguments = (args: readonly GraphQLArgument[]): Argument[] => {
  return args.map(arg => {
    const typeInfo = getTypeInfo(arg.type);
    return {
      name: arg.name,
      type: `${typeInfo.isArray ? '[' : ''}${typeInfo.typeName}${typeInfo.isArray ? ']' : ''}`,
      isRequired: typeInfo.isRequired
    };
  });
};

// Helper function to convert GraphQL field to our format
const convertField = (field: GraphQLField<any, any>): Field => {
  const typeInfo = getTypeInfo(field.type);
  const sourceInfo = extractSourceInfo(field);

  return {
    name: field.name,
    type: typeInfo.typeName,
    isArray: typeInfo.isArray,
    isRequired: typeInfo.isRequired,
    sourceInfo,
    description: field.description || undefined,
    args: field.args.length > 0 ? convertArguments(field.args) : undefined
  };
};

// Main parsing function using graphql.js
const parseSDL = (sdl: string): Schema => {
  try {
    const graphqlSchema: GraphQLSchema = buildSchema(sdl);
    const schema: Schema = { queries: [], types: {} };

    // Get Query type
    const queryType = graphqlSchema.getQueryType();
    if (queryType) {
      const queryFields = queryType.getFields();
      schema.queries = Object.values(queryFields).map(convertField);
    }

    // Get all custom types
    const typeMap = graphqlSchema.getTypeMap();
    Object.entries(typeMap).forEach(([typeName, type]) => {
      // Skip built-in scalar types and introspection types
      if (typeName.startsWith('__') || ['String', 'Int', 'Float', 'Boolean', 'ID'].includes(typeName)) {
        return;
      }

      if (isObjectType(type) && typeName !== 'Query') {
        const fields = type.getFields();
        schema.types[typeName] = {
          name: typeName,
          fields: Object.values(fields).map(convertField)
        };
      }
    });

    return schema;
  } catch (error) {
    console.error('Error parsing SDL:', error);
    // Return empty schema on error
    return { queries: [], types: {} };
  }
};

// Helper function to get all leaf paths for a given field
const getAllLeafPaths = (currentPath: string, currentField: Field, schema: Schema): string[] => {
  const customType = schema.types[currentField.type];
  if (!customType) {
    return [currentPath];
  }
  
  const leafPaths: string[] = [];
  customType.fields.forEach(childField => {
    leafPaths.push(...getAllLeafPaths(`${currentPath}.${childField.name}`, childField, schema));
  });
  return leafPaths;
};

// Component for rendering individual fields
interface FieldNodeProps {
  field: Field;
  path: string;
  schema: Schema;
  selectedFields: Set<string>;
  expandedNodes: Set<string>;
  onFieldToggle: (path: string, isSelected: boolean, field: Field) => void;
  onNodeToggle: (path: string) => void;
  level: number;
}

const FieldNode: React.FC<FieldNodeProps> = ({
  field,
  path,
  schema,
  selectedFields,
  expandedNodes,
  onFieldToggle,
  onNodeToggle,
  level
}) => {
  const isCustomType = schema.types[field.type];
  const isExpanded = expandedNodes.has(path);
  const hasChildren = !!isCustomType;
  
  // For parent nodes, check if all children are selected
  const isSelected = useMemo(() => {
    if (!hasChildren) {
      return selectedFields.has(path);
    }
    
    // For parent nodes, check if all leaf children are selected
    const allLeafPaths = getAllLeafPaths(path, field, schema);
    return allLeafPaths.length > 0 && allLeafPaths.every((leafPath: string) => selectedFields.has(leafPath));
  }, [selectedFields, path, field, hasChildren, schema]);
  
  const handleCheckboxChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onFieldToggle(path, e.target.checked, field);
  };
  
  const handleExpand = () => {
    if (hasChildren) {
      onNodeToggle(path);
    }
  };
  
  return (
    <div className="select-none">
      <div 
        className={`flex items-center py-2 px-3 rounded-lg hover:bg-white hover:shadow-sm transition-all duration-150 ${isSelected ? 'bg-blue-50 border border-blue-200' : 'border border-transparent'}`}
        style={{ marginLeft: `${level * 20}px` }}
      >
        <div className="flex items-center flex-1 min-w-0">
          {hasChildren && (
            <button 
              type="button"
              onClick={handleExpand}
              className="mr-2 p-1 hover:bg-gray-200 rounded-md transition-colors"
            >
              {isExpanded ? 
                <ChevronDown className="w-4 h-4 text-gray-600" /> : 
                <ChevronRight className="w-4 h-4 text-gray-600" />
              }
            </button>
          )}
          {!hasChildren && <div className="w-6" />}
          
          <input
            type="checkbox"
            checked={isSelected}
            onChange={handleCheckboxChange}
            className="mr-3 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
          />
          
          <div className="flex-1 min-w-0">
            <div className="flex items-center flex-wrap gap-2 mb-1">
              <span className="font-mono text-sm font-medium text-gray-900">
                {field.name}
                {field.args && field.args.length > 0 && (
                  <span className="text-gray-500 font-normal">
                    ({field.args.map(arg => 
                      `${arg.name}: ${arg.type}${arg.isRequired ? '!' : ''}`
                    ).join(', ')})
                  </span>
                )}
              </span>
              
              <span className="text-xs font-mono text-gray-500 bg-gray-100 px-2 py-0.5 rounded">
                {field.isArray && '['}
                {field.type}
                {field.isArray && ']'}
                {field.isRequired && '!'}
              </span>
            </div>
            
            <div className="flex items-center gap-2 flex-wrap">
              {field.sourceInfo && (
                <div className="flex items-center text-xs text-blue-700 bg-blue-100 px-2 py-1 rounded-md">
                  <Info className="w-3 h-3 mr-1 flex-shrink-0" />
                  <span className="font-medium">{field.sourceInfo.providerKey}</span>
                  <span className="text-blue-500 mx-1">•</span>
                  <span className="text-blue-600 font-mono text-xs">{field.sourceInfo.schemaId}</span>
                  <span className="text-blue-500 mx-1">→</span>
                  <span className="truncate max-w-32">{field.sourceInfo.providerField}</span>
                </div>
              )}
              
              {field.description && (
                <div className="relative group">
                  <div className="flex items-center text-xs text-green-700 bg-green-100 px-2 py-1 rounded-md cursor-help">
                    <HelpCircle className="w-3 h-3 mr-1 flex-shrink-0" />
                    <span className="truncate max-w-24">
                      {field.description.length > 15 
                        ? `${field.description.substring(0, 15)}...` 
                        : field.description
                      }
                    </span>
                  </div>
                  <div className="absolute bottom-full left-0 mb-2 px-3 py-2 bg-gray-900 text-white text-xs rounded-lg shadow-xl opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none z-20 min-w-72 max-w-sm">
                    <div className="whitespace-normal break-words leading-relaxed">
                      {field.description}
                    </div>
                    <div className="absolute top-full left-4 w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-gray-900"></div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
      
      {hasChildren && isExpanded && (
        <div className="mt-1 border-l-2 border-gray-200 ml-6">
          <div className="space-y-1">
            {schema.types[field.type].fields.map((childField) => (
              <FieldNode
                key={childField.name}
                field={childField}
                path={`${path}.${childField.name}`}
                schema={schema}
                selectedFields={selectedFields}
                expandedNodes={expandedNodes}
                onFieldToggle={onFieldToggle}
                onNodeToggle={onNodeToggle}
                level={level + 1}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

// Main component
interface GraphQLSchemaExplorerProps {
  sdl: string;
  onSelectionChange?: (selectedFields: SelectedField[]) => void;
}

export const GraphQLSchemaExplorer: React.FC<GraphQLSchemaExplorerProps> = ({
  sdl,
  onSelectionChange
}) => {
  const [selectedFields, setSelectedFields] = useState<Set<string>>(new Set());
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());
  
  const schema = useMemo(() => {
    return parseSDL(sdl);
  }, [sdl]);
  
  const handleFieldToggle = (path: string, isSelected: boolean, field: Field) => {
    const newSelectedFields = new Set(selectedFields);
    
    if (isSelected) {
      const customType = schema.types[field.type];
      if (customType) {
        customType.fields.forEach(childField => {
          addFieldAndChildren(`${path}.${childField.name}`, childField, newSelectedFields);
        });
      } else {
        newSelectedFields.add(path);
      }
    } else {
      removeFieldAndChildren(path, newSelectedFields);
    }
    
    setSelectedFields(newSelectedFields);
    
    // Convert to SelectedField format for callback
    const selectedFieldsArray: SelectedField[] = Array.from(newSelectedFields).map(fieldPath => {
      // Find the field to get its schemaId
      const fieldObj = findFieldByPath(fieldPath, schema);
      return {
        fieldName: fieldPath,
        schemaId: fieldObj?.sourceInfo?.schemaId || 'unknown-schema'
      };
    });
    
    onSelectionChange?.(selectedFieldsArray);
  };

  const findFieldByPath = (path: string, schema: Schema): Field | undefined => {
    const pathParts = path.split('.');
    const queryField = schema.queries.find(q => q.name === pathParts[0]);
    
    if (!queryField) return undefined;
    if (pathParts.length === 1) return queryField;
    
    let currentField = queryField;
    for (let i = 1; i < pathParts.length; i++) {
      const currentType = schema.types[currentField.type];
      if (!currentType) return undefined;
      
      const foundField = currentType.fields.find(f => f.name === pathParts[i]);
      if (!foundField) return undefined;
      currentField = foundField;
    }
    
    return currentField;
  };

  const addFieldAndChildren = (path: string, field: Field, fieldsSet: Set<string>) => {
    const customType = schema.types[field.type];
    if (!customType) {
      fieldsSet.add(path);
    } else {
      customType.fields.forEach(childField => {
        addFieldAndChildren(`${path}.${childField.name}`, childField, fieldsSet);
      });
    }
  };

  const removeFieldAndChildren = (path: string, fieldsSet: Set<string>) => {
    const fieldsToRemove = Array.from(fieldsSet).filter(field => 
      field === path || field.startsWith(`${path}.`)
    );
    fieldsToRemove.forEach(field => fieldsSet.delete(field));
  };

  const handleNodeToggle = (path: string) => {
    const newExpandedNodes = new Set(expandedNodes);
    if (newExpandedNodes.has(path)) {
      newExpandedNodes.delete(path);
    } else {
      newExpandedNodes.add(path);
    }
    setExpandedNodes(newExpandedNodes);
  };
  
  return (
    <div className="border border-gray-200 rounded-xl overflow-hidden bg-white shadow-sm">
      <div className="bg-gray-50 px-6 py-4 border-b border-gray-200">
        <h3 className="text-lg font-semibold text-gray-900 mb-1">Available Data Fields</h3>
        <p className="text-sm text-gray-600">
          Select the fields your application needs access to. Hover over descriptions for more details.
        </p>
      </div>
      
      <div className="p-4">
        <div className="border border-gray-200 rounded-lg bg-gray-50 p-4 max-h-96 overflow-y-auto">
          {schema.queries.length > 0 ? (
            <div className="space-y-1">
              {schema.queries.map((query) => (
                <FieldNode
                  key={query.name}
                  field={query}
                  path={query.name}
                  schema={schema}
                  selectedFields={selectedFields}
                  expandedNodes={expandedNodes}
                  onFieldToggle={handleFieldToggle}
                  onNodeToggle={handleNodeToggle}
                  level={0}
                />
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <div className="text-gray-400 text-sm">No fields found in schema</div>
            </div>
          )}
        </div>
        
        {selectedFields.size > 0 && (
          <div className="mt-6 bg-blue-50 border border-blue-200 rounded-lg p-4">
            <h4 className="text-sm font-semibold text-blue-900 mb-3 flex items-center">
              <div className="w-2 h-2 bg-blue-500 rounded-full mr-2"></div>
              Selected Fields ({selectedFields.size})
            </h4>
            <div className="bg-white rounded-md p-3 max-h-32 overflow-y-auto border border-blue-200">
              <div className="text-xs font-mono text-gray-700 space-y-1">
                {Array.from(selectedFields).sort().map((fieldPath) => {
                  const field = findFieldByPath(fieldPath, schema);
                  return (
                    <div key={fieldPath} className="py-1 px-2 bg-gray-50 rounded text-gray-800 border-l-2 border-blue-400 flex justify-between items-center">
                      <span>{fieldPath}</span>
                      {field?.sourceInfo?.schemaId && (
                        <span className="text-blue-600 bg-blue-100 px-1 py-0.5 rounded text-xs">
                          {field.sourceInfo.schemaId}
                        </span>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};