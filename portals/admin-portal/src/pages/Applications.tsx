import React, { useState, useEffect } from 'react';
import { 
    FileText, 
    Search, 
    Download, 
    RefreshCw, 
    Eye,
    Edit3,
    Clock,
    CheckCircle,
    User,
    Calendar
} from 'lucide-react';

import type { ApprovedApplication, ApplicationSubmission } from '../types/applications';
import { ApplicationService } from '../services/applicationService';

interface FilterOptionsApproved {
    searchByName?: string; 
    searchByDescription?: string;
    searchByConsumerId?: string;
    searchByVersion?: string;
}

interface FilterOptionsSubmissions {
    searchByName?: string;
    searchByDescription?: string;
    searchByConsumerId?: string;
    searchByStatus?: string;
}
interface ApplicationsProps {
}

export const Applications: React.FC<ApplicationsProps> = () => {
    const [submissions, setSubmissions] = useState<ApplicationSubmission[]>([]);
    const [approved, setApproved] = useState<ApprovedApplication[]>([]);
    const [filteredSubmissions, setFilteredSubmissions] = useState<ApplicationSubmission[]>([]);
    const [filteredApproved, setFilteredApproved] = useState<ApprovedApplication[]>([]);
    const [loading, setLoading] = useState(false);
    const [autoRefresh, setAutoRefresh] = useState(false);
    const [submissionFilters, setSubmissionFilters] = useState<FilterOptionsSubmissions>({
        searchByName: '',
        searchByDescription: '',
        searchByConsumerId: '',
        searchByStatus: ''
    });
    const [approvedFilters, setApprovedFilters] = useState<FilterOptionsApproved>({
        searchByName: '',
        searchByDescription: '',
        searchByConsumerId: '',
        searchByVersion: ''
    });
    const updateSubmissionFilter = (key: keyof FilterOptionsSubmissions, value: string) => { 
        setSubmissionFilters(prev => ({ ...prev, [key]: value }));
    };
    const clearSubmissionFilters = () => {
        setSubmissionFilters({
            searchByName: '',
            searchByDescription: '',
            searchByConsumerId: '',
            searchByStatus: ''
        });
    };
    const updateApprovedFilter = (key: keyof FilterOptionsApproved, value: string) => {
        setApprovedFilters(prev => ({ ...prev, [key]: value }));
    };
    const clearApprovedFilters = () => {
        setApprovedFilters({
            searchByName: '',
            searchByDescription: '',
            searchByConsumerId: '',
            searchByVersion: ''
        });
    };

    const fetchApplications = async () => {
        setLoading(true);
        try {
            const [submissionsData, approvedData] = await Promise.all([
                ApplicationService.getApplicationSubmissions(),
                ApplicationService.getApprovedApplications()
            ]);
            
            setSubmissions(submissionsData);
            setApproved(approvedData);
            setFilteredSubmissions(submissionsData);
            setFilteredApproved(approvedData);
        } catch (error) {
            console.error('Error fetching applications:', error);
            setSubmissions([]);
            setApproved([]);
            setFilteredSubmissions([]);
            setFilteredApproved([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchApplications();
    }, []);

    // Auto-refresh functionality
    useEffect(() => {
        if (autoRefresh) {
            const interval = setInterval(() => {
                fetchApplications();
            }, 30000); // Refresh every 30 seconds

            return () => clearInterval(interval);
        }
    }, [autoRefresh]);

    useEffect(() => {
        let filtered = submissions;
        if (submissionFilters.searchByName) {
            filtered = filtered.filter(app =>
                app.applicationName.toLowerCase().includes(submissionFilters.searchByName!.toLowerCase())
            );
        }
        if (submissionFilters.searchByDescription) {
            filtered = filtered.filter(app =>
                app.applicationDescription?.toLowerCase().includes(submissionFilters.searchByDescription!.toLowerCase())
            );
        }
        if (submissionFilters.searchByConsumerId) {
            filtered = filtered.filter(app =>
                app.consumerId.toLowerCase().includes(submissionFilters.searchByConsumerId!.toLowerCase())
            );
        }
        if (submissionFilters.searchByStatus && submissionFilters.searchByStatus !== 'all') {
            filtered = filtered.filter(app =>
                app.status === submissionFilters.searchByStatus
            );
        }
        setFilteredSubmissions(filtered);
    }, [submissions, submissionFilters]);

    useEffect(() => {
        let filtered = approved;
        if (approvedFilters.searchByName) {
            filtered = filtered.filter(app =>
                app.applicationName.toLowerCase().includes(approvedFilters.searchByName!.toLowerCase())
            );
        }
        if (approvedFilters.searchByDescription) {
            filtered = filtered.filter(app =>
                app.applicationDescription?.toLowerCase().includes(approvedFilters.searchByDescription!.toLowerCase())
            );
        }
        if (approvedFilters.searchByConsumerId) {
            filtered = filtered.filter(app =>
                app.consumerId.toLowerCase().includes(approvedFilters.searchByConsumerId!.toLowerCase())
            );
        }
        if (approvedFilters.searchByVersion && approvedFilters.searchByVersion !== 'all') {
            filtered = filtered.filter(app =>
                app.version === approvedFilters.searchByVersion
            );
        }
        setFilteredApproved(filtered);
    }, [approved, approvedFilters]);

    const formatTimestamp = (timestamp: string) => {
        return new Date(timestamp).toLocaleString();
    };

    // Helper function for application type icons (placeholder since we don't have applicationTypes)
    const getApplicationTypeIcon = (_consumerId: string) => {
        return <FileText className="w-5 h-5 text-blue-500" />;
    };

    // Helper function for application type border colors (placeholder)
    const getApplicationTypeBorderColor = (_consumerId: string) => {
        return 'border-l-blue-500 bg-blue-50';
    };

    const handleRefresh = () => {
        fetchApplications();
    };

    const handleExportSubmissions = async () => {
        try {
            const dataToExport = filteredSubmissions.map(app => ({
                submissionId: app.submissionId,
                applicationName: app.applicationName,
                applicationDescription: app.applicationDescription || '',
                consumerId: app.consumerId,
                status: app.status,
                selectedFields: app.selectedFields?.join('; ') || '',
                fieldCount: app.selectedFields?.length || 0
            }));

            const csvContent = [
                // CSV headers
                ['Submission ID', 'Application Name', 'Description', 'Consumer ID', 'Status', 'Selected Fields', 'Field Count'].join(','),
                // CSV data
                ...dataToExport.map(row => [
                    row.submissionId,
                    `"${row.applicationName.replace(/"/g, '""')}"`,
                    `"${row.applicationDescription.replace(/"/g, '""')}"`,
                    row.consumerId,
                    row.status,
                    `"${row.selectedFields.replace(/"/g, '""')}"`,
                    row.fieldCount
                ].join(','))
            ].join('\n');

            const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
            const link = document.createElement('a');
            const url = URL.createObjectURL(blob);
            link.setAttribute('href', url);
            link.setAttribute('download', `application-submissions-${new Date().toISOString().split('T')[0]}.csv`);
            link.style.visibility = 'hidden';
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        } catch (error) {
            console.error('Error exporting application submissions:', error);
        }
    };

    const handleExportApproved = async () => {
        try {
            const dataToExport = filteredApproved.map(app => ({
                applicationId: app.applicationId,
                applicationName: app.applicationName,
                applicationDescription: app.applicationDescription || '',
                consumerId: app.consumerId,
                version: app.version,
                selectedFields: app.selectedFields?.join('; ') || '',
                fieldCount: app.selectedFields?.length || 0,
                createdAt: app.createdAt,
                updatedAt: app.updatedAt
            }));

            const csvContent = [
                // CSV headers
                ['Application ID', 'Application Name', 'Description', 'Consumer ID', 'Version', 'Selected Fields', 'Field Count', 'Created At', 'Updated At'].join(','),
                // CSV data
                ...dataToExport.map(row => [
                    row.applicationId,
                    `"${row.applicationName.replace(/"/g, '""')}"`,
                    `"${row.applicationDescription.replace(/"/g, '""')}"`,
                    row.consumerId,
                    row.version,
                    `"${row.selectedFields.replace(/"/g, '""')}"`,
                    row.fieldCount,
                    row.createdAt,
                    row.updatedAt
                ].join(','))
            ].join('\n');

            const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
            const link = document.createElement('a');
            const url = URL.createObjectURL(blob);
            link.setAttribute('href', url);
            link.setAttribute('download', `approved-applications-${new Date().toISOString().split('T')[0]}.csv`);
            link.style.visibility = 'hidden';
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        } catch (error) {
            console.error('Error exporting approved applications:', error);
        }
    };

    const handleReview = (application: ApplicationSubmission) => {
        // TODO: Navigate to application review page or open modal
        // For now, log the application details
        console.log('Opening application for review:', {
            submissionId: application.submissionId,
            applicationName: application.applicationName,
            consumerId: application.consumerId,
            status: application.status,
            selectedFields: application.selectedFields,
            fieldCount: application.selectedFields?.length || 0
        });
        
        // In a real implementation, this would:
        // 1. Navigate to a dedicated review page: navigate(`/applications/review/${application.submissionId}`)
        // 2. Or open a modal with application details and approval/rejection controls
        // 3. Allow viewing and editing the selected fields
        // 4. Provide approval/rejection functionality with comments
        // 5. Show field configuration and access control settings
    };

    const handleEdit = (application: ApprovedApplication) => {
        // TODO: Navigate to application edit page or open modal
        // For now, log the application details
        console.log('Opening application for editing:', {
            applicationId: application.applicationId,
            applicationName: application.applicationName,
            consumerId: application.consumerId,
            version: application.version,
            selectedFields: application.selectedFields,
            fieldCount: application.selectedFields?.length || 0
        });
        
        // In a real implementation, this would:
        // 1. Navigate to a dedicated edit page: navigate(`/applications/edit/${application.applicationId}`)
        // 2. Or open a modal with editable application details
        // 3. Allow updating application name, description, and selected fields
        // 4. Provide version management capabilities
        // 5. Handle application update through ApplicationService.updateApplication
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
                                Applications
                            </h1>
                            <p className="text-lg text-gray-600">
                                Manage application submissions and approved applications
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
                                <FileText className="w-4 h-4" />
                                <span>Auto Refresh</span>
                            </button>
                        </div>
                    </div>
                </div>

                {/* Applications Statistics */}
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
                                <p className="text-sm font-medium text-gray-600">Approved Apps</p>
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
                                <p className="text-2xl font-bold text-gray-900">{submissions.filter(app => app.status === 'pending').length}</p>
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
                                <p className="text-sm font-medium text-gray-600">Active Versions</p>
                                <p className="text-2xl font-bold text-gray-900">{approved.filter(app => app.version === 'active').length}</p>
                                <p className="text-xs text-gray-500">currently active</p>
                            </div>
                            <div className="p-3 rounded-full bg-blue-100">
                                <FileText className="w-5 h-5 text-blue-500" />
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
                                        Pending Submissions ({filteredSubmissions.length})
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
                                                placeholder="Search by application name..."
                                                value={submissionFilters.searchByName || ''}
                                                onChange={(e) => updateSubmissionFilter('searchByName', e.target.value)}
                                                className="pl-10 pr-4 py-2 w-full border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-orange-500"
                                            />
                                        </div>
                                    </div>
                                    <div className="flex flex-col sm:flex-row gap-4">
                                        <select
                                            value={submissionFilters.searchByStatus || 'all'}
                                            onChange={(e) => updateSubmissionFilter('searchByStatus', e.target.value)}
                                            className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-orange-500"
                                        >
                                            <option value="all">All Status</option>
                                            <option value="pending">Pending</option>
                                            <option value="approved">Approved</option>
                                            <option value="rejected">Rejected</option>
                                        </select>
                                    </div>
                                </div>
                                
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <input
                                        type="text"
                                        placeholder="Filter by description"
                                        value={submissionFilters.searchByDescription || ''}
                                        onChange={(e) => updateSubmissionFilter('searchByDescription', e.target.value)}
                                        className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-orange-500"
                                    />
                                    <input
                                        type="text"
                                        placeholder="Filter by consumer ID"
                                        value={submissionFilters.searchByConsumerId || ''}
                                        onChange={(e) => updateSubmissionFilter('searchByConsumerId', e.target.value)}
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
                                        {submissionFilters.searchByName || submissionFilters.searchByDescription || submissionFilters.searchByConsumerId || (submissionFilters.searchByStatus && submissionFilters.searchByStatus !== 'all')
                                            ? 'No submissions match your filters' 
                                            : 'No submissions available'
                                        }
                                    </p>
                                </div>
                            ) : (
                                <div className="divide-y divide-gray-200">
                                    {filteredSubmissions.map((application) => (
                                        <div key={application.submissionId} className={`p-4 border-l-4 hover:bg-gray-50 transition-colors ${getApplicationTypeBorderColor(application.consumerId)}`}>
                                            <div className="flex items-start space-x-3">
                                                <div className="flex-shrink-0 mt-1">
                                                    {getApplicationTypeIcon(application.consumerId)}
                                                </div>
                                                <div className="flex-1 min-w-0">
                                                    <div className="flex items-center justify-between mb-2">
                                                        <h3 className="text-lg font-semibold text-gray-900">
                                                            {application.applicationName}
                                                        </h3>
                                                        <div className="flex items-center text-xs text-gray-500">
                                                            <Calendar className="w-3 h-3 mr-1" />
                                                            Consumer: {application.consumerId}
                                                        </div>
                                                    </div>
                                                    
                                                    <p className="text-sm text-gray-600 mb-3">{application.applicationDescription || 'No description provided'}</p>
                                                    
                                                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mb-3">
                                                        <div className="flex items-center text-sm text-gray-600">
                                                            <User className="w-4 h-4 mr-2 text-gray-400" />
                                                            <span>Fields: {application.selectedFields?.length || 0}</span>
                                                        </div>
                                                        <div className="flex items-center text-sm text-gray-600">
                                                            <span className="font-medium">Status:</span> 
                                                            <span className="ml-1 capitalize">{application.status}</span>
                                                        </div>
                                                    </div>

                                                    <div className="flex items-center justify-between">
                                                        <div className="flex items-center gap-2">
                                                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                                                application.status === 'pending' 
                                                                    ? 'bg-orange-100 text-orange-800' 
                                                                    : application.status === 'approved'
                                                                    ? 'bg-green-100 text-green-800'
                                                                    : 'bg-red-100 text-red-800'
                                                            }`}>
                                                                {application.status.charAt(0).toUpperCase() + application.status.slice(1)}
                                                            </span>
                                                            
                                                            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                                                Submission
                                                            </span>
                                                        </div>
                                                        
                                                        <button
                                                            onClick={() => handleReview(application)}
                                                            className="flex items-center space-x-2 px-4 py-2 bg-orange-600 text-white rounded-lg hover:bg-orange-700 transition-colors"
                                                        >
                                                            <Eye className="w-4 h-4" />
                                                            <span>Review</span>
                                                        </button>
                                                    </div>

                                                    <div className="text-xs text-gray-500 mt-2">
                                                        <p><span className="font-medium">Submission ID:</span> {application.submissionId}</p>
                                                        {application.selectedFields && application.selectedFields.length > 0 && (
                                                            <p><span className="font-medium">Selected Fields:</span> {application.selectedFields.join(', ')}</p>
                                                        )}
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

                {/* Approved Applications Section */}
                <div className="bg-white rounded-xl shadow-lg overflow-hidden">
                    <div className="bg-gradient-to-r from-green-600 to-green-700 px-6 py-4">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center">
                                <CheckCircle className="w-6 h-6 text-white mr-3" />
                                <h2 className="text-xl font-semibold text-white">
                                    Approved Applications ({filteredApproved.length})
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
                                            placeholder="Search by application name..."
                                            value={approvedFilters.searchByName || ''}
                                            onChange={(e) => updateApprovedFilter('searchByName', e.target.value)}
                                            className="pl-10 pr-4 py-2 w-full border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500"
                                        />
                                    </div>
                                </div>
                                <div className="flex flex-col sm:flex-row gap-4">
                                    <select
                                        value={approvedFilters.searchByVersion || 'all'}
                                        onChange={(e) => updateApprovedFilter('searchByVersion', e.target.value)}
                                        className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500"
                                    >
                                        <option value="all">All Versions</option>
                                        <option value="active">Active</option>
                                        <option value="deprecated">Deprecated</option>
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
                                    placeholder="Filter by consumer ID"
                                    value={approvedFilters.searchByConsumerId || ''}
                                    onChange={(e) => updateApprovedFilter('searchByConsumerId', e.target.value)}
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
                                    {approvedFilters.searchByName || approvedFilters.searchByDescription || approvedFilters.searchByConsumerId || (approvedFilters.searchByVersion && approvedFilters.searchByVersion !== 'all')
                                        ? 'No approved applications match your filters' 
                                        : 'No approved applications available'
                                    }
                                </p>
                            </div>
                        ) : (
                            <div className="divide-y divide-gray-200">
                                {filteredApproved.map((application) => (
                                    <div key={application.applicationId} className={`p-4 border-l-4 hover:bg-gray-50 transition-colors ${getApplicationTypeBorderColor(application.consumerId)}`}>
                                        <div className="flex items-start space-x-3">
                                            <div className="flex-shrink-0 mt-1">
                                                {getApplicationTypeIcon(application.consumerId)}
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <div className="flex items-center justify-between mb-2">
                                                    <h3 className="text-lg font-semibold text-gray-900">
                                                        {application.applicationName}
                                                    </h3>
                                                    <div className="flex items-center text-xs text-gray-500">
                                                        <Calendar className="w-3 h-3 mr-1" />
                                                        Updated: {formatTimestamp(application.updatedAt)}
                                                    </div>
                                                </div>
                                                
                                                <p className="text-sm text-gray-600 mb-3">{application.applicationDescription || 'No description provided'}</p>
                                                
                                                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mb-3">
                                                    <div className="flex items-center text-sm text-gray-600">
                                                        <User className="w-4 h-4 mr-2 text-gray-400" />
                                                        <span>Consumer: {application.consumerId}</span>
                                                    </div>
                                                    <div className="flex items-center text-sm text-gray-600">
                                                        <span className="font-medium">Fields:</span> 
                                                        <span className="ml-1">{application.selectedFields?.length || 0}</span>
                                                    </div>
                                                </div>

                                                <div className="flex items-center justify-between">
                                                    <div className="flex items-center gap-2">
                                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                                            application.version === 'active' 
                                                                ? 'bg-green-100 text-green-800' 
                                                                : 'bg-gray-100 text-gray-800'
                                                        }`}>
                                                            {application.version.charAt(0).toUpperCase() + application.version.slice(1)}
                                                        </span>
                                                        
                                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                                            Approved
                                                        </span>
                                                    </div>
                                                    
                                                    <button
                                                        onClick={() => handleEdit(application)}
                                                        className="flex items-center space-x-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                                    >
                                                        <Edit3 className="w-4 h-4" />
                                                        <span>Edit</span>
                                                    </button>
                                                </div>

                                                <div className="text-xs text-gray-500 mt-2 space-y-1">
                                                    <p><span className="font-medium">Application ID:</span> {application.applicationId}</p>
                                                    <p><span className="font-medium">Created:</span> {formatTimestamp(application.createdAt)} | <span className="font-medium">Updated:</span> {formatTimestamp(application.updatedAt)}</p>
                                                    {application.selectedFields && application.selectedFields.length > 0 && (
                                                        <p><span className="font-medium">Selected Fields:</span> {application.selectedFields.join(', ')}</p>
                                                    )}
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
