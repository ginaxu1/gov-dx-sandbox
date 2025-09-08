// components/SchemaInput.tsx
import React, { useState } from 'react';
import type { IntrospectionResult } from '../types/graphql';
import { SchemaService } from '../services/schemaService';

interface SchemaInputProps {
  onSchemaLoaded: (schema: IntrospectionResult) => void;
  onError: (error: string) => void;
  loading: boolean;
  setLoading: (loading: boolean) => void;
}

export const SchemaInput: React.FC<SchemaInputProps> = ({
  onSchemaLoaded,
  onError,
  loading,
  setLoading
}) => {
  const [inputMethod, setInputMethod] = useState<'endpoint' | 'json' | 'sdl'>('endpoint');
  const [endpoint, setEndpoint] = useState('');
  const [sdlContent, setSdlContent] = useState('');
  const [jsonFile, setJsonFile] = useState<File | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    onError('');

    try {
      let schema: IntrospectionResult;

      switch (inputMethod) {
        case 'endpoint':
          if (!endpoint.trim()) {
            throw new Error('Please enter a GraphQL endpoint');
          }
          schema = await SchemaService.fetchSchemaFromEndpoint(endpoint.trim());
          break;

        case 'json':
          if (!jsonFile) {
            throw new Error('Please select a JSON file');
          }
          schema = await SchemaService.parseIntrospectionJSON(jsonFile);
          break;

        case 'sdl':
          if (!sdlContent.trim()) {
            throw new Error('Please enter SDL content');
          }
          schema = await SchemaService.parseSDL(sdlContent.trim());
          break;

        default:
          throw new Error('Invalid input method');
      }

      onSchemaLoaded(schema);
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Unknown error occurred');
    } finally {
      setLoading(false);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setJsonFile(file);
    }
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow-md">
      <h2 className="text-2xl font-bold mb-6 text-gray-800">Provide GraphQL Schema</h2>
      
      <div className="mb-6">
        <div className="flex space-x-4 mb-4">
          <button
            type="button"
            onClick={() => setInputMethod('endpoint')}
            className={`px-4 py-2 rounded-md transition-colors ${
              inputMethod === 'endpoint'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            GraphQL Endpoint
          </button>
          <button
            type="button"
            onClick={() => setInputMethod('json')}
            className={`px-4 py-2 rounded-md transition-colors ${
              inputMethod === 'json'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            JSON File
          </button>
          <button
            type="button"
            onClick={() => setInputMethod('sdl')}
            className={`px-4 py-2 rounded-md transition-colors ${
              inputMethod === 'sdl'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            SDL
          </button>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        {inputMethod === 'endpoint' && (
          <div>
            <label htmlFor="endpoint" className="block text-sm font-medium text-gray-700 mb-2">
              GraphQL Endpoint URL (with introspection enabled)
            </label>
            <input
              type="url"
              id="endpoint"
              value={endpoint}
              onChange={(e) => setEndpoint(e.target.value)}
              placeholder="https://api.example.com/graphql"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              required
              disabled={loading}
            />
          </div>
        )}

        {inputMethod === 'json' && (
          <div>
            <label htmlFor="jsonFile" className="block text-sm font-medium text-gray-700 mb-2">
              Upload Introspection Result JSON
            </label>
            <input
              type="file"
              id="jsonFile"
              accept=".json,application/json"
              onChange={handleFileChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              required
              disabled={loading}
            />
            {jsonFile && (
              <p className="mt-2 text-sm text-gray-600">
                Selected file: {jsonFile.name}
              </p>
            )}
          </div>
        )}

        {inputMethod === 'sdl' && (
          <div>
            <label htmlFor="sdl" className="block text-sm font-medium text-gray-700 mb-2">
              GraphQL Schema Definition Language (SDL)
            </label>
            <textarea
              id="sdl"
              value={sdlContent}
              onChange={(e) => setSdlContent(e.target.value)}
              placeholder="type Query {
  hello: String
}

type User {
  id: ID!
  name: String!
  email: String!
}"
              rows={10}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
              required
              disabled={loading}
            />
          </div>
        )}

        <button
          type="submit"
          disabled={loading}
          className="w-full bg-blue-500 text-white py-2 px-4 rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {loading ? 'Fetching Schema...' : 'Fetch Schema'}
        </button>
      </form>
    </div>
  );
};