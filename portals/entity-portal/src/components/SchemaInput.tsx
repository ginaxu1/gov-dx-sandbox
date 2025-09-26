// components/SchemaInput.tsx
import { useState } from 'react';
import type { IntrospectionResult } from '../types/graphql';
import { SchemaService } from '../services/schemaService';

interface SchemaInputProps {
  onSchemaLoaded: (schema: IntrospectionResult, endpoint: string) => void;
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

      onSchemaLoaded(schema, endpoint);
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

  const clearMemory = () => {
    setEndpoint('');
    setSdlContent('');
    setJsonFile(null);
  }

  return (
    <div className="bg-white p-6 rounded-xl shadow-lg border border-gray-200">
      <h2 className="text-2xl font-bold mb-6 text-gray-900">Provide GraphQL Schema</h2>
      
      <div className="mb-6">
        <div className="flex flex-col sm:flex-row gap-2 mb-6 bg-gray-50 p-2 rounded-lg">
          <button
            type="button"
            onClick={() => {setInputMethod('endpoint'); clearMemory();}}
            className={`flex-1 px-4 py-3 rounded-lg font-medium transition-all duration-200 ${
              inputMethod === 'endpoint'
                ? 'bg-blue-600 text-white shadow-sm'
                : 'bg-white text-gray-700 hover:bg-gray-100 border border-gray-200'
            }`}
          >
            <div className="flex items-center justify-center space-x-2">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
              <span>GraphQL Endpoint</span>
            </div>
          </button>
          <button
            type="button"
            onClick={() => {setInputMethod('json'); clearMemory();}}
            className={`flex-1 px-4 py-3 rounded-lg font-medium transition-all duration-200 ${
              inputMethod === 'json'
                ? 'bg-blue-600 text-white shadow-sm'
                : 'bg-white text-gray-700 hover:bg-gray-100 border border-gray-200'
            }`}
          >
            <div className="flex items-center justify-center space-x-2">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <span>JSON File</span>
            </div>
          </button>
          <button
            type="button"
            onClick={() => {setInputMethod('sdl'); clearMemory();}}
            className={`flex-1 px-4 py-3 rounded-lg font-medium transition-all duration-200 ${
              inputMethod === 'sdl'
                ? 'bg-blue-600 text-white shadow-sm'
                : 'bg-white text-gray-700 hover:bg-gray-100 border border-gray-200'
            }`}
          >
            <div className="flex items-center justify-center space-x-2">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
              </svg>
              <span>SDL</span>
            </div>
          </button>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {inputMethod === 'endpoint' && (
          <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
            <label htmlFor="endpoint" className="block text-sm font-semibold text-gray-800 mb-3">
              GraphQL Endpoint URL (with introspection enabled)<span className="text-red-600">*</span>
            </label>
            <input
              type="url"
              id="endpoint"
              value={endpoint}
              onChange={(e) => setEndpoint(e.target.value)}
              placeholder="https://api.example.com/graphql"
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
              required
              disabled={loading}
            />
          </div>
        )}

        {inputMethod === 'json' && (
          <div className="space-y-4">
            <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
              <label htmlFor="jsonFile" className="block text-sm font-semibold text-gray-800 mb-3">
                Upload Introspection Result JSON<span className="text-red-600">*</span>
              </label>
              <input
                type="file"
                id="jsonFile"
                accept=".json,application/json"
                onChange={handleFileChange}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-medium file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
                required
                disabled={loading}
              />
              {jsonFile && (
                <div className="mt-3 p-3 bg-green-50 rounded-lg border border-green-200">
                  <div className="flex items-center space-x-2">
                    <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <p className="text-sm text-green-800 font-medium">
                      Selected file: {jsonFile.name}
                    </p>
                  </div>
                </div>
              )}
            </div>
            <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
              <label htmlFor="endpoint" className="block text-sm font-semibold text-gray-800 mb-3">
                GraphQL Endpoint URL<span className="text-red-600">*</span>
              </label>
              <input
                type="url"
                id="endpoint"
                value={endpoint}
                onChange={(e) => setEndpoint(e.target.value)}
                placeholder="https://api.example.com/graphql"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
                required
                disabled={loading}
              />
            </div>
          </div>
        )}

        {inputMethod === 'sdl' && (
          <div className="space-y-4">
            <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
              <label htmlFor="sdl" className="block text-sm font-semibold text-gray-800 mb-3">
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
                rows={12}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm transition-colors resize-vertical"
                required
                disabled={loading}
              />
            </div>
            <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
              <label htmlFor="endpoint" className="block text-sm font-semibold text-gray-800 mb-3">
                GraphQL Endpoint URL<span className="text-red-600">*</span>
              </label>
              <input
                type="url"
                id="endpoint"
                value={endpoint}
                onChange={(e) => setEndpoint(e.target.value)}
                placeholder="https://api.example.com/graphql"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
                required
                disabled={loading}
              />
            </div>
          </div>
        )}

        <button
          type="submit"
          disabled={loading || !endpoint || (inputMethod==="json" && !jsonFile) || (inputMethod==="sdl" && !sdlContent)}
          className="w-full bg-blue-600 text-white py-3 px-6 rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 font-medium shadow-sm disabled:shadow-none"
        >
          {loading ? (
            <div className="flex items-center justify-center space-x-2">
              <svg className="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>Fetching Schema...</span>
            </div>
          ) : (
            <div className="flex items-center justify-center space-x-2">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              <span>Fetch Schema</span>
            </div>
          )}
        </button>
      </form>
    </div>
  );
};