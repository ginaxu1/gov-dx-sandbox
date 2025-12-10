import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
    Activity,
    Search,
    Download,
    RefreshCw,
    CheckCircle,
    XCircle,
    Info,
    Clock,
    ChevronLeft,
    ChevronRight
} from 'lucide-react';
import { LogService } from '../services/logService';
import type { LogEntry, LogResponse } from '../services/logService';

interface FilterOptions {
    status?: 'all' | 'failure' | 'success';
    startDate?: string;
    endDate?: string;
    byConsumerId?: string;
    byProviderId?: string;
    searchTerm?: string;
}

interface LogsProps {
    role: 'provider' | 'consumer';
    memberId: string;
}

export const Logs: React.FC<LogsProps> = ({ role, memberId }) => {
    const [logs, setLogs] = useState<LogEntry[]>([]);
    const [filteredLogs, setFilteredLogs] = useState<LogEntry[]>([]);
    const [logResponse, setLogResponse] = useState<LogResponse | null>(null);
    const [currentPage, setCurrentPage] = useState(1);
    const [pageSize, setPageSize] = useState(20);
    const [filters, setFilters] = useState<FilterOptions>({
        status: 'all',
        searchTerm: '',
        startDate: '',
        endDate: '',
        byConsumerId: role === 'provider' ? '' : undefined, // Only show for providers
        byProviderId: role === 'consumer' ? '' : undefined  // Only show for consumers
    });
    const [loading, setLoading] = useState(true);
    const [autoRefresh, setAutoRefresh] = useState(false);

    // Helper function to update filters
    const updateFilter = <K extends keyof FilterOptions>(key: K, value: FilterOptions[K]) => {
        const newFilters = { ...filters, [key]: value };
        setFilters(newFilters);

        // If this is a server-side filter, reset to page 1 and fetch new data
        if (key === 'status' || key === 'startDate' || key === 'endDate') {
            setCurrentPage(1);
            // Fetch with the new filters immediately
            fetchLogsWithFilters(1, newFilters);
        }
    };

    // Helper function to clear all filters
    const clearAllFilters = () => {
        const clearedFilters = {
            status: 'all' as const,
            searchTerm: '',
            startDate: '',
            endDate: '',
            byConsumerId: role === 'provider' ? '' : undefined,
            byProviderId: role === 'consumer' ? '' : undefined
        };
        setFilters(clearedFilters);
        setCurrentPage(1);
        fetchLogsWithFilters(1, clearedFilters);
    };

    const filtersRef = useRef(filters);
    useEffect(() => {
        filtersRef.current = filters;
    }, [filters]);

    // Helper function to fetch logs with specific filters
    const fetchLogsWithFilters = useCallback(async (page: number, currentFilters: FilterOptions) => {
        setLoading(true);
        try {
            const offset = (page - 1) * pageSize;
            const response = await LogService.fetchLogsWithParams({
                [role === 'consumer' ? 'consumerId' : 'providerId']: memberId,
                limit: pageSize,
                offset: offset,
                // Server-side filters
                status: currentFilters.status !== 'all' ? currentFilters.status : undefined,
                startDate: currentFilters.startDate || undefined,
                endDate: currentFilters.endDate || undefined
            });

            setLogResponse(response);
            setLogs(response.logs || []);
            setFilteredLogs(response.logs || []);
        } catch (error) {
            console.error('Error fetching logs:', error);
            // Optionally show user-friendly error message
            setLogResponse(null);
            setLogs([]);
            setFilteredLogs([]);
        } finally {
            setLoading(false);
        }
    }, [pageSize, role, memberId]);

    useEffect(() => {
        setCurrentPage(1); // Reset to first page when role/IDs change
        fetchLogsWithFilters(1, filtersRef.current);
    }, [role, memberId, fetchLogsWithFilters]);

    // Auto-refresh functionality
    useEffect(() => {
        if (autoRefresh) {
            const interval = setInterval(() => {
                fetchLogsWithFilters(currentPage, filtersRef.current);
            }, 30000); // Refresh every 30 seconds

            return () => clearInterval(interval);
        }
    }, [autoRefresh, currentPage, fetchLogsWithFilters]);

    // Client-side filtering for additional filters not supported by server
    // Note: For better performance, these filters should be moved to server-side
    useEffect(() => {
        let filtered = logs;

        // Filter by search term (client-side for now)
        if (filters.searchTerm) {
            filtered = filtered.filter(log =>
                log.requestedData.toLowerCase().includes(filters.searchTerm!.toLowerCase()) ||
                log.consumerId.toLowerCase().includes(filters.searchTerm!.toLowerCase()) ||
                log.providerId.toLowerCase().includes(filters.searchTerm!.toLowerCase()) ||
                log.id.toLowerCase().includes(filters.searchTerm!.toLowerCase())
            );
        }

        // Additional client-side filters for byConsumerId and byProviderId
        // (These should ideally be handled server-side as part of the main query)

        // Filter by consumer ID (only for providers)
        if (role === 'provider' && filters.byConsumerId) {
            filtered = filtered.filter(log =>
                log.consumerId.toLowerCase().includes(filters.byConsumerId!.toLowerCase())
            );
        }

        // Filter by provider ID (only for consumers)
        if (role === 'consumer' && filters.byProviderId) {
            filtered = filtered.filter(log =>
                log.providerId.toLowerCase().includes(filters.byProviderId!.toLowerCase())
            );
        }

        setFilteredLogs(filtered);
    }, [logs, filters, role]);

    const getLogIcon = (status: string) => {
        switch (status) {
            case 'failure':
                return <XCircle className="w-5 h-5 text-red-500" />;
            case 'success':
                return <CheckCircle className="w-5 h-5 text-green-500" />;
            default:
                return <Info className="w-5 h-5 text-blue-500" />;
        }
    };

    const getLogBorderColor = (status: string) => {
        switch (status) {
            case 'failure':
                return 'border-l-red-500 bg-red-50';
            case 'success':
                return 'border-l-green-500 bg-green-50';
            default:
                return 'border-l-blue-500 bg-blue-50';
        }
    };

    const formatTimestamp = (timestamp: string) => {
        return new Date(timestamp).toLocaleString();
    };

    const statuses = ['success', 'failure'];

    const handleRefresh = () => {
        fetchLogsWithFilters(currentPage, filters);
    };

    const handleExport = async () => {
        try {
            if (!memberId) {
                throw new Error(`Missing ${role} ID`);
            }

            // Convert current filters to query params for export (export all, not just current page)
            const queryParams = {
                [role === 'consumer' ? 'consumerId' : 'providerId']: memberId,
                status: filters.status !== 'all' ? filters.status : undefined,
                ...(role === 'provider' && filters.byConsumerId ? { consumerId: filters.byConsumerId } : {}),
                ...(role === 'consumer' && filters.byProviderId ? { providerId: filters.byProviderId } : {}),
                startDate: filters.startDate || undefined,
                endDate: filters.endDate || undefined,
                search: filters.searchTerm || undefined
                // Note: No limit/offset for export - we want all matching records
            };

            const blob = await LogService.exportLogs(queryParams, 'csv');
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${role}-logs-${new Date().toISOString().split('T')[0]}.csv`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
        } catch (error) {
            console.error('Error exporting logs:', error);
            // Optionally show user-friendly error message
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
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                {/* Header */}
                <div className="mb-8">
                    <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between">
                        <div className="mb-6 lg:mb-0">
                            <h1 className="text-4xl font-bold text-gray-900 mb-2">
                                System Logs
                            </h1>
                            <p className="text-lg text-gray-600">
                                Monitor system activities and troubleshoot issues
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
                                className={`flex items-center space-x-2 px-4 py-2 rounded-lg transition-colors ${autoRefresh
                                    ? 'bg-green-100 text-green-700 border border-green-300'
                                    : 'bg-white border border-gray-300 hover:bg-gray-50'
                                    }`}
                            >
                                <Activity className="w-4 h-4" />
                                <span>Auto Refresh</span>
                            </button>
                            <button
                                onClick={handleExport}
                                className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                            >
                                <Download className="w-4 h-4" />
                                <span>Export</span>
                            </button>
                        </div>
                    </div>
                </div>

                {/* Filters */}
                <div className="bg-white rounded-xl shadow-lg p-6 mb-8">
                    <div className="space-y-4">
                        {/* Search and Status Row */}
                        <div className="flex flex-col lg:flex-row gap-4">
                            <div className="flex-1">
                                <div className="relative">
                                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                                    <input
                                        type="text"
                                        placeholder="Search logs..."
                                        value={filters.searchTerm || ''}
                                        onChange={(e) => updateFilter('searchTerm', e.target.value)}
                                        className="pl-10 pr-4 py-2 w-full border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                    />
                                </div>
                            </div>
                            <div className="flex flex-col sm:flex-row gap-4">
                                <select
                                    value={filters.status || 'all'}
                                    onChange={(e) => updateFilter('status', e.target.value as 'all' | 'failure' | 'success')}
                                    className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                >
                                    <option value="all">All Status</option>
                                    {statuses.map(status => (
                                        <option key={status} value={status}>
                                            {status.charAt(0).toUpperCase() + status.slice(1)}
                                        </option>
                                    ))}
                                </select>
                            </div>
                        </div>

                        {/* Additional Filters Row - Role specific */}
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {/* Show Consumer ID filter only for providers */}
                            {role === 'provider' && (
                                <input
                                    type="text"
                                    placeholder="Filter by Consumer ID"
                                    value={filters.byConsumerId || ''}
                                    onChange={(e) => updateFilter('byConsumerId', e.target.value)}
                                    className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                />
                            )}

                            {/* Show Provider ID filter only for consumers */}
                            {role === 'consumer' && (
                                <input
                                    type="text"
                                    placeholder="Filter by Provider ID"
                                    value={filters.byProviderId || ''}
                                    onChange={(e) => updateFilter('byProviderId', e.target.value)}
                                    className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                />
                            )}

                            <input
                                type="date"
                                placeholder="Start Date"
                                value={filters.startDate || ''}
                                onChange={(e) => updateFilter('startDate', e.target.value)}
                                className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            <input
                                type="date"
                                placeholder="End Date"
                                value={filters.endDate || ''}
                                onChange={(e) => updateFilter('endDate', e.target.value)}
                                className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                        </div>

                        {/* Clear Filters Button */}
                        <div className="flex justify-end">
                            <button
                                onClick={clearAllFilters}
                                className="px-4 py-2 text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
                            >
                                Clear All Filters
                            </button>
                        </div>
                    </div>
                </div>

                {/* Log Statistics */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                    {/* Total Logs */}
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Total Logs</p>
                                <p className="text-2xl font-bold text-gray-900">{logResponse?.total || 0}</p>
                                <p className="text-xs text-gray-500">server-side count</p>
                            </div>
                            <div className="p-3 rounded-full bg-gray-100">
                                <Activity className="w-5 h-5 text-blue-500" />
                            </div>
                        </div>
                    </div>

                    {/* Current Page Stats */}
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Current Page</p>
                                <p className="text-2xl font-bold text-gray-900">{filteredLogs.length}</p>
                                <p className="text-xs text-gray-500">of {pageSize} per page</p>
                            </div>
                            <div className="p-3 rounded-full bg-gray-100">
                                <Activity className="w-5 h-5 text-green-500" />
                            </div>
                        </div>
                    </div>

                    {statuses.map(status => {
                        const count = filteredLogs.filter(log => log.status === status).length;
                        const percentage = filteredLogs.length > 0 ? (count / filteredLogs.length * 100) : 0;

                        return (
                            <div key={status} className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                                <div className="flex items-center justify-between">
                                    <div>
                                        <p className="text-sm font-medium text-gray-600 capitalize">{status} (Page)</p>
                                        <p className="text-2xl font-bold text-gray-900">{count}</p>
                                        <p className="text-xs text-gray-500">{percentage.toFixed(1)}% of current page</p>
                                    </div>
                                    <div className="p-3 rounded-full bg-gray-100">
                                        {getLogIcon(status)}
                                    </div>
                                </div>
                            </div>
                        );
                    }).slice(0, 1)} {/* Only show first status to keep 3 cards total */}
                </div>

                {/* Pagination Controls */}
                {logResponse && logResponse.total > 0 && (
                    <div className="bg-white rounded-xl shadow-lg p-6 mb-8">
                        <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
                            <div className="flex items-center gap-4">
                                <label className="text-sm font-medium text-gray-700">
                                    Page size:
                                </label>
                                <select
                                    value={pageSize}
                                    onChange={(e) => {
                                        const newPageSize = parseInt(e.target.value);
                                        setPageSize(newPageSize);
                                        setCurrentPage(1);
                                        fetchLogsWithFilters(1, filters);
                                    }}
                                    className="px-3 py-1 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                >
                                    <option value={10}>10</option>
                                    <option value={20}>20</option>
                                    <option value={50}>50</option>
                                    <option value={100}>100</option>
                                </select>
                            </div>

                            <div className="flex items-center gap-2">
                                <span className="text-sm text-gray-700">
                                    Showing {((currentPage - 1) * pageSize) + 1} to {Math.min(currentPage * pageSize, logResponse.total)} of {logResponse.total} entries
                                </span>
                            </div>

                            <div className="flex items-center gap-2">
                                <button
                                    onClick={() => {
                                        const newPage = currentPage - 1;
                                        setCurrentPage(newPage);
                                        fetchLogsWithFilters(newPage, filters);
                                    }}
                                    disabled={currentPage === 1}
                                    className="flex items-center gap-1 px-3 py-1 text-sm bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    <ChevronLeft className="w-4 h-4" />
                                    Previous
                                </button>

                                <span className="px-3 py-1 text-sm">
                                    Page {currentPage} of {Math.ceil(logResponse.total / pageSize)}
                                </span>

                                <button
                                    onClick={() => {
                                        const newPage = currentPage + 1;
                                        setCurrentPage(newPage);
                                        fetchLogsWithFilters(newPage, filters);
                                    }}
                                    disabled={currentPage >= Math.ceil(logResponse.total / pageSize)}
                                    className="flex items-center gap-1 px-3 py-1 text-sm bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    Next
                                    <ChevronRight className="w-4 h-4" />
                                </button>
                            </div>
                        </div>
                    </div>
                )}

                {/* Logs List */}
                <div className="bg-white rounded-xl shadow-lg overflow-hidden">
                    <div className="bg-gradient-to-r from-gray-600 to-gray-700 px-6 py-4">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center">
                                <Activity className="w-6 h-6 text-white mr-3" />
                                <h2 className="text-xl font-semibold text-white">
                                    Log Entries
                                    {logResponse && (
                                        <span className="text-sm font-normal text-gray-200 ml-2">
                                            (Page {currentPage}: {filteredLogs.length} of {logResponse.total} total)
                                        </span>
                                    )}
                                </h2>
                            </div>
                            <div className="text-sm text-gray-200">
                                Last updated: {new Date().toLocaleTimeString()}
                            </div>
                        </div>
                    </div>
                    <div className="max-h-96 overflow-y-auto">
                        {filteredLogs.length === 0 ? (
                            <div className="text-center py-12">
                                <Activity className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                                <p className="text-gray-500 text-lg">
                                    {filters.searchTerm || filters.status !== 'all' ||
                                        (role === 'provider' && filters.byConsumerId) ||
                                        (role === 'consumer' && filters.byProviderId) ||
                                        filters.startDate || filters.endDate
                                        ? 'No logs match your filters'
                                        : 'No logs available'
                                    }
                                </p>
                            </div>
                        ) : (
                            <div className="divide-y divide-gray-200">
                                {filteredLogs.map((log) => (
                                    <div key={log.id} className={`p-4 border-l-4 hover:bg-gray-50 transition-colors ${getLogBorderColor(log.status)}`}>
                                        <div className="flex items-start space-x-3">
                                            <div className="flex-shrink-0 mt-1">
                                                {getLogIcon(log.status)}
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <div className="flex items-center justify-between mb-1">
                                                    <p className="text-sm font-medium text-gray-900">
                                                        Data Request: {log.requestedData}
                                                    </p>
                                                    <div className="flex items-center text-xs text-gray-500">
                                                        <Clock className="w-3 h-3 mr-1" />
                                                        {formatTimestamp(log.timestamp)}
                                                    </div>
                                                </div>
                                                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mb-2">
                                                    <p className="text-sm text-gray-600">
                                                        <span className="font-medium">Consumer:</span> {log.consumerId}
                                                    </p>
                                                    <p className="text-sm text-gray-600">
                                                        <span className="font-medium">Provider:</span> {log.providerId}
                                                    </p>
                                                </div>
                                                <div className="flex items-center gap-2">
                                                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${log.status === 'success'
                                                        ? 'bg-green-100 text-green-800'
                                                        : 'bg-red-100 text-red-800'
                                                        }`}>
                                                        {log.status.charAt(0).toUpperCase() + log.status.slice(1)}
                                                    </span>
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
