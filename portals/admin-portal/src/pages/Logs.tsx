import React, { useState, useEffect } from 'react';
import { 
    Activity, 
    Search, 
    Download, 
    RefreshCw, 
    CheckCircle, 
    XCircle, 
    Info,
    Clock
} from 'lucide-react';
import { LogService } from '../services/logService';
import type { LogEntry } from '../services/logService';

interface FilterOptions {
    status?: 'all' | 'failure' | 'success';
    startDate?: string;
    endDate?: string;
    byConsumerId?: string;
    byProviderId?: string;
    searchTerm?: string;
}

interface LogsProps {
}

export const Logs: React.FC<LogsProps> = () => {
    const [logs, setLogs] = useState<LogEntry[]>([]);
    const [filteredLogs, setFilteredLogs] = useState<LogEntry[]>([]);
    const [filters, setFilters] = useState<FilterOptions>({
        status: 'all',
        searchTerm: '',
        startDate: '',
        endDate: '',
        byConsumerId: '',
        byProviderId: ''
    });
    const [loading, setLoading] = useState(true);
    const [autoRefresh, setAutoRefresh] = useState(false);

    // Helper function to update filters
    const updateFilter = <K extends keyof FilterOptions>(key: K, value: FilterOptions[K]) => {
        setFilters(prev => ({ ...prev, [key]: value }));
    };

    // Helper function to clear all filters
    const clearAllFilters = () => {
        setFilters({
            status: 'all',
            searchTerm: '',
            startDate: '',
            endDate: '',
            byConsumerId: '',
            byProviderId: ''
        });
    };

    const fetchLogs = async () => {
        setLoading(true);
        try {
            const logs = await LogService.fetchLogsWithParams();
            setLogs(logs);
            setFilteredLogs(logs);
        } catch (error) {
            console.error('Error fetching logs:', error);
            // Optionally show user-friendly error message
            setLogs([]);
            setFilteredLogs([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchLogs();
    }, []);

    // Auto-refresh functionality
    useEffect(() => {
        if (autoRefresh) {
            const interval = setInterval(() => {
                fetchLogs();
            }, 30000); // Refresh every 30 seconds

            return () => clearInterval(interval);
        }
    }, [autoRefresh]);

    useEffect(() => {
        let filtered = logs;
        
        // Filter by search term
        if (filters.searchTerm) {
            filtered = filtered.filter(log =>
                log.requestedData.toLowerCase().includes(filters.searchTerm!.toLowerCase()) ||
                log.consumerId.toLowerCase().includes(filters.searchTerm!.toLowerCase()) ||
                log.providerId.toLowerCase().includes(filters.searchTerm!.toLowerCase()) ||
                log.id.toLowerCase().includes(filters.searchTerm!.toLowerCase())
            );
        }

        // Filter by status
        if (filters.status && filters.status !== 'all') {
            filtered = filtered.filter(log => log.status === filters.status);
        }

        // Filter by consumer ID
        if (filters.byConsumerId) {
            filtered = filtered.filter(log => 
                log.consumerId.toLowerCase().includes(filters.byConsumerId!.toLowerCase())
            );
        }

        // Filter by provider ID
        if (filters.byProviderId) {
            filtered = filtered.filter(log => 
                log.providerId.toLowerCase().includes(filters.byProviderId!.toLowerCase())
            );
        }

        // Filter by date range
        if (filters.startDate) {
            filtered = filtered.filter(log => 
                new Date(log.timestamp) >= new Date(filters.startDate!)
            );
        }

        if (filters.endDate) {
            filtered = filtered.filter(log => 
                new Date(log.timestamp) <= new Date(filters.endDate!)
            );
        }

        setFilteredLogs(filtered);
    }, [logs, filters]);

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
        fetchLogs();
    };

    const handleExport = async () => {
        try {
            // Convert current filters to query params for export
            const queryParams = {
                status: filters.status !== 'all' ? filters.status : undefined,
                consumerId: filters.byConsumerId || undefined,
                providerId: filters.byProviderId || undefined,
                startDate: filters.startDate || undefined,
                endDate: filters.endDate || undefined,
                search: filters.searchTerm || undefined
            };

            const blob = await LogService.exportLogs(queryParams, 'csv');
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `logs-${new Date().toISOString().split('T')[0]}.csv`;
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
                                className={`flex items-center space-x-2 px-4 py-2 rounded-lg transition-colors ${
                                    autoRefresh 
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
                        
                        {/* Additional Filters Row */}
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                            <input
                                type="text"
                                placeholder="Filter by Consumer ID"
                                value={filters.byConsumerId || ''}
                                onChange={(e) => updateFilter('byConsumerId', e.target.value)}
                                className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            <input
                                type="text"
                                placeholder="Filter by Provider ID"
                                value={filters.byProviderId || ''}
                                onChange={(e) => updateFilter('byProviderId', e.target.value)}
                                className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
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
                                <p className="text-sm font-medium text-gray-600">Filtered Logs</p>
                                <p className="text-2xl font-bold text-gray-900">{filteredLogs.length}</p>
                                <p className="text-xs text-gray-500">of {logs.length} total</p>
                            </div>
                            <div className="p-3 rounded-full bg-gray-100">
                                <Activity className="w-5 h-5 text-blue-500" />
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
                                        <p className="text-sm font-medium text-gray-600 capitalize">{status} Logs</p>
                                        <p className="text-2xl font-bold text-gray-900">{count}</p>
                                        <p className="text-xs text-gray-500">{percentage.toFixed(1)}% of total</p>
                                    </div>
                                    <div className="p-3 rounded-full bg-gray-100">
                                        {getLogIcon(status)}
                                    </div>
                                </div>
                            </div>
                        );
                    })}
                </div>

                {/* Logs List */}
                <div className="bg-white rounded-xl shadow-lg overflow-hidden">
                    <div className="bg-gradient-to-r from-gray-600 to-gray-700 px-6 py-4">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center">
                                <Activity className="w-6 h-6 text-white mr-3" />
                                <h2 className="text-xl font-semibold text-white">
                                    Log Entries ({filteredLogs.length})
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
                                    {filters.searchTerm || filters.status !== 'all' || filters.byConsumerId || filters.byProviderId || filters.startDate || filters.endDate
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
                                                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                                        log.status === 'success' 
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
