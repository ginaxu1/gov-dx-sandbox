import React, { useState, useEffect } from 'react';
import { 
    Database, 
    Search, 
    Download, 
    RefreshCw, 
    Eye,
    Edit3,
    Clock,
    CheckCircle,
    User,
    Calendar,
    FileCode,
    Layers
} from 'lucide-react';

interface Schema {
    schemaId: string;
    name: string;
    schemaType: 'data' | 'api' | 'event';
    version: string;
    description: string;
    submittedBy: string;
    submittedByEmail: string;
    status: 'submitted' | 'approved' | 'rejected';
    createdAt: string;
    updatedAt: string;
    approvedAt?: string;
    approvedBy?: string;
    fields?: number;
    usageCount?: number;
}

interface FilterOptions {
    schemaType?: 'all' | 'data' | 'api' | 'event';
    searchByName?: string;
    submittedBy?: string;
    startDate?: string;
    endDate?: string;
    version?: string;
}

interface SchemasProps {
}

export const Schemas: React.FC<SchemasProps> = () => {
    const [submissions, setSubmissions] = useState<Schema[]>([]);
    const [approved, setApproved] = useState<Schema[]>([]);
    const [filteredSubmissions, setFilteredSubmissions] = useState<Schema[]>([]);
    const [filteredApproved, setFilteredApproved] = useState<Schema[]>([]);
    
    // Separate filters for submissions and approved
    const [submissionFilters, setSubmissionFilters] = useState<FilterOptions>({
        schemaType: 'all',
        searchByName: '',
        submittedBy: '',
        startDate: '',
        endDate: '',
        version: ''
    });
    
    const [approvedFilters, setApprovedFilters] = useState<FilterOptions>({
        schemaType: 'all',
        searchByName: '',
        submittedBy: '',
        startDate: '',
        endDate: '',
        version: ''
    });
    
    const [loading, setLoading] = useState(true);
    const [autoRefresh, setAutoRefresh] = useState(false);

    // Helper function to update submission filters
    const updateSubmissionFilter = <K extends keyof FilterOptions>(key: K, value: FilterOptions[K]) => {
        setSubmissionFilters(prev => ({ ...prev, [key]: value }));
    };

    // Helper function to update approved filters
    const updateApprovedFilter = <K extends keyof FilterOptions>(key: K, value: FilterOptions[K]) => {
        setApprovedFilters(prev => ({ ...prev, [key]: value }));
    };

    // Helper function to clear submission filters
    const clearSubmissionFilters = () => {
        setSubmissionFilters({
            schemaType: 'all',
            searchByName: '',
            submittedBy: '',
            startDate: '',
            endDate: '',
            version: ''
        });
    };

    // Helper function to clear approved filters
    const clearApprovedFilters = () => {
        setApprovedFilters({
            schemaType: 'all',
            searchByName: '',
            submittedBy: '',
            startDate: '',
            endDate: '',
            version: ''
        });
    };

    // Mock API calls - replace with actual service calls later
    const fetchSubmissions = async () => {
        try {
            // Mock data for schema submissions
            const mockSubmissions: Schema[] = [
                {
                    schemaId: 'schema_001',
                    name: 'Patient Health Record',
                    schemaType: 'data',
                    version: '1.0.0',
                    description: 'Schema for patient health records in healthcare system',
                    submittedBy: 'Ministry of Health',
                    submittedByEmail: 'health@gov.lk',
                    status: 'submitted',
                    createdAt: '2025-10-12T11:30:00Z',
                    updatedAt: '2025-10-12T11:30:00Z',
                    fields: 25,
                    usageCount: 0
                },
                {
                    schemaId: 'schema_002',
                    name: 'Tax Submission API',
                    schemaType: 'api',
                    version: '2.1.0',
                    description: 'API schema for tax submission endpoints',
                    submittedBy: 'Tax Authority',
                    submittedByEmail: 'tax@gov.lk',
                    status: 'submitted',
                    createdAt: '2025-10-11T16:20:00Z',
                    updatedAt: '2025-10-11T16:20:00Z',
                    fields: 18,
                    usageCount: 0
                },
                {
                    schemaId: 'schema_003',
                    name: 'Vehicle Registration Event',
                    schemaType: 'event',
                    version: '1.2.0',
                    description: 'Event schema for vehicle registration notifications',
                    submittedBy: 'DMV',
                    submittedByEmail: 'dmv@gov.lk',
                    status: 'submitted',
                    createdAt: '2025-10-10T10:15:00Z',
                    updatedAt: '2025-10-10T10:15:00Z',
                    fields: 12,
                    usageCount: 0
                }
            ];
            return mockSubmissions;
        } catch (error) {
            console.error('Error fetching schema submissions:', error);
            return [];
        }
    };

    const fetchApproved = async () => {
        try {
            // Mock data for approved schemas
            const mockApproved: Schema[] = [
                {
                    schemaId: 'schema_101',
                    name: 'Identity Verification Schema',
                    schemaType: 'data',
                    version: '1.5.0',
                    description: 'Schema for identity verification data structure',
                    submittedBy: 'Registrar General',
                    submittedByEmail: 'rg@gov.lk',
                    status: 'approved',
                    createdAt: '2025-10-01T09:00:00Z',
                    updatedAt: '2025-10-05T15:30:00Z',
                    approvedAt: '2025-10-05T15:30:00Z',
                    approvedBy: 'admin@gov.lk',
                    fields: 30,
                    usageCount: 45
                },
                {
                    schemaId: 'schema_102',
                    name: 'Education Records API',
                    schemaType: 'api',
                    version: '3.0.0',
                    description: 'API schema for education records endpoints',
                    submittedBy: 'Ministry of Education',
                    submittedByEmail: 'edu@gov.lk',
                    status: 'approved',
                    createdAt: '2025-09-28T13:00:00Z',
                    updatedAt: '2025-10-03T17:45:00Z',
                    approvedAt: '2025-10-03T17:45:00Z',
                    approvedBy: 'admin@gov.lk',
                    fields: 22,
                    usageCount: 78
                },
                {
                    schemaId: 'schema_103',
                    name: 'Business Registration Event',
                    schemaType: 'event',
                    version: '2.0.0',
                    description: 'Event schema for business registration notifications',
                    submittedBy: 'Business Registry',
                    submittedByEmail: 'business@gov.lk',
                    status: 'approved',
                    createdAt: '2025-09-25T12:30:00Z',
                    updatedAt: '2025-09-30T11:15:00Z',
                    approvedAt: '2025-09-30T11:15:00Z',
                    approvedBy: 'admin@gov.lk',
                    fields: 16,
                    usageCount: 32
                }
            ];
            return mockApproved;
        } catch (error) {
            console.error('Error fetching approved schemas:', error);
            return [];
        }
    };

    const fetchSchemas = async () => {
        setLoading(true);
        try {
            const [submissionsData, approvedData] = await Promise.all([
                fetchSubmissions(),
                fetchApproved()
            ]);
            
            setSubmissions(submissionsData);
            setApproved(approvedData);
            setFilteredSubmissions(submissionsData);
            setFilteredApproved(approvedData);
        } catch (error) {
            console.error('Error fetching schemas:', error);
            setSubmissions([]);
            setApproved([]);
            setFilteredSubmissions([]);
            setFilteredApproved([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchSchemas();
    }, []);

    // Auto-refresh functionality
    useEffect(() => {
        if (autoRefresh) {
            const interval = setInterval(() => {
                fetchSchemas();
            }, 30000); // Refresh every 30 seconds

            return () => clearInterval(interval);
        }
    }, [autoRefresh]);

    // Filter submissions
    useEffect(() => {
        let filtered = submissions;
        
        if (submissionFilters.searchByName) {
            filtered = filtered.filter(schema =>
                schema.name.toLowerCase().includes(submissionFilters.searchByName!.toLowerCase()) ||
                schema.description.toLowerCase().includes(submissionFilters.searchByName!.toLowerCase())
            );
        }

        if (submissionFilters.schemaType && submissionFilters.schemaType !== 'all') {
            filtered = filtered.filter(schema => schema.schemaType === submissionFilters.schemaType);
        }

        if (submissionFilters.submittedBy) {
            filtered = filtered.filter(schema => 
                schema.submittedBy.toLowerCase().includes(submissionFilters.submittedBy!.toLowerCase())
            );
        }

        if (submissionFilters.version) {
            filtered = filtered.filter(schema => 
                schema.version.toLowerCase().includes(submissionFilters.version!.toLowerCase())
            );
        }

        if (submissionFilters.startDate) {
            filtered = filtered.filter(schema => 
                new Date(schema.createdAt) >= new Date(submissionFilters.startDate!)
            );
        }

        if (submissionFilters.endDate) {
            filtered = filtered.filter(schema => 
                new Date(schema.createdAt) <= new Date(submissionFilters.endDate!)
            );
        }

        setFilteredSubmissions(filtered);
    }, [submissions, submissionFilters]);

    // Filter approved schemas
    useEffect(() => {
        let filtered = approved;
        
        if (approvedFilters.searchByName) {
            filtered = filtered.filter(schema =>
                schema.name.toLowerCase().includes(approvedFilters.searchByName!.toLowerCase()) ||
                schema.description.toLowerCase().includes(approvedFilters.searchByName!.toLowerCase())
            );
        }

        if (approvedFilters.schemaType && approvedFilters.schemaType !== 'all') {
            filtered = filtered.filter(schema => schema.schemaType === approvedFilters.schemaType);
        }

        if (approvedFilters.submittedBy) {
            filtered = filtered.filter(schema => 
                schema.submittedBy.toLowerCase().includes(approvedFilters.submittedBy!.toLowerCase())
            );
        }

        if (approvedFilters.version) {
            filtered = filtered.filter(schema => 
                schema.version.toLowerCase().includes(approvedFilters.version!.toLowerCase())
            );
        }

        if (approvedFilters.startDate) {
            filtered = filtered.filter(schema => 
                new Date(schema.createdAt) >= new Date(approvedFilters.startDate!)
            );
        }

        if (approvedFilters.endDate) {
            filtered = filtered.filter(schema => 
                new Date(schema.createdAt) <= new Date(approvedFilters.endDate!)
            );
        }

        setFilteredApproved(filtered);
    }, [approved, approvedFilters]);

    const getSchemaTypeIcon = (type: string) => {
        switch (type) {
            case 'data':
                return <Database className="w-5 h-5 text-blue-500" />;
            case 'api':
                return <FileCode className="w-5 h-5 text-green-500" />;
            case 'event':
                return <Layers className="w-5 h-5 text-purple-500" />;
            default:
                return <Database className="w-5 h-5 text-gray-500" />;
        }
    };

    const getSchemaTypeBorderColor = (type: string) => {
        switch (type) {
            case 'data':
                return 'border-l-blue-500 bg-blue-50';
            case 'api':
                return 'border-l-green-500 bg-green-50';
            case 'event':
                return 'border-l-purple-500 bg-purple-50';
            default:
                return 'border-l-gray-500 bg-gray-50';
        }
    };

    const formatTimestamp = (timestamp: string) => {
        return new Date(timestamp).toLocaleString();
    };

    const schemaTypes = ['data', 'api', 'event'];

    const handleRefresh = () => {
        fetchSchemas();
    };

    const handleExportSubmissions = async () => {
        try {
            console.log('Exporting schema submissions with filters:', submissionFilters);
            // TODO: Implement actual export logic
        } catch (error) {
            console.error('Error exporting schema submissions:', error);
        }
    };

    const handleExportApproved = async () => {
        try {
            console.log('Exporting approved schemas with filters:', approvedFilters);
            // TODO: Implement actual export logic
        } catch (error) {
            console.error('Error exporting approved schemas:', error);
        }
    };

    const handleReview = (schema: Schema) => {
        console.log('Reviewing schema:', schema.schemaId, schema.name);
        // TODO: Implement review logic
    };

    const handleEdit = (schema: Schema) => {
        console.log('Editing schema:', schema.schemaId, schema.name);
        // TODO: Implement edit logic
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-gradient-to-br from-gray-50 to-slate-100">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                    <div className="animate-pulse">
                        <div className="h-8 bg-gray-300 rounded-md w-1/4 mb-4"></div>
                        <div className="h-4 bg-gray-300 rounded-md w-1/2 mb-8"></div>
                        <div className="space-y-4">
                            {[...Array(5)].map((_, i) => (
                                <div key={i} className="bg-white rounded-lg p-4 border-l-4 border-gray-300">
                                    <div className="h-4 bg-gray-200 rounded-md w-3/4 mb-2"></div>
                                    <div className="h-3 bg-gray-200 rounded-md w-1/2"></div>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gradient-to-br from-gray-50 to-slate-100">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                {/* Header */}
                <div className="mb-8">
                    <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between">
                        <div className="mb-6 lg:mb-0">
                            <h1 className="text-4xl font-bold text-gray-900 mb-2">
                                Schemas
                            </h1>
                            <p className="text-lg text-gray-600">
                                Manage schema submissions and approved schemas
                            </p>
                        </div>
                        <div className="flex flex-col sm:flex-row gap-4">
                            <button
                                onClick={handleRefresh}
                                className="flex items-center space-x-2 px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                            >
                                <RefreshCw className="w-4 h-4" />
                                <span>Refresh</span>
                            </button>
                            <button
                                onClick={() => setAutoRefresh(!autoRefresh)}
                                className={`flex items-center space-x-2 px-4 py-2 rounded-lg transition-colors ${
                                    autoRefresh 
                                        ? 'bg-green-100 text-green-700 border border-green-300' 
                                        : 'bg-white border border-gray-300 hover:bg-gray-50'
                                }`}
                            >
                                <Database className="w-4 h-4" />
                                <span>Auto Refresh</span>
                            </button>
                        </div>
                    </div>
                </div>

                {/* Schema Statistics */}
                <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Total Submissions</p>
                                <p className="text-2xl font-bold text-gray-900">{filteredSubmissions.length}</p>
                                <p className="text-xs text-gray-500">of {submissions.length} total</p>
                            </div>
                            <div className="p-3 rounded-full bg-orange-100">
                                <Clock className="w-5 h-5 text-orange-500" />
                            </div>
                        </div>
                    </div>
                    
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Approved Schemas</p>
                                <p className="text-2xl font-bold text-gray-900">{filteredApproved.length}</p>
                                <p className="text-xs text-gray-500">of {approved.length} total</p>
                            </div>
                            <div className="p-3 rounded-full bg-green-100">
                                <CheckCircle className="w-5 h-5 text-green-500" />
                            </div>
                        </div>
                    </div>
                    
                    {schemaTypes.map(type => {
                        const submissionCount = filteredSubmissions.filter(schema => schema.schemaType === type).length;
                        const approvedCount = filteredApproved.filter(schema => schema.schemaType === type).length;
                        const totalCount = submissionCount + approvedCount;
                        
                        return (
                            <div key={type} className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                                <div className="flex items-center justify-between">
                                    <div>
                                        <p className="text-sm font-medium text-gray-600 capitalize">{type} Schemas</p>
                                        <p className="text-2xl font-bold text-gray-900">{totalCount}</p>
                                        <p className="text-xs text-gray-500">{submissionCount} pending, {approvedCount} approved</p>
                                    </div>
                                    <div className="p-3 rounded-full bg-gray-100">
                                        {getSchemaTypeIcon(type)}
                                    </div>
                                </div>
                            </div>
                        );
                    })}
                </div>

                {/* Submissions Section */}
                <div className="mb-8">
                    <div className="bg-white rounded-xl shadow-lg overflow-hidden">
                        <div className="bg-gradient-to-r from-orange-600 to-orange-700 px-6 py-4">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center">
                                    <Clock className="w-6 h-6 text-white mr-3" />
                                    <h2 className="text-xl font-semibold text-white">
                                        Pending Schema Submissions ({filteredSubmissions.length})
                                    </h2>
                                </div>
                                <button 
                                    onClick={handleExportSubmissions}
                                    className="flex items-center space-x-2 px-4 py-2 bg-orange-800 text-white rounded-lg hover:bg-orange-900 transition-colors"
                                >
                                    <Download className="w-4 h-4" />
                                    <span>Export</span>
                                </button>
                            </div>
                        </div>

                        {/* Submission Filters */}
                        <div className="p-6 border-b">
                            <div className="space-y-4">
                                <div className="flex flex-col lg:flex-row gap-4">
                                    <div className="flex-1">
                                        <div className="relative">
                                            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                                            <input
                                                type="text"
                                                placeholder="Search schema submissions..."
                                                value={submissionFilters.searchByName || ''}
                                                onChange={(e) => updateSubmissionFilter('searchByName', e.target.value)}
                                                className="pl-10 pr-4 py-2 w-full border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-orange-500"
                                            />
                                        </div>
                                    </div>
                                    <div className="flex flex-col sm:flex-row gap-4">
                                        <select
                                            value={submissionFilters.schemaType || 'all'}
                                            onChange={(e) => updateSubmissionFilter('schemaType', e.target.value as 'all' | 'data' | 'api' | 'event')}
                                            className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-orange-500"
                                        >
                                            <option value="all">All Types</option>
                                            {schemaTypes.map(type => (
                                                <option key={type} value={type}>
                                                    {type.charAt(0).toUpperCase() + type.slice(1)}
                                                </option>
                                            ))}
                                        </select>
                                    </div>
                                </div>
                                
                                <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                                    <input
                                        type="text"
                                        placeholder="Filter by submitter"
                                        value={submissionFilters.submittedBy || ''}
                                        onChange={(e) => updateSubmissionFilter('submittedBy', e.target.value)}
                                        className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-orange-500"
                                    />
                                    <input
                                        type="text"
                                        placeholder="Filter by version"
                                        value={submissionFilters.version || ''}
                                        onChange={(e) => updateSubmissionFilter('version', e.target.value)}
                                        className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-orange-500"
                                    />
                                    <input
                                        type="date"
                                        placeholder="Start Date"
                                        value={submissionFilters.startDate || ''}
                                        onChange={(e) => updateSubmissionFilter('startDate', e.target.value)}
                                        className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-orange-500"
                                    />
                                    <input
                                        type="date"
                                        placeholder="End Date"
                                        value={submissionFilters.endDate || ''}
                                        onChange={(e) => updateSubmissionFilter('endDate', e.target.value)}
                                        className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-orange-500"
                                    />
                                </div>
                                
                                <div className="flex justify-end">
                                    <button
                                        onClick={clearSubmissionFilters}
                                        className="px-4 py-2 text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
                                    >
                                        Clear Filters
                                    </button>
                                </div>
                            </div>
                        </div>

                        <div className="max-h-96 overflow-y-auto">
                            {filteredSubmissions.length === 0 ? (
                                <div className="text-center py-12">
                                    <Clock className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                                    <p className="text-gray-500 text-lg">
                                        {submissionFilters.searchByName || submissionFilters.schemaType !== 'all' || submissionFilters.submittedBy || submissionFilters.version || submissionFilters.startDate || submissionFilters.endDate
                                            ? 'No schema submissions match your filters' 
                                            : 'No schema submissions available'
                                        }
                                    </p>
                                </div>
                            ) : (
                                <div className="divide-y divide-gray-200">
                                    {filteredSubmissions.map((schema) => (
                                        <div key={schema.schemaId} className={`p-4 border-l-4 hover:bg-gray-50 transition-colors ${getSchemaTypeBorderColor(schema.schemaType)}`}>
                                            <div className="flex items-start space-x-3">
                                                <div className="flex-shrink-0 mt-1">
                                                    {getSchemaTypeIcon(schema.schemaType)}
                                                </div>
                                                <div className="flex-1 min-w-0">
                                                    <div className="flex items-center justify-between mb-2">
                                                        <h3 className="text-lg font-semibold text-gray-900">
                                                            {schema.name} v{schema.version}
                                                        </h3>
                                                        <div className="flex items-center text-xs text-gray-500">
                                                            <Calendar className="w-3 h-3 mr-1" />
                                                            Submitted: {formatTimestamp(schema.createdAt)}
                                                        </div>
                                                    </div>
                                                    
                                                    <p className="text-sm text-gray-600 mb-3">{schema.description}</p>
                                                    
                                                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mb-3">
                                                        <div className="flex items-center text-sm text-gray-600">
                                                            <User className="w-4 h-4 mr-2 text-gray-400" />
                                                            <span>{schema.submittedBy}</span>
                                                        </div>
                                                        <div className="flex items-center text-sm text-gray-600">
                                                            <span className="font-medium">Fields:</span> 
                                                            <span className="ml-1">{schema.fields}</span>
                                                        </div>
                                                    </div>

                                                    <div className="flex items-center justify-between">
                                                        <div className="flex items-center gap-2">
                                                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                                                schema.schemaType === 'data' 
                                                                    ? 'bg-blue-100 text-blue-800' 
                                                                    : schema.schemaType === 'api'
                                                                    ? 'bg-green-100 text-green-800'
                                                                    : 'bg-purple-100 text-purple-800'
                                                            }`}>
                                                                {schema.schemaType.charAt(0).toUpperCase() + schema.schemaType.slice(1)}
                                                            </span>
                                                            
                                                            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-orange-100 text-orange-800">
                                                                Pending Review
                                                            </span>
                                                        </div>
                                                        
                                                        <button
                                                            onClick={() => handleReview(schema)}
                                                            className="flex items-center space-x-2 px-4 py-2 bg-orange-600 text-white rounded-lg hover:bg-orange-700 transition-colors"
                                                        >
                                                            <Eye className="w-4 h-4" />
                                                            <span>Review</span>
                                                        </button>
                                                    </div>

                                                    <div className="text-xs text-gray-500 mt-2">
                                                        <p><span className="font-medium">Schema ID:</span> {schema.schemaId}</p>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                    </div>
                </div>

                {/* Approved Schemas Section */}
                <div className="bg-white rounded-xl shadow-lg overflow-hidden">
                    <div className="bg-gradient-to-r from-green-600 to-green-700 px-6 py-4">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center">
                                <CheckCircle className="w-6 h-6 text-white mr-3" />
                                <h2 className="text-xl font-semibold text-white">
                                    Approved Schemas ({filteredApproved.length})
                                </h2>
                            </div>
                            <button 
                                onClick={handleExportApproved}
                                className="flex items-center space-x-2 px-4 py-2 bg-green-800 text-white rounded-lg hover:bg-green-900 transition-colors"
                            >
                                <Download className="w-4 h-4" />
                                <span>Export</span>
                            </button>
                        </div>
                    </div>

                    {/* Approved Filters */}
                    <div className="p-6 border-b">
                        <div className="space-y-4">
                            <div className="flex flex-col lg:flex-row gap-4">
                                <div className="flex-1">
                                    <div className="relative">
                                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                                        <input
                                            type="text"
                                            placeholder="Search approved schemas..."
                                            value={approvedFilters.searchByName || ''}
                                            onChange={(e) => updateApprovedFilter('searchByName', e.target.value)}
                                            className="pl-10 pr-4 py-2 w-full border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500"
                                        />
                                    </div>
                                </div>
                                <div className="flex flex-col sm:flex-row gap-4">
                                    <select
                                        value={approvedFilters.schemaType || 'all'}
                                        onChange={(e) => updateApprovedFilter('schemaType', e.target.value as 'all' | 'data' | 'api' | 'event')}
                                        className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500"
                                    >
                                        <option value="all">All Types</option>
                                        {schemaTypes.map(type => (
                                            <option key={type} value={type}>
                                                {type.charAt(0).toUpperCase() + type.slice(1)}
                                            </option>
                                        ))}
                                    </select>
                                </div>
                            </div>
                            
                            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                                <input
                                    type="text"
                                    placeholder="Filter by submitter"
                                    value={approvedFilters.submittedBy || ''}
                                    onChange={(e) => updateApprovedFilter('submittedBy', e.target.value)}
                                    className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500"
                                />
                                <input
                                    type="text"
                                    placeholder="Filter by version"
                                    value={approvedFilters.version || ''}
                                    onChange={(e) => updateApprovedFilter('version', e.target.value)}
                                    className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500"
                                />
                                <input
                                    type="date"
                                    placeholder="Start Date"
                                    value={approvedFilters.startDate || ''}
                                    onChange={(e) => updateApprovedFilter('startDate', e.target.value)}
                                    className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500"
                                />
                                <input
                                    type="date"
                                    placeholder="End Date"
                                    value={approvedFilters.endDate || ''}
                                    onChange={(e) => updateApprovedFilter('endDate', e.target.value)}
                                    className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500"
                                />
                            </div>
                            
                            <div className="flex justify-end">
                                <button
                                    onClick={clearApprovedFilters}
                                    className="px-4 py-2 text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
                                >
                                    Clear Filters
                                </button>
                            </div>
                        </div>
                    </div>

                    <div className="max-h-96 overflow-y-auto">
                        {filteredApproved.length === 0 ? (
                            <div className="text-center py-12">
                                <CheckCircle className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                                <p className="text-gray-500 text-lg">
                                    {approvedFilters.searchByName || approvedFilters.schemaType !== 'all' || approvedFilters.submittedBy || approvedFilters.version || approvedFilters.startDate || approvedFilters.endDate
                                        ? 'No approved schemas match your filters' 
                                        : 'No approved schemas available'
                                    }
                                </p>
                            </div>
                        ) : (
                            <div className="divide-y divide-gray-200">
                                {filteredApproved.map((schema) => (
                                    <div key={schema.schemaId} className={`p-4 border-l-4 hover:bg-gray-50 transition-colors ${getSchemaTypeBorderColor(schema.schemaType)}`}>
                                        <div className="flex items-start space-x-3">
                                            <div className="flex-shrink-0 mt-1">
                                                {getSchemaTypeIcon(schema.schemaType)}
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <div className="flex items-center justify-between mb-2">
                                                    <h3 className="text-lg font-semibold text-gray-900">
                                                        {schema.name} v{schema.version}
                                                    </h3>
                                                    <div className="flex items-center text-xs text-gray-500">
                                                        <Calendar className="w-3 h-3 mr-1" />
                                                        Approved: {formatTimestamp(schema.approvedAt!)}
                                                    </div>
                                                </div>
                                                
                                                <p className="text-sm text-gray-600 mb-3">{schema.description}</p>
                                                
                                                <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 mb-3">
                                                    <div className="flex items-center text-sm text-gray-600">
                                                        <User className="w-4 h-4 mr-2 text-gray-400" />
                                                        <span>{schema.submittedBy}</span>
                                                    </div>
                                                    <div className="flex items-center text-sm text-gray-600">
                                                        <span className="font-medium">Fields:</span> 
                                                        <span className="ml-1">{schema.fields}</span>
                                                    </div>
                                                    <div className="flex items-center text-sm text-gray-600">
                                                        <span className="font-medium">Usage:</span> 
                                                        <span className="ml-1">{schema.usageCount}</span>
                                                    </div>
                                                </div>

                                                <div className="flex items-center justify-between">
                                                    <div className="flex items-center gap-2">
                                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                                            schema.schemaType === 'data' 
                                                                ? 'bg-blue-100 text-blue-800' 
                                                                : schema.schemaType === 'api'
                                                                ? 'bg-green-100 text-green-800'
                                                                : 'bg-purple-100 text-purple-800'
                                                        }`}>
                                                            {schema.schemaType.charAt(0).toUpperCase() + schema.schemaType.slice(1)}
                                                        </span>
                                                        
                                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                                                            Approved
                                                        </span>
                                                    </div>
                                                    
                                                    <button
                                                        onClick={() => handleEdit(schema)}
                                                        className="flex items-center space-x-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                                    >
                                                        <Edit3 className="w-4 h-4" />
                                                        <span>Edit</span>
                                                    </button>
                                                </div>

                                                <div className="text-xs text-gray-500 mt-2 space-y-1">
                                                    <p><span className="font-medium">Schema ID:</span> {schema.schemaId}</p>
                                                    <p><span className="font-medium">Submitted:</span> {formatTimestamp(schema.createdAt)} | <span className="font-medium">Approved by:</span> {schema.approvedBy}</p>
                                                </div>
                                            </div>
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
