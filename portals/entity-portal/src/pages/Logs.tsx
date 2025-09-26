import React, { useState, useEffect } from 'react';
import { 
    Activity, 
    Search, 
    Download, 
    RefreshCw, 
    AlertTriangle, 
    CheckCircle, 
    XCircle, 
    Info,
    Clock
} from 'lucide-react';

interface LogEntry {
    id: string;
    timestamp: string;
    level: 'info' | 'warning' | 'error' | 'success';
    service: string;
    message: string;
    details?: string;
}

export const Logs: React.FC = () => {
    const [logs, setLogs] = useState<LogEntry[]>([]);
    const [filteredLogs, setFilteredLogs] = useState<LogEntry[]>([]);
    const [searchTerm, setSearchTerm] = useState('');
    const [levelFilter, setLevelFilter] = useState<string>('all');
    const [serviceFilter, setServiceFilter] = useState<string>('all');
    const [loading, setLoading] = useState(true);
    const [autoRefresh, setAutoRefresh] = useState(false);

    // Mock log data
    const mockLogs: LogEntry[] = [
        {
            id: '1',
            timestamp: '2024-01-15T10:30:45Z',
            level: 'info',
            service: 'GraphQL Schema Registry',
            message: 'Schema registered successfully',
            details: 'Government Identity Schema v1.2 registered by provider-123'
        },
        {
            id: '2',
            timestamp: '2024-01-15T10:28:12Z',
            level: 'success',
            service: 'API Gateway',
            message: 'API request processed',
            details: 'GET /api/person/info - Response time: 145ms'
        },
        {
            id: '3',
            timestamp: '2024-01-15T10:25:33Z',
            level: 'warning',
            service: 'Authentication Service',
            message: 'Rate limit approaching',
            details: 'Client app-456 has made 950/1000 requests in the current hour'
        },
        {
            id: '4',
            timestamp: '2024-01-15T10:22:18Z',
            level: 'error',
            service: 'Data Provider',
            message: 'External service timeout',
            details: 'Healthcare Records API failed to respond within 30 seconds'
        },
        {
            id: '5',
            timestamp: '2024-01-15T10:20:05Z',
            level: 'info',
            service: 'Application Registry',
            message: 'New application approved',
            details: 'Tax Management Portal approved for production access'
        },
        {
            id: '6',
            timestamp: '2024-01-15T10:18:42Z',
            level: 'success',
            service: 'Schema Validator',
            message: 'Schema validation passed',
            details: 'Vehicle Registration Schema passed all validation checks'
        }
    ];

    useEffect(() => {
        const fetchLogs = async () => {
            setLoading(true);
            // Simulate API call
            await new Promise(resolve => setTimeout(resolve, 1000));
            setLogs(mockLogs);
            setFilteredLogs(mockLogs);
            setLoading(false);
        };

        fetchLogs();
    }, []);

    useEffect(() => {
        let filtered = logs;

        // Filter by search term
        if (searchTerm) {
            filtered = filtered.filter(log =>
                log.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
                log.service.toLowerCase().includes(searchTerm.toLowerCase()) ||
                log.details?.toLowerCase().includes(searchTerm.toLowerCase())
            );
        }

        // Filter by level
        if (levelFilter !== 'all') {
            filtered = filtered.filter(log => log.level === levelFilter);
        }

        // Filter by service
        if (serviceFilter !== 'all') {
            filtered = filtered.filter(log => log.service === serviceFilter);
        }

        setFilteredLogs(filtered);
    }, [logs, searchTerm, levelFilter, serviceFilter]);

    const getLogIcon = (level: string) => {
        switch (level) {
            case 'error':
                return <XCircle className="w-5 h-5 text-red-500" />;
            case 'warning':
                return <AlertTriangle className="w-5 h-5 text-yellow-500" />;
            case 'success':
                return <CheckCircle className="w-5 h-5 text-green-500" />;
            default:
                return <Info className="w-5 h-5 text-blue-500" />;
        }
    };

    const getLogBorderColor = (level: string) => {
        switch (level) {
            case 'error':
                return 'border-l-red-500 bg-red-50';
            case 'warning':
                return 'border-l-yellow-500 bg-yellow-50';
            case 'success':
                return 'border-l-green-500 bg-green-50';
            default:
                return 'border-l-blue-500 bg-blue-50';
        }
    };

    const formatTimestamp = (timestamp: string) => {
        return new Date(timestamp).toLocaleString();
    };

    const services = [...new Set(logs.map(log => log.service))];
    const levels = ['info', 'success', 'warning', 'error'];

    const handleRefresh = () => {
        setLogs([...mockLogs].sort(() => Math.random() - 0.5));
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
                            <button className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
                                <Download className="w-4 h-4" />
                                <span>Export</span>
                            </button>
                        </div>
                    </div>
                </div>

                {/* Filters */}
                <div className="bg-white rounded-xl shadow-lg p-6 mb-8">
                    <div className="flex flex-col lg:flex-row gap-4">
                        <div className="flex-1">
                            <div className="relative">
                                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                                <input
                                    type="text"
                                    placeholder="Search logs..."
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    className="pl-10 pr-4 py-2 w-full border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                />
                            </div>
                        </div>
                        <div className="flex flex-col sm:flex-row gap-4">
                            <select
                                value={levelFilter}
                                onChange={(e) => setLevelFilter(e.target.value)}
                                className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            >
                                <option value="all">All Levels</option>
                                {levels.map(level => (
                                    <option key={level} value={level}>
                                        {level.charAt(0).toUpperCase() + level.slice(1)}
                                    </option>
                                ))}
                            </select>
                            <select
                                value={serviceFilter}
                                onChange={(e) => setServiceFilter(e.target.value)}
                                className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            >
                                <option value="all">All Services</option>
                                {services.map(service => (
                                    <option key={service} value={service}>
                                        {service}
                                    </option>
                                ))}
                            </select>
                        </div>
                    </div>
                </div>

                {/* Log Statistics */}
                <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
                    {levels.map(level => {
                        const count = logs.filter(log => log.level === level).length;
                        const percentage = logs.length > 0 ? (count / logs.length * 100) : 0;
                        
                        return (
                            <div key={level} className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                                <div className="flex items-center justify-between">
                                    <div>
                                        <p className="text-sm font-medium text-gray-600 capitalize">{level} Logs</p>
                                        <p className="text-2xl font-bold text-gray-900">{count}</p>
                                        <p className="text-xs text-gray-500">{percentage.toFixed(1)}% of total</p>
                                    </div>
                                    <div className="p-3 rounded-full bg-gray-100">
                                        {getLogIcon(level)}
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
                                    {searchTerm || levelFilter !== 'all' || serviceFilter !== 'all' 
                                        ? 'No logs match your filters' 
                                        : 'No logs available'
                                    }
                                </p>
                            </div>
                        ) : (
                            <div className="divide-y divide-gray-200">
                                {filteredLogs.map((log) => (
                                    <div key={log.id} className={`p-4 border-l-4 hover:bg-gray-50 transition-colors ${getLogBorderColor(log.level)}`}>
                                        <div className="flex items-start space-x-3">
                                            <div className="flex-shrink-0 mt-1">
                                                {getLogIcon(log.level)}
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <div className="flex items-center justify-between mb-1">
                                                    <p className="text-sm font-medium text-gray-900">{log.message}</p>
                                                    <div className="flex items-center text-xs text-gray-500">
                                                        <Clock className="w-3 h-3 mr-1" />
                                                        {formatTimestamp(log.timestamp)}
                                                    </div>
                                                </div>
                                                <p className="text-sm text-gray-600 mb-2">{log.service}</p>
                                                {log.details && (
                                                    <p className="text-xs text-gray-500 bg-white p-2 rounded border">{log.details}</p>
                                                )}
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
