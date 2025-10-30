import React, { useEffect, useState } from "react";
import { useNavigate } from 'react-router-dom';
import { Database, Plus, Clock, CheckCircle, Search, Code, AlertTriangle, AlertCircle } from 'lucide-react';
import { SchemaService } from '../services/schemaService';
import type { ApprovedSchema, SchemaSubmission } from '../types/graphql';

interface SchemasPageProps {
    memberId: string;
}

export const SchemasPage: React.FC<SchemasPageProps> = ({ memberId }) => {
    const navigate = useNavigate();
    const [registeredSchemas, setRegisteredSchemas] = useState<ApprovedSchema[]>([]);
    const [pendingSchemas, setPendingSchemas] = useState<SchemaSubmission[]>([]);
    const [searchTerm, setSearchTerm] = useState('');
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [usingMockData, setUsingMockData] = useState(false);

    useEffect(() => {
        const fetchSchemas = async () => {
            try {
                setLoading(true);
                setError(null);

                console.log('Fetching schemas for provider:', memberId);
                
                // Try to fetch real data from the API
                const [approvedSchemas, schemaSubmissions] = await Promise.all([
                    SchemaService.getApprovedSchemas(memberId),
                    SchemaService.getSchemaSubmissions(memberId)
                ]);

                console.log('Fetched approved schemas:', approvedSchemas);
                console.log('Fetched schema submissions:', schemaSubmissions);

                setRegisteredSchemas(approvedSchemas);
                setPendingSchemas(schemaSubmissions);
                setUsingMockData(false);
            } catch (error) {
                console.error('Error fetching schemas:', error);
                setError(error instanceof Error ? error.message : 'Failed to fetch schemas');
                
                console.log('Using mock data as fallback');
                // Use mock data as fallback
                const mockApprovedSchemas: ApprovedSchema[] = [];

                const mockPendingSchemas: SchemaSubmission[] = [];

                setRegisteredSchemas(mockApprovedSchemas);
                setPendingSchemas(mockPendingSchemas);
                setUsingMockData(true);
            } finally {
                setLoading(false);
            }
        };

        fetchSchemas();
    }, [memberId]);

    const handleCreateNewSchema = () => {
        navigate('/schemas/new');
    };

    const displayName = (schema: ApprovedSchema | SchemaSubmission) => {
        return schema.schemaName || schema.schemaEndpoint;
    };

    // Separate active and deprecated schemas
    const activeSchemas = registeredSchemas.filter(schema => schema.version === 'active');
    const deprecatedSchemas = registeredSchemas.filter(schema => schema.version === 'deprecated');

    if (loading) {
        return (
            <div className="min-h-screen bg-gradient-to-br from-purple-50 to-indigo-100">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                    <div className="animate-pulse">
                        <div className="h-8 bg-gray-300 rounded-md w-1/4 mb-4"></div>
                        <div className="h-4 bg-gray-300 rounded-md w-1/2 mb-8"></div>
                        <div className="bg-white rounded-xl shadow-lg p-6 mb-6">
                            <div className="h-6 bg-gray-300 rounded-md w-1/3 mb-4"></div>
                            <div className="space-y-3">
                                <div className="h-4 bg-gray-200 rounded-md"></div>
                                <div className="h-4 bg-gray-200 rounded-md"></div>
                                <div className="h-4 bg-gray-200 rounded-md"></div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gradient-to-br from-purple-50 to-indigo-100">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                {/* Error/Mock Data Banner */}
                {usingMockData && (
                    <div className="mb-6 bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                        <div className="flex items-center">
                            <AlertTriangle className="w-5 h-5 text-yellow-600 mr-2" />
                            <div>
                                <p className="text-yellow-800 font-medium">Using Mock Data</p>
                                <p className="text-yellow-700 text-sm">
                                    Unable to connect to the API. Displaying sample data for demonstration.
                                    {error && ` Error: ${error}`}
                                </p>
                            </div>
                        </div>
                    </div>
                )}

                {/* Header Section */}
                <div className="mb-8">
                    <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between">
                        <div className="mb-6 lg:mb-0">
                            <h1 className="text-4xl font-bold text-gray-900 mb-2">
                                GraphQL Schemas
                            </h1>
                            <p className="text-lg text-gray-600">
                                Manage your GraphQL schemas and track registration status
                            </p>
                        </div>
                        <div className="flex flex-col sm:flex-row gap-4">
                            <div className="relative">
                                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                                <input
                                    type="text"
                                    placeholder="Search schemas..."
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    className="pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 w-full sm:w-64"
                                />
                            </div>
                            <button
                                onClick={handleCreateNewSchema}
                                className="bg-gradient-to-r from-purple-600 to-purple-700 text-white px-6 py-2 rounded-lg hover:from-purple-700 hover:to-purple-800 transition-all duration-200 font-medium shadow-lg hover:shadow-xl flex items-center space-x-2"
                            >
                                <Plus className="w-5 h-5" />
                                <span>New Schema</span>
                            </button>
                        </div>
                    </div>
                </div>

                {/* Statistics Cards */}
                <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center">
                            <div className="p-3 bg-green-100 rounded-full">
                                <CheckCircle className="w-6 h-6 text-green-600" />
                            </div>
                            <div className="ml-4">
                                <p className="text-sm font-medium text-gray-600">Active Schemas</p>
                                <p className="text-2xl font-bold text-gray-900">{activeSchemas.length}</p>
                            </div>
                        </div>
                    </div>
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center">
                            <div className="p-3 bg-orange-100 rounded-full">
                                <AlertCircle className="w-6 h-6 text-orange-600" />
                            </div>
                            <div className="ml-4">
                                <p className="text-sm font-medium text-gray-600">Deprecated</p>
                                <p className="text-2xl font-bold text-gray-900">{deprecatedSchemas.length}</p>
                            </div>
                        </div>
                    </div>
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center">
                            <div className="p-3 bg-yellow-100 rounded-full">
                                <Clock className="w-6 h-6 text-yellow-600" />
                            </div>
                            <div className="ml-4">
                                <p className="text-sm font-medium text-gray-600">Pending Review</p>
                                <p className="text-2xl font-bold text-gray-900">{pendingSchemas.length}</p>
                            </div>
                        </div>
                    </div>
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center">
                            <div className="p-3 bg-purple-100 rounded-full">
                                <Database className="w-6 h-6 text-purple-600" />
                            </div>
                            <div className="ml-4">
                                <p className="text-sm font-medium text-gray-600">Total Schemas</p>
                                <p className="text-2xl font-bold text-gray-900">{registeredSchemas.length + pendingSchemas.length}</p>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Active Schemas */}
                <div className="bg-white rounded-xl shadow-lg mb-8 overflow-hidden">
                    <div className="bg-gradient-to-r from-green-600 to-green-700 px-6 py-4">
                        <div className="flex items-center">
                            <CheckCircle className="w-6 h-6 text-white mr-3" />
                            <h2 className="text-xl font-semibold text-white">Active Schemas</h2>
                        </div>
                    </div>
                    <div className="p-6">
                        {activeSchemas.length === 0 ? (
                            <div className="text-center py-12">
                                <CheckCircle className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                                <p className="text-gray-500 text-lg">
                                    {searchTerm ? 'No active schemas match your search' : 'No active schemas yet'}
                                </p>
                                {!searchTerm && (
                                    <button
                                        onClick={handleCreateNewSchema}
                                        className="mt-4 text-purple-600 hover:text-purple-700 font-medium"
                                    >
                                        Register your first schema
                                    </button>
                                )}
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                                {activeSchemas.map(schema => (
                                    <div key={schema.schemaId} className="bg-green-50 rounded-lg p-4 hover:bg-green-100 transition-colors border border-green-200">
                                        <div className="flex items-start justify-between">
                                            <div className="flex-1">
                                                <div className="flex items-center mb-2">
                                                    <Code className="w-5 h-5 text-green-600 mr-2" />
                                                    <h3 className="font-semibold text-gray-900">{displayName(schema)}</h3>
                                                </div>
                                                <div className="flex items-center text-sm mb-2">
                                                    <CheckCircle className="w-4 h-4 mr-1 text-green-600" />
                                                    <span className="text-green-600 font-medium">
                                                        {schema.version}
                                                    </span>
                                                </div>
                                                <p className="text-xs text-gray-500">
                                                    Created: {new Date(schema.createdAt).toLocaleDateString()}
                                                </p>
                                            </div>
                                            <button className="text-gray-400 hover:text-gray-600">
                                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                </div>

                {/* Deprecated Schemas */}
                {deprecatedSchemas.length > 0 && (
                    <div className="bg-white rounded-xl shadow-lg mb-8 overflow-hidden">
                        <div className="bg-gradient-to-r from-orange-600 to-orange-700 px-6 py-4">
                            <div className="flex items-center">
                                <AlertCircle className="w-6 h-6 text-white mr-3" />
                                <h2 className="text-xl font-semibold text-white">Deprecated Schemas</h2>
                            </div>
                        </div>
                        <div className="p-6">
                            <div className="mb-4 p-3 bg-orange-50 border border-orange-200 rounded-lg">
                                <p className="text-orange-800 text-sm">
                                    <AlertCircle className="w-4 h-4 inline mr-1" />
                                    These schemas are deprecated and should be migrated to newer versions.
                                </p>
                            </div>
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                                {deprecatedSchemas.map(schema => (
                                    <div key={schema.schemaId} className="bg-orange-50 rounded-lg p-4 hover:bg-orange-100 transition-colors border border-orange-200">
                                        <div className="flex items-start justify-between">
                                            <div className="flex-1">
                                                <div className="flex items-center mb-2">
                                                    <Code className="w-5 h-5 text-orange-600 mr-2" />
                                                    <h3 className="font-semibold text-gray-900">{displayName(schema)}</h3>
                                                </div>
                                                <div className="flex items-center text-sm mb-2">
                                                    <AlertCircle className="w-4 h-4 mr-1 text-orange-600" />
                                                    <span className="text-orange-600 font-medium">
                                                        {schema.version}
                                                    </span>
                                                </div>
                                                <p className="text-xs text-gray-500">
                                                    Created: {new Date(schema.createdAt).toLocaleDateString()}
                                                </p>
                                            </div>
                                            <button className="text-gray-400 hover:text-gray-600">
                                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                )}

                {/* Pending Schemas */}
                <div className="bg-white rounded-xl shadow-lg overflow-hidden">
                    <div className="bg-gradient-to-r from-yellow-600 to-yellow-700 px-6 py-4">
                        <div className="flex items-center">
                            <Clock className="w-6 h-6 text-white mr-3" />
                            <h2 className="text-xl font-semibold text-white">Pending Review</h2>
                        </div>
                    </div>
                    <div className="p-6">
                        {pendingSchemas.length === 0 ? (
                            <div className="text-center py-12">
                                <Clock className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                                <p className="text-gray-500 text-lg">
                                    {searchTerm ? 'No pending schemas match your search' : 'No pending schemas'}
                                </p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                                {pendingSchemas.map(schema => (
                                    <div key={schema.submissionId} className="bg-gray-50 rounded-lg p-4 hover:bg-gray-100 transition-colors border border-gray-200">
                                        <div className="flex items-start justify-between">
                                            <div className="flex-1">
                                                <div className="flex items-center mb-2">
                                                    <Code className="w-5 h-5 text-gray-500 mr-2" />
                                                    <h3 className="font-semibold text-gray-900">{displayName(schema)}</h3>
                                                </div>
                                                <div className="flex items-center text-sm mb-2">
                                                    <Clock className="w-4 h-4 mr-1" />
                                                    <span className={`capitalize ${
                                                        schema.status === 'pending' ? 'text-yellow-600' :
                                                        schema.status === 'approved' ? 'text-green-600' :
                                                        'text-red-600'
                                                    }`}>
                                                        {schema.status}
                                                    </span>
                                                </div>
                                                <p className="text-xs text-gray-500">
                                                    Submitted: {new Date(schema.createdAt).toLocaleDateString()}
                                                </p>
                                            </div>
                                            <button className="text-gray-400 hover:text-gray-600">
                                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};