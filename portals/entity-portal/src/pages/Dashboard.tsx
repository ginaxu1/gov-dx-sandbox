import React from 'react';
import { 
    BarChart3, 
    Users, 
    Database, 
    Activity, 
    TrendingUp, 
    FileText, 
    Clock, 
    CheckCircle,
    AlertTriangle,
    ArrowUpRight
} from 'lucide-react';

export const Dashboard: React.FC = () => {
    // Mock data for demonstration
    const stats = {
        totalRequests: 12543,
        activeSchemas: 8,
        registeredApps: 15,
        responseTime: 120
    };

    const recentActivity = [
        { id: 1, action: 'Schema registered', item: 'Government Identity Schema', time: '2 hours ago', status: 'success' },
        { id: 2, action: 'Application approved', item: 'Healthcare Management System', time: '4 hours ago', status: 'success' },
        { id: 3, action: 'API request spike', item: 'Vehicle Registration API', time: '6 hours ago', status: 'warning' },
        { id: 4, action: 'New application', item: 'Tax Management Portal', time: '1 day ago', status: 'pending' },
    ];

    const topApis = [
        { name: 'Personal Information API', requests: 4521, growth: 12 },
        { name: 'Vehicle Registration API', requests: 3247, growth: 8 },
        { name: 'Healthcare Records API', requests: 2891, growth: -3 },
        { name: 'Education Credentials API', requests: 1884, growth: 15 },
    ];

    return (
        <div className="min-h-screen bg-gradient-to-br from-gray-50 to-blue-50">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                {/* Header */}
                <div className="mb-8">
                    <h1 className="text-4xl font-bold text-gray-900 mb-2">
                        Dashboard Overview
                    </h1>
                    <p className="text-lg text-gray-600">
                        Monitor your data services and application performance
                    </p>
                </div>

                {/* Stats Grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Total API Requests</p>
                                <p className="text-3xl font-bold text-gray-900">{stats.totalRequests.toLocaleString()}</p>
                                <div className="flex items-center mt-2">
                                    <TrendingUp className="w-4 h-4 text-green-500 mr-1" />
                                    <span className="text-sm text-green-600 font-medium">+12% from last month</span>
                                </div>
                            </div>
                            <div className="p-3 bg-blue-100 rounded-full">
                                <BarChart3 className="w-8 h-8 text-blue-600" />
                            </div>
                        </div>
                    </div>

                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Active Schemas</p>
                                <p className="text-3xl font-bold text-gray-900">{stats.activeSchemas}</p>
                                <div className="flex items-center mt-2">
                                    <CheckCircle className="w-4 h-4 text-green-500 mr-1" />
                                    <span className="text-sm text-green-600 font-medium">All operational</span>
                                </div>
                            </div>
                            <div className="p-3 bg-purple-100 rounded-full">
                                <Database className="w-8 h-8 text-purple-600" />
                            </div>
                        </div>
                    </div>

                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Registered Apps</p>
                                <p className="text-3xl font-bold text-gray-900">{stats.registeredApps}</p>
                                <div className="flex items-center mt-2">
                                    <TrendingUp className="w-4 h-4 text-green-500 mr-1" />
                                    <span className="text-sm text-green-600 font-medium">+3 this week</span>
                                </div>
                            </div>
                            <div className="p-3 bg-green-100 rounded-full">
                                <Users className="w-8 h-8 text-green-600" />
                            </div>
                        </div>
                    </div>

                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Avg Response Time</p>
                                <p className="text-3xl font-bold text-gray-900">{stats.responseTime}ms</p>
                                <div className="flex items-center mt-2">
                                    <Activity className="w-4 h-4 text-green-500 mr-1" />
                                    <span className="text-sm text-green-600 font-medium">Excellent</span>
                                </div>
                            </div>
                            <div className="p-3 bg-yellow-100 rounded-full">
                                <Activity className="w-8 h-8 text-yellow-600" />
                            </div>
                        </div>
                    </div>
                </div>

                {/* Content Grid */}
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                    {/* Recent Activity */}
                    <div className="bg-white rounded-xl shadow-lg overflow-hidden">
                        <div className="bg-gradient-to-r from-gray-600 to-gray-700 px-6 py-4">
                            <div className="flex items-center">
                                <Clock className="w-6 h-6 text-white mr-3" />
                                <h2 className="text-xl font-semibold text-white">Recent Activity</h2>
                            </div>
                        </div>
                        <div className="p-6">
                            <div className="space-y-4">
                                {recentActivity.map(activity => (
                                    <div key={activity.id} className="flex items-start space-x-4 p-3 rounded-lg hover:bg-gray-50 transition-colors">
                                        <div className={`p-2 rounded-full ${
                                            activity.status === 'success' ? 'bg-green-100' :
                                            activity.status === 'warning' ? 'bg-yellow-100' :
                                            'bg-blue-100'
                                        }`}>
                                            {activity.status === 'success' && <CheckCircle className="w-4 h-4 text-green-600" />}
                                            {activity.status === 'warning' && <AlertTriangle className="w-4 h-4 text-yellow-600" />}
                                            {activity.status === 'pending' && <Clock className="w-4 h-4 text-blue-600" />}
                                        </div>
                                        <div className="flex-1 min-w-0">
                                            <p className="text-sm font-medium text-gray-900">{activity.action}</p>
                                            <p className="text-sm text-gray-600 truncate">{activity.item}</p>
                                            <p className="text-xs text-gray-500 mt-1">{activity.time}</p>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>

                    {/* Top APIs */}
                    <div className="bg-white rounded-xl shadow-lg overflow-hidden">
                        <div className="bg-gradient-to-r from-blue-600 to-blue-700 px-6 py-4">
                            <div className="flex items-center">
                                <BarChart3 className="w-6 h-6 text-white mr-3" />
                                <h2 className="text-xl font-semibold text-white">Top Performing APIs</h2>
                            </div>
                        </div>
                        <div className="p-6">
                            <div className="space-y-4">
                                {topApis.map((api, index) => (
                                    <div key={index} className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 transition-colors">
                                        <div className="flex items-center space-x-3">
                                            <div className="flex items-center justify-center w-8 h-8 bg-blue-100 rounded-full">
                                                <span className="text-sm font-bold text-blue-600">{index + 1}</span>
                                            </div>
                                            <div>
                                                <p className="font-medium text-gray-900">{api.name}</p>
                                                <p className="text-sm text-gray-600">{api.requests.toLocaleString()} requests</p>
                                            </div>
                                        </div>
                                        <div className="flex items-center space-x-2">
                                            <div className={`flex items-center ${api.growth >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                                                <ArrowUpRight className={`w-4 h-4 mr-1 ${api.growth < 0 ? 'rotate-90' : ''}`} />
                                                <span className="text-sm font-medium">{Math.abs(api.growth)}%</span>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                </div>

                {/* Quick Actions */}
                <div className="mt-8">
                    <div className="bg-white rounded-xl shadow-lg p-6">
                        <h3 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h3>
                        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                            <button className="flex items-center justify-center p-4 bg-blue-50 rounded-lg hover:bg-blue-100 transition-colors group">
                                <Database className="w-6 h-6 text-blue-600 mr-2" />
                                <span className="font-medium text-blue-700 group-hover:text-blue-800">Register Schema</span>
                            </button>
                            <button className="flex items-center justify-center p-4 bg-green-50 rounded-lg hover:bg-green-100 transition-colors group">
                                <FileText className="w-6 h-6 text-green-600 mr-2" />
                                <span className="font-medium text-green-700 group-hover:text-green-800">View Applications</span>
                            </button>
                            <button className="flex items-center justify-center p-4 bg-purple-50 rounded-lg hover:bg-purple-100 transition-colors group">
                                <BarChart3 className="w-6 h-6 text-purple-600 mr-2" />
                                <span className="font-medium text-purple-700 group-hover:text-purple-800">View Analytics</span>
                            </button>
                            <button className="flex items-center justify-center p-4 bg-yellow-50 rounded-lg hover:bg-yellow-100 transition-colors group">
                                <Activity className="w-6 h-6 text-yellow-600 mr-2" />
                                <span className="font-medium text-yellow-700 group-hover:text-yellow-800">System Logs</span>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};
