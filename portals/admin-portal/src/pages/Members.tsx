import React, { useState, useEffect } from 'react';
import { 
    Users, 
    Search, 
    Download, 
    RefreshCw, 
    Building2, 
    Shield,
    Mail,
    Phone,
    Calendar,
    User
} from 'lucide-react';
import { memberService } from '../services/memberService';
import type { Entity } from '../services/memberService';

interface FilterOptions {
    entityType?: 'all' | 'gov' | 'admin' | 'private';
    searchByName?: string;
    allProviders?: 'all' | 'providers-only' | 'non-providers';
    allConsumers?: 'all' | 'consumers-only' | 'non-consumers';
}

interface MembersProps {
}

export const Members: React.FC<MembersProps> = () => {
    const [entities, setEntities] = useState<Entity[]>([]);
    const [filteredEntities, setFilteredEntities] = useState<Entity[]>([]);
    const [filters, setFilters] = useState<FilterOptions>({
        entityType: 'all',
        searchByName: '',
        allProviders: 'all',
        allConsumers: 'all'
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
            entityType: 'all',
            searchByName: '',
            allProviders: 'all',
            allConsumers: 'all'
        });
    };

    const fetchEntities = async () => {
        setLoading(true);
        try {
            const data = await memberService.fetchEntities();
            setEntities(data.items);
            setFilteredEntities(data.items);
        } catch (error) {
            console.error('Error fetching entities:', error);
            // Optionally show user-friendly error message
            setEntities([]);
            setFilteredEntities([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchEntities();
    }, []);

    // Auto-refresh functionality
    useEffect(() => {
        if (autoRefresh) {
            const interval = setInterval(() => {
                fetchEntities();
            }, 30000); // Refresh every 30 seconds

            return () => clearInterval(interval);
        }
    }, [autoRefresh]);

    useEffect(() => {
        let filtered = entities;
        
        // Filter by search term (name)
        if (filters.searchByName) {
            filtered = filtered.filter(entity =>
                entity.name.toLowerCase().includes(filters.searchByName!.toLowerCase())
            );
        }

        // Filter by entity type
        if (filters.entityType && filters.entityType !== 'all') {
            filtered = filtered.filter(entity => entity.entityType === filters.entityType);
        }

        // Filter by provider status
        if (filters.allProviders && filters.allProviders !== 'all') {
            if (filters.allProviders === 'providers-only') {
                filtered = filtered.filter(entity => entity.providerId);
            } else if (filters.allProviders === 'non-providers') {
                filtered = filtered.filter(entity => !entity.providerId);
            }
        }

        // Filter by consumer status
        if (filters.allConsumers && filters.allConsumers !== 'all') {
            if (filters.allConsumers === 'consumers-only') {
                filtered = filtered.filter(entity => entity.consumerId);
            } else if (filters.allConsumers === 'non-consumers') {
                filtered = filtered.filter(entity => !entity.consumerId);
            }
        }

        setFilteredEntities(filtered);
    }, [entities, filters]);

    const getEntityTypeIcon = (entityType: string) => {
        switch (entityType) {
            case 'gov':
                return <Building2 className="w-5 h-5 text-blue-500" />;
            case 'admin':
                return <Shield className="w-5 h-5 text-red-500" />;
            case 'private':
                return <User className="w-5 h-5 text-green-500" />;
            default:
                return <Users className="w-5 h-5 text-gray-500" />;
        }
    };

    const getEntityTypeBorderColor = (entityType: string) => {
        switch (entityType) {
            case 'gov':
                return 'border-l-blue-500 bg-blue-50';
            case 'admin':
                return 'border-l-red-500 bg-red-50';
            case 'private':
                return 'border-l-green-500 bg-green-50';
            default:
                return 'border-l-gray-500 bg-gray-50';
        }
    };

    const formatTimestamp = (timestamp: string) => {
        return new Date(timestamp).toLocaleString();
    };

    const entityTypes = ['gov', 'admin', 'private'];

    const handleRefresh = () => {
        fetchEntities();
    };

    const handleExport = async () => {
        try {
            const blob = await memberService.exportEntities('csv');
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `entities-${new Date().toISOString().split('T')[0]}.csv`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
        } catch (error) {
            console.error('Error exporting entities:', error);
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
                                Members
                            </h1>
                            <p className="text-lg text-gray-600">
                                Manage and monitor registered members in the system
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
                                <Users className="w-4 h-4" />
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
                        {/* Search and Entity Type Row */}
                        <div className="flex flex-col lg:flex-row gap-4">
                            <div className="flex-1">
                                <div className="relative">
                                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                                    <input
                                        type="text"
                                        placeholder="Search by entity name..."
                                        value={filters.searchByName || ''}
                                        onChange={(e) => updateFilter('searchByName', e.target.value)}
                                        className="pl-10 pr-4 py-2 w-full border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                    />
                                </div>
                            </div>
                            <div className="flex flex-col sm:flex-row gap-4">
                                <select
                                    value={filters.entityType || 'all'}
                                    onChange={(e) => updateFilter('entityType', e.target.value as 'all' | 'gov' | 'admin' | 'private')}
                                    className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                >
                                    <option value="all">All Entity Types</option>
                                    {entityTypes.map(type => (
                                        <option key={type} value={type}>
                                            {type.charAt(0).toUpperCase() + type.slice(1)}
                                        </option>
                                    ))}
                                </select>
                            </div>
                        </div>
                        
                        {/* Provider and Consumer Filters Row */}
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <select
                                value={filters.allProviders || 'all'}
                                onChange={(e) => updateFilter('allProviders', e.target.value as 'all' | 'providers-only' | 'non-providers')}
                                className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            >
                                <option value="all">All Provider Status</option>
                                <option value="providers-only">Providers Only</option>
                                <option value="non-providers">Non-Providers Only</option>
                            </select>
                            <select
                                value={filters.allConsumers || 'all'}
                                onChange={(e) => updateFilter('allConsumers', e.target.value as 'all' | 'consumers-only' | 'non-consumers')}
                                className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            >
                                <option value="all">All Consumer Status</option>
                                <option value="consumers-only">Consumers Only</option>
                                <option value="non-consumers">Non-Consumers Only</option>
                            </select>
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

                {/* Entity Statistics */}
                <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
                    {/* Total Entities */}
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Filtered Entities</p>
                                <p className="text-2xl font-bold text-gray-900">{filteredEntities.length}</p>
                                <p className="text-xs text-gray-500">of {entities.length} total</p>
                            </div>
                            <div className="p-3 rounded-full bg-gray-100">
                                <Users className="w-5 h-5 text-blue-500" />
                            </div>
                        </div>
                    </div>
                    
                    {entityTypes.map(type => {
                        const count = filteredEntities.filter(entity => entity.entityType === type).length;
                        const percentage = filteredEntities.length > 0 ? (count / filteredEntities.length * 100) : 0;
                        
                        return (
                            <div key={type} className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                                <div className="flex items-center justify-between">
                                    <div>
                                        <p className="text-sm font-medium text-gray-600 capitalize">{type} Entities</p>
                                        <p className="text-2xl font-bold text-gray-900">{count}</p>
                                        <p className="text-xs text-gray-500">{percentage.toFixed(1)}% of filtered</p>
                                    </div>
                                    <div className="p-3 rounded-full bg-gray-100">
                                        {getEntityTypeIcon(type)}
                                    </div>
                                </div>
                            </div>
                        );
                    })}
                </div>

                {/* Entities List */}
                <div className="bg-white rounded-xl shadow-lg overflow-hidden">
                    <div className="bg-gradient-to-r from-gray-600 to-gray-700 px-6 py-4">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center">
                                <Users className="w-6 h-6 text-white mr-3" />
                                <h2 className="text-xl font-semibold text-white">
                                    Entity Members ({filteredEntities.length})
                                </h2>
                            </div>
                            <div className="text-sm text-gray-200">
                                Last updated: {new Date().toLocaleTimeString()}
                            </div>
                        </div>
                    </div>
                    <div className="max-h-96 overflow-y-auto">
                        {filteredEntities.length === 0 ? (
                            <div className="text-center py-12">
                                <Users className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                                <p className="text-gray-500 text-lg">
                                    {filters.searchByName || filters.entityType !== 'all' || filters.allProviders !== 'all' || filters.allConsumers !== 'all'
                                        ? 'No entities match your filters' 
                                        : 'No entities available'
                                    }
                                </p>
                            </div>
                        ) : (
                            <div className="divide-y divide-gray-200">
                                {filteredEntities.map((entity) => (
                                    <div key={entity.entityId} className={`p-4 border-l-4 hover:bg-gray-50 transition-colors ${getEntityTypeBorderColor(entity.entityType)}`}>
                                        <div className="flex items-start space-x-3">
                                            <div className="flex-shrink-0 mt-1">
                                                {getEntityTypeIcon(entity.entityType)}
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <div className="flex items-center justify-between mb-2">
                                                    <h3 className="text-lg font-semibold text-gray-900">
                                                        {entity.name}
                                                    </h3>
                                                    <div className="flex items-center text-xs text-gray-500">
                                                        <Calendar className="w-3 h-3 mr-1" />
                                                        Created: {formatTimestamp(entity.createdAt)}
                                                    </div>
                                                </div>
                                                
                                                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mb-3">
                                                    <div className="flex items-center text-sm text-gray-600">
                                                        <Mail className="w-4 h-4 mr-2 text-gray-400" />
                                                        <span>{entity.email}</span>
                                                    </div>
                                                    <div className="flex items-center text-sm text-gray-600">
                                                        <Phone className="w-4 h-4 mr-2 text-gray-400" />
                                                        <span>{entity.phoneNumber}</span>
                                                    </div>
                                                </div>

                                                <div className="flex items-center gap-2 mb-2">
                                                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                                        entity.entityType === 'gov' 
                                                            ? 'bg-blue-100 text-blue-800' 
                                                            : entity.entityType === 'admin'
                                                            ? 'bg-red-100 text-red-800'
                                                            : 'bg-green-100 text-green-800'
                                                    }`}>
                                                        {entity.entityType.charAt(0).toUpperCase() + entity.entityType.slice(1)}
                                                    </span>
                                                    
                                                    {entity.providerId && (
                                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                                                            Provider
                                                        </span>
                                                    )}
                                                    
                                                    {entity.consumerId && (
                                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-orange-100 text-orange-800">
                                                            Consumer
                                                        </span>
                                                    )}
                                                </div>

                                                <div className="text-xs text-gray-500 space-y-1">
                                                    <p><span className="font-medium">Entity ID:</span> {entity.entityId}</p>
                                                    {entity.consumerId && (
                                                        <p><span className="font-medium">Consumer ID:</span> {entity.consumerId}</p>
                                                    )}
                                                    {entity.providerId && (
                                                        <p><span className="font-medium">Provider ID:</span> {entity.providerId}</p>
                                                    )}
                                                    <p><span className="font-medium">Last updated:</span> {formatTimestamp(entity.updatedAt)}</p>
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
