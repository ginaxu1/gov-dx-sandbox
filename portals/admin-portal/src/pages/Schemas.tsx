import React, { useState, useEffect } from 'react';
import { 
    Database, 
    Search, 
    Download, 
    RefreshCw, 
    Eye,
    Clock,
    CheckCircle,
    User,
    Calendar,
    FileCode,
    Layers,
    XCircle,
    X,
    ThumbsUp,
    ThumbsDown,
    MessageSquare
} from 'lucide-react';

import type { ApprovedSchema, SchemaSubmission  } from '../types/graphql';
import { SchemaService } from '../services/schemaService';

interface FilterOptionsApproved {
    searchByName?: string;
    searchByDescription?: string;
    searchByProviderId?: string;
    searchByVersion: 'active' | 'deprecated' | 'all';
}

const schemaVersions = ['active', 'deprecated'];
const schemaStatuses = ['pending', 'approved', 'rejected'];

interface FilterOptionsSubmissions {
    searchByName?: string;
    searchByDescription?: string;
    searchByProviderId?: string;
    searchByStatus: 'pending' | 'approved' | 'rejected' | 'all';
}

interface SchemasProps {
}

export const Schemas: React.FC<SchemasProps> = () => {
    const [submissions, setSubmissions] = useState<SchemaSubmission[]>([]);
    const [approved, setApproved] = useState<ApprovedSchema[]>([]);
    const [filteredSubmissions, setFilteredSubmissions] = useState<SchemaSubmission[]>([]);
    const [filteredApproved, setFilteredApproved] = useState<ApprovedSchema[]>([]);

    const getSchemaTypeIcon = (schemaStatus?: 'pending' | 'approved' | 'rejected', schemaVersion?: 'active' | 'deprecated') => {
        if (schemaStatus === 'pending') return <Clock className="text-yellow-500" />;
        if (schemaStatus === 'approved') return <CheckCircle className="text-green-500" />;
        if (schemaStatus === 'rejected') return <XCircle className="text-red-500" />;
        if (schemaVersion === 'active') return <Layers className="text-blue-500" />;
        if (schemaVersion === 'deprecated') return <FileCode className="text-gray-500" />;
        return <Database className="text-gray-400" />;
    };

    const getSchemaTypeBorderColor = (schemaStatus?: 'pending' | 'approved' | 'rejected', schemaVersion?: 'active' | 'deprecated') => {
        if (schemaStatus === 'pending') return 'border-yellow-500';
        if (schemaStatus === 'approved') return 'border-green-500';
        if (schemaStatus === 'rejected') return 'border-red-500';
        if (schemaVersion === 'active') return 'border-blue-500';
        if (schemaVersion === 'deprecated') return 'border-gray-500';
        return 'border-gray-200';
    };

    // Separate filters for submissions and approved
    const [submissionFilters, setSubmissionFilters] = useState<FilterOptionsSubmissions>({
        searchByName: '',
        searchByDescription: '',
        searchByProviderId: '',
        searchByStatus: 'all'
    });

    const [approvedFilters, setApprovedFilters] = useState<FilterOptionsApproved>({
        searchByName: '',
        searchByDescription: '',
        searchByProviderId: '',
        searchByVersion: 'all'
    });
    
    const [loading, setLoading] = useState(true);
    const [autoRefresh, setAutoRefresh] = useState(false);
    const [reviewModal, setReviewModal] = useState<{
        isOpen: boolean;
        schema: SchemaSubmission | null;
    }>({
        isOpen: false,
        schema: null
    });
    const [reviewComment, setReviewComment] = useState('');
    const [reviewAction, setReviewAction] = useState<'approve' | 'reject' | null>(null);
    const [submittingReview, setSubmittingReview] = useState(false);

    // Helper function to update submission filters
    const updateSubmissionFilter = <K extends keyof FilterOptionsSubmissions>(key: K, value: FilterOptionsSubmissions[K]) => {
        setSubmissionFilters(prev => ({ ...prev, [key]: value }));
    };

    // Helper function to update approved filters
    const updateApprovedFilter = <K extends keyof FilterOptionsApproved>(key: K, value: FilterOptionsApproved[K]) => {
        setApprovedFilters(prev => ({ ...prev, [key]: value }));
    };

    // Helper function to clear submission filters
    const clearSubmissionFilters = () => {
        setSubmissionFilters({
            searchByName: '',
            searchByDescription: '',
            searchByProviderId: '',
            searchByStatus: 'all'
        });
    };

    // Helper function to clear approved filters
    const clearApprovedFilters = () => {
        setApprovedFilters({
            searchByName: '',
            searchByDescription: '',
            searchByProviderId: '',
            searchByVersion: 'all'
        });
    };

    const fetchSubmissions = async () => {
        try {
            const data : SchemaSubmission[] = await SchemaService.getSchemaSubmissions();
            return data;
        } catch (error) {
            console.error('Error fetching schema submissions:', error);
            return [];
        }
    };

    const fetchApproved = async () => {
        try {
            const data : ApprovedSchema[] = await SchemaService.getApprovedSchemas();
            return data;
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
                schema.schemaName.toLowerCase().includes(submissionFilters.searchByName!.toLowerCase())
            );
        }
        if (submissionFilters.searchByDescription) {
            filtered = filtered.filter(schema =>
                schema.schemaDescription?.toLowerCase().includes(submissionFilters.searchByDescription!.toLowerCase())
            );
        }
        if (submissionFilters.searchByProviderId) {
            filtered = filtered.filter(schema =>
                schema.providerId.toLowerCase().includes(submissionFilters.searchByProviderId!.toLowerCase())
            );
        }
        if (submissionFilters.searchByStatus && submissionFilters.searchByStatus !== 'all') {
            filtered = filtered.filter(schema =>
                schema.status === submissionFilters.searchByStatus
            );
        }
        setFilteredSubmissions(filtered);
    }, [submissions, submissionFilters]);

    // Filter approved schemas
    useEffect(() => {
        let filtered = approved;
        if (approvedFilters.searchByName) {
            filtered = filtered.filter(schema =>
                schema.schemaName.toLowerCase().includes(approvedFilters.searchByName!.toLowerCase())
            );
        }
        if (approvedFilters.searchByDescription) {
            filtered = filtered.filter(schema =>
                schema.schemaDescription?.toLowerCase().includes(approvedFilters.searchByDescription!.toLowerCase())
            );
        }
        if (approvedFilters.searchByProviderId) {
            filtered = filtered.filter(schema =>
                schema.providerId.toLowerCase().includes(approvedFilters.searchByProviderId!.toLowerCase())
            );
        }
        if (approvedFilters.searchByVersion && approvedFilters.searchByVersion !== 'all') {
            filtered = filtered.filter(schema =>
                schema.version === approvedFilters.searchByVersion
            );
        }
        setFilteredApproved(filtered);
    }, [approved, approvedFilters]);

    const formatTimestamp = (timestamp: string) => {
        return new Date(timestamp).toLocaleString();
    };

    const handleRefresh = () => {
        fetchSchemas();
    };

    const handleExportSubmissions = async () => {
        try {
            const dataToExport = filteredSubmissions.map(schema => ({
                submissionId: schema.submissionId,
                schemaName: schema.schemaName,
                schemaDescription: schema.schemaDescription || '',
                providerId: schema.providerId,
                status: schema.status,
                schemaEndpoint: schema.schemaEndpoint,
                createdAt: schema.createdAt,
                updatedAt: schema.updatedAt
            }));

            const csvContent = [
                // CSV headers
                ['Submission ID', 'Schema Name', 'Description', 'Provider ID', 'Status', 'Endpoint', 'Created At', 'Updated At'].join(','),
                // CSV data
                ...dataToExport.map(row => [
                    row.submissionId,
                    `"${row.schemaName.replace(/"/g, '""')}"`,
                    `"${row.schemaDescription.replace(/"/g, '""')}"`,
                    row.providerId,
                    row.status,
                    `"${row.schemaEndpoint.replace(/"/g, '""')}"`,
                    row.createdAt,
                    row.updatedAt
                ].join(','))
            ].join('\n');

            const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
            const link = document.createElement('a');
            const url = URL.createObjectURL(blob);
            link.setAttribute('href', url);
            link.setAttribute('download', `schema-submissions-${new Date().toISOString().split('T')[0]}.csv`);
            link.style.visibility = 'hidden';
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        } catch (error) {
            console.error('Error exporting schema submissions:', error);
        }
    };

    const handleExportApproved = async () => {
        try {
            const dataToExport = filteredApproved.map(schema => ({
                schemaId: schema.schemaId,
                schemaName: schema.schemaName,
                schemaDescription: schema.schemaDescription || '',
                providerId: schema.providerId,
                version: schema.version,
                schemaEndpoint: schema.schemaEndpoint,
                createdAt: schema.createdAt,
                updatedAt: schema.updatedAt
            }));

            const csvContent = [
                // CSV headers
                ['Schema ID', 'Schema Name', 'Description', 'Provider ID', 'Version', 'Endpoint', 'Created At', 'Updated At'].join(','),
                // CSV data
                ...dataToExport.map(row => [
                    row.schemaId,
                    `"${row.schemaName.replace(/"/g, '""')}"`,
                    `"${row.schemaDescription.replace(/"/g, '""')}"`,
                    row.providerId,
                    row.version,
                    `"${row.schemaEndpoint.replace(/"/g, '""')}"`,
                    row.createdAt,
                    row.updatedAt
                ].join(','))
            ].join('\n');

            const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
            const link = document.createElement('a');
            const url = URL.createObjectURL(blob);
            link.setAttribute('href', url);
            link.setAttribute('download', `approved-schemas-${new Date().toISOString().split('T')[0]}.csv`);
            link.style.visibility = 'hidden';
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        } catch (error) {
            console.error('Error exporting approved schemas:', error);
        }
    };

    const handleReview = (schema: SchemaSubmission) => {
        setReviewModal({
            isOpen: true,
            schema: schema
        });
        setReviewComment('');
        setReviewAction(null);
    };

    const handleCloseReview = () => {
        setReviewModal({
            isOpen: false,
            schema: null
        });
        setReviewComment('');
        setReviewAction(null);
    };

    const handleSubmitReview = async () => {
        if (!reviewModal.schema || !reviewAction) return;

        setSubmittingReview(true);
        try {
            await SchemaService.addReviewToSchemaSubmission(reviewModal.schema.submissionId, reviewComment, reviewAction);

            // Refresh the data after review
            await fetchSchemas();
            
            // Close the modal
            handleCloseReview();
        } catch (error) {
            console.error('Error submitting review:', error);
        } finally {
            setSubmittingReview(false);
        }
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
            {/* Review Modal */}
            {reviewModal.isOpen && reviewModal.schema && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
                    <div className="bg-white rounded-xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
                        <div className="bg-gradient-to-r from-orange-600 to-orange-700 px-6 py-4 rounded-t-xl">
                            <div className="flex items-center justify-between">
                                <h2 className="text-xl font-semibold text-white">
                                    Review Schema Submission
                                </h2>
                                <button
                                    onClick={handleCloseReview}
                                    className="text-white hover:text-gray-200 transition-colors"
                                >
                                    <X className="w-6 h-6" />
                                </button>
                            </div>
                        </div>

                        <div className="p-6">
                            {/* Schema Details */}
                            <div className="mb-6">
                                <h3 className="text-lg font-semibold text-gray-900 mb-4">Schema Details</h3>
                                <div className="bg-gray-50 rounded-lg p-4 space-y-3">
                                    <div>
                                        <span className="font-medium text-gray-700">Name:</span>
                                        <span className="ml-2 text-gray-900">{reviewModal.schema.schemaName}</span>
                                    </div>
                                    <div>
                                        <span className="font-medium text-gray-700">Description:</span>
                                        <span className="ml-2 text-gray-900">{reviewModal.schema.schemaDescription || 'No description provided'}</span>
                                    </div>
                                    <div>
                                        <span className="font-medium text-gray-700">Provider ID:</span>
                                        <span className="ml-2 text-gray-900">{reviewModal.schema.providerId}</span>
                                    </div>
                                    <div>
                                        <span className="font-medium text-gray-700">Endpoint:</span>
                                        <span className="ml-2 text-gray-900 text-sm break-all">{reviewModal.schema.schemaEndpoint}</span>
                                    </div>
                                    <div>
                                        <span className="font-medium text-gray-700">Submission ID:</span>
                                        <span className="ml-2 text-gray-900 text-sm">{reviewModal.schema.submissionId}</span>
                                    </div>
                                    <div>
                                        <span className="font-medium text-gray-700">Submitted:</span>
                                        <span className="ml-2 text-gray-900">{formatTimestamp(reviewModal.schema.createdAt)}</span>
                                    </div>
                                </div>
                            </div>

                            {/* Review Action */}
                            <div className="mb-6">
                                <h3 className="text-lg font-semibold text-gray-900 mb-4">Review Action</h3>
                                <div className="flex gap-4 mb-4">
                                    <button
                                        onClick={() => setReviewAction('approve')}
                                        className={`flex items-center space-x-2 px-6 py-3 rounded-lg transition-colors ${
                                            reviewAction === 'approve'
                                                ? 'bg-green-600 text-white'
                                                : 'bg-green-100 text-green-700 hover:bg-green-200'
                                        }`}
                                    >
                                        <ThumbsUp className="w-5 h-5" />
                                        <span>Approve</span>
                                    </button>
                                    <button
                                        onClick={() => setReviewAction('reject')}
                                        className={`flex items-center space-x-2 px-6 py-3 rounded-lg transition-colors ${
                                            reviewAction === 'reject'
                                                ? 'bg-red-600 text-white'
                                                : 'bg-red-100 text-red-700 hover:bg-red-200'
                                        }`}
                                    >
                                        <ThumbsDown className="w-5 h-5" />
                                        <span>Reject</span>
                                    </button>
                                </div>
                            </div>

                            {/* Comment Section */}
                            <div className="mb-6">
                                <h3 className="text-lg font-semibold text-gray-900 mb-4">Comments</h3>
                                <div className="relative">
                                    <MessageSquare className="absolute left-3 top-3 text-gray-400 w-5 h-5" />
                                    <textarea
                                        value={reviewComment}
                                        onChange={(e) => setReviewComment(e.target.value)}
                                        placeholder="Add your review comments here..."
                                        rows={4}
                                        className="pl-10 pr-4 py-3 w-full border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-orange-500 resize-none"
                                    />
                                </div>
                            </div>

                            {/* Action Buttons */}
                            <div className="flex justify-end gap-4">
                                <button
                                    onClick={handleCloseReview}
                                    className="px-6 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
                                >
                                    Cancel
                                </button>
                                <button
                                    onClick={handleSubmitReview}
                                    disabled={!reviewAction || submittingReview}
                                    className={`px-6 py-2 rounded-lg transition-colors ${
                                        !reviewAction || submittingReview
                                            ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                                            : reviewAction === 'approve'
                                            ? 'bg-green-600 text-white hover:bg-green-700'
                                            : 'bg-red-600 text-white hover:bg-red-700'
                                    }`}
                                >
                                    {submittingReview ? 'Submitting...' : reviewAction === 'approve' ? 'Approve Schema' : 'Reject Schema'}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}

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
                    
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Pending Reviews</p>
                                <p className="text-2xl font-bold text-gray-900">{submissions.filter(s => s.status === 'pending').length}</p>
                                <p className="text-xs text-gray-500">awaiting action</p>
                            </div>
                            <div className="p-3 rounded-full bg-yellow-100">
                                <Clock className="w-5 h-5 text-yellow-500" />
                            </div>
                        </div>
                    </div>
                    
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Active Schemas</p>
                                <p className="text-2xl font-bold text-gray-900">{approved.filter(s => s.version === 'active').length}</p>
                                <p className="text-xs text-gray-500">currently active</p>
                            </div>
                            <div className="p-3 rounded-full bg-blue-100">
                                <Layers className="w-5 h-5 text-blue-500" />
                            </div>
                        </div>
                    </div>
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
                                            value={submissionFilters.searchByStatus || 'all'}
                                            onChange={(e) => updateSubmissionFilter('searchByStatus', e.target.value as FilterOptionsSubmissions['searchByStatus'])}
                                            className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-orange-500"
                                        >
                                            <option value="all">All Types</option>
                                            {schemaStatuses.map(type => (
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
                                        placeholder="Filter by description"
                                        value={submissionFilters.searchByDescription || ''}
                                        onChange={(e) => updateSubmissionFilter('searchByDescription', e.target.value)}
                                        className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-orange-500"
                                    />
                                    <input
                                        type="text"
                                        placeholder="Filter by provider ID"
                                        value={submissionFilters.searchByProviderId || ''}
                                        onChange={(e) => updateSubmissionFilter('searchByProviderId', e.target.value)}
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
                        
                        {/* Submission List */}
                        <div className="max-h-96 overflow-y-auto">
                            {filteredSubmissions.length === 0 ? (
                                <div className="text-center py-12">
                                    <Clock className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                                    <p className="text-gray-500 text-lg">
                                        {submissionFilters.searchByName || submissionFilters.searchByStatus !== 'all' || submissionFilters.searchByDescription || submissionFilters.searchByProviderId
                                            ? 'No schema submissions match your filters' 
                                            : 'No schema submissions available'
                                        }
                                    </p>
                                </div>
                            ) : (
                                <div className="divide-y divide-gray-200">
                                    {filteredSubmissions.map((schema) => (
                                        <div key={schema.submissionId} className={`p-4 border-l-4 hover:bg-gray-50 transition-colors ${getSchemaTypeBorderColor(schema.status, undefined)}`}>
                                            <div className="flex items-start space-x-3">
                                                <div className="flex-shrink-0 mt-1">
                                                    {getSchemaTypeIcon(schema.status, undefined)}
                                                </div>
                                                <div className="flex-1 min-w-0">
                                                    <div className="flex items-center justify-between mb-2">
                                                        <h3 className="text-lg font-semibold text-gray-900">
                                                            {schema.schemaName}
                                                        </h3>
                                                        <div className="flex items-center text-xs text-gray-500">
                                                            <Calendar className="w-3 h-3 mr-1" />
                                                            Submitted: {formatTimestamp(schema.createdAt)}
                                                        </div>
                                                    </div>
                                                    
                                                    <p className="text-sm text-gray-600 mb-3">{schema.schemaDescription}</p>
                                                    
                                                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mb-3">
                                                        <div className="flex items-center text-sm text-gray-600">
                                                            <User className="w-4 h-4 mr-2 text-gray-400" />
                                                            <span>Provider: {schema.providerId}</span>
                                                        </div>
                                                        <div className="flex items-center text-sm text-gray-600">
                                                            <span className="font-medium">Endpoint:</span> 
                                                            <span className="ml-1 text-xs truncate">{schema.schemaEndpoint}</span>
                                                        </div>
                                                    </div>

                                                    <div className="flex items-center justify-between">
                                                        <div className="flex items-center gap-2">
                                                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                                                schema.status === 'pending' 
                                                                    ? 'bg-yellow-100 text-yellow-800' 
                                                                    : schema.status === 'approved'
                                                                    ? 'bg-green-100 text-green-800'
                                                                    : 'bg-red-100 text-red-800'
                                                            }`}>
                                                                {schema.status.charAt(0).toUpperCase() + schema.status.slice(1)}
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
                                                        <p><span className="font-medium">Submission ID:</span> {schema.submissionId}</p>
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
                                        value={approvedFilters.searchByVersion || 'all'}
                                        onChange={(e) => updateApprovedFilter('searchByVersion', e.target.value as FilterOptionsApproved['searchByVersion'])}
                                        className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500"
                                    >
                                        <option value="all">All Versions</option>
                                        {schemaVersions.map(version => (
                                            <option key={version} value={version}>
                                                {version.charAt(0).toUpperCase() + version.slice(1)}
                                            </option>
                                        ))}
                                    </select>
                                </div>
                            </div>
                            
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <input
                                    type="text"
                                    placeholder="Filter by description"
                                    value={approvedFilters.searchByDescription || ''}
                                    onChange={(e) => updateApprovedFilter('searchByDescription', e.target.value)}
                                    className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500"
                                />
                                <input
                                    type="text"
                                    placeholder="Filter by provider ID"
                                    value={approvedFilters.searchByProviderId || ''}
                                    onChange={(e) => updateApprovedFilter('searchByProviderId', e.target.value)}
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
                                    {approvedFilters.searchByName || approvedFilters.searchByVersion !== 'all' || approvedFilters.searchByDescription || approvedFilters.searchByProviderId
                                        ? 'No approved schemas match your filters' 
                                        : 'No approved schemas available'
                                    }
                                </p>
                            </div>
                        ) : (
                            <div className="divide-y divide-gray-200">
                                {filteredApproved.map((schema) => (
                                    <div key={schema.schemaId} className={`p-4 border-l-4 hover:bg-gray-50 transition-colors ${getSchemaTypeBorderColor(undefined, schema.version)}`}>
                                        <div className="flex items-start space-x-3">
                                            <div className="flex-shrink-0 mt-1">
                                                {getSchemaTypeIcon(undefined, schema.version)}
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <div className="flex items-center justify-between mb-2">
                                                    <h3 className="text-lg font-semibold text-gray-900">
                                                        {schema.schemaName}
                                                    </h3>
                                                    <div className="flex items-center text-xs text-gray-500">
                                                        <Calendar className="w-3 h-3 mr-1" />
                                                        Updated: {formatTimestamp(schema.updatedAt)}
                                                    </div>
                                                </div>
                                                
                                                <p className="text-sm text-gray-600 mb-3">{schema.schemaDescription || 'No description available'}</p>
                                                
                                                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mb-3">
                                                    <div className="flex items-center text-sm text-gray-600">
                                                        <User className="w-4 h-4 mr-2 text-gray-400" />
                                                        <span>Provider: {schema.providerId}</span>
                                                    </div>
                                                    <div className="flex items-center text-sm text-gray-600">
                                                        <span className="font-medium">Endpoint:</span> 
                                                        <span className="ml-1 text-xs truncate">{schema.schemaEndpoint}</span>
                                                    </div>
                                                </div>

                                                <div className="flex items-center justify-between">
                                                    <div className="flex items-center gap-2">
                                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                                            schema.version === 'active' 
                                                                ? 'bg-blue-100 text-blue-800' 
                                                                : 'bg-gray-100 text-gray-800'
                                                        }`}>
                                                            {schema.version.charAt(0).toUpperCase() + schema.version.slice(1)}
                                                        </span>
                                                        
                                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                                                            Approved
                                                        </span>
                                                    </div>
                                                    
                                                    <button
                                                        onClick={() => console.log('View schema:', schema.schemaId)}
                                                        className="flex items-center space-x-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                                    >
                                                        <Eye className="w-4 h-4" />
                                                        <span>View</span>
                                                    </button>
                                                </div>

                                                <div className="text-xs text-gray-500 mt-2 space-y-1">
                                                    <p><span className="font-medium">Schema ID:</span> {schema.schemaId}</p>
                                                    <p><span className="font-medium">Created:</span> {formatTimestamp(schema.createdAt)}</p>
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
