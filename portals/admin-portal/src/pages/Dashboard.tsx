import React from 'react';
import { 
    Users, 
    Database, 
    Activity, 
    FileText, 
    Clock, 
    CheckCircle,
    AlertTriangle,
    UserPlus,
    Plus
} from 'lucide-react';

export const Dashboard: React.FC = () => {
    // Mock data for demonstration - will be replaced with API calls later
    const stats = {
        totalMembers: 47,
        activeSchemas: 12,
        registeredApplications: 23,
        // Additional breakdown data
        membersByType: {
            government: 28,
            private: 15,
            admin: 4
        },
        schemasByCategory: {
            identity: 5,
            vehicle: 3,
            healthcare: 2,
            education: 2
        },
        applicationsByStatus: {
            active: 18,
            pending: 3,
            inactive: 2
        }
    };

    const recentActivity = [
        { id: 1, action: 'New member registered', item: 'Department of Motor Vehicles', time: '2 hours ago', status: 'success' },
        { id: 2, action: 'Schema updated', item: 'Personal Identity Schema v2.1', time: '4 hours ago', status: 'success' },
        { id: 3, action: 'Application submitted', item: 'Healthcare Records Portal', time: '6 hours ago', status: 'pending' },
        { id: 4, action: 'Member approved', item: 'National Tax Authority', time: '1 day ago', status: 'success' },
        { id: 5, action: 'Schema validation failed', item: 'Education Credentials Schema', time: '2 days ago', status: 'warning' },
    ];

    const topSchemas = [
        { name: 'Personal Identity Schema', usage: 28, members: 15 },
        { name: 'Vehicle Registration Schema', usage: 24, members: 12 },
        { name: 'Healthcare Records Schema', usage: 18, members: 8 },
        { name: 'Education Credentials Schema', usage: 14, members: 6 },
        { name: 'Business Registration Schema', usage: 10, members: 4 },
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
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                    {/* Total Members */}
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Total Members</p>
                                <p className="text-3xl font-bold text-gray-900">{stats.totalMembers}</p>
                                <div className="flex items-center mt-2">
                                    <div className="flex space-x-4 text-xs">
                                        <span className="text-blue-600">{stats.membersByType.government} Gov</span>
                                        <span className="text-green-600">{stats.membersByType.private} Private</span>
                                        <span className="text-red-600">{stats.membersByType.admin} Admin</span>
                                    </div>
                                </div>
                            </div>
                            <div className="p-3 bg-blue-100 rounded-full">
                                <Users className="w-8 h-8 text-blue-600" />
                            </div>
                        </div>
                    </div>

                    {/* Active Schemas */}
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Active Schemas</p>
                                <p className="text-3xl font-bold text-gray-900">{stats.activeSchemas}</p>
                                <div className="flex items-center mt-2">
                                    <div className="flex space-x-3 text-xs">
                                        <span className="text-purple-600">{stats.schemasByCategory.identity} Identity</span>
                                        <span className="text-orange-600">{stats.schemasByCategory.vehicle} Vehicle</span>
                                        <span className="text-teal-600">{stats.schemasByCategory.healthcare} Health</span>
                                    </div>
                                </div>
                            </div>
                            <div className="p-3 bg-purple-100 rounded-full">
                                <Database className="w-8 h-8 text-purple-600" />
                            </div>
                        </div>
                    </div>

                    {/* Registered Applications */}
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm font-medium text-gray-600">Applications</p>
                                <p className="text-3xl font-bold text-gray-900">{stats.registeredApplications}</p>
                                <div className="flex items-center mt-2">
                                    <div className="flex space-x-3 text-xs">
                                        <span className="text-green-600">{stats.applicationsByStatus.active} Active</span>
                                        <span className="text-yellow-600">{stats.applicationsByStatus.pending} Pending</span>
                                        <span className="text-gray-600">{stats.applicationsByStatus.inactive} Inactive</span>
                                    </div>
                                </div>
                            </div>
                            <div className="p-3 bg-green-100 rounded-full">
                                <FileText className="w-8 h-8 text-green-600" />
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

                    {/* Top Schemas */}
                    <div className="bg-white rounded-xl shadow-lg overflow-hidden">
                        <div className="bg-gradient-to-r from-purple-600 to-purple-700 px-6 py-4">
                            <div className="flex items-center">
                                <Database className="w-6 h-6 text-white mr-3" />
                                <h2 className="text-xl font-semibold text-white">Most Used Schemas</h2>
                            </div>
                        </div>
                        <div className="p-6">
                            <div className="space-y-4">
                                {topSchemas.map((schema, index) => (
                                    <div key={index} className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 transition-colors">
                                        <div className="flex items-center space-x-3">
                                            <div className="flex items-center justify-center w-8 h-8 bg-purple-100 rounded-full">
                                                <span className="text-sm font-bold text-purple-600">{index + 1}</span>
                                            </div>
                                            <div>
                                                <p className="font-medium text-gray-900">{schema.name}</p>
                                                <p className="text-sm text-gray-600">{schema.members} members using</p>
                                            </div>
                                        </div>
                                        <div className="flex items-center space-x-2">
                                            <div className="flex items-center text-purple-600">
                                                <Activity className="w-4 h-4 mr-1" />
                                                <span className="text-sm font-medium">{schema.usage} times</span>
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
                                <UserPlus className="w-6 h-6 text-blue-600 mr-2" />
                                <span className="font-medium text-blue-700 group-hover:text-blue-800">Add Member</span>
                            </button>
                            <button className="flex items-center justify-center p-4 bg-purple-50 rounded-lg hover:bg-purple-100 transition-colors group">
                                <Plus className="w-6 h-6 text-purple-600 mr-2" />
                                <span className="font-medium text-purple-700 group-hover:text-purple-800">Create Schema</span>
                            </button>
                            <button className="flex items-center justify-center p-4 bg-green-50 rounded-lg hover:bg-green-100 transition-colors group">
                                <FileText className="w-6 h-6 text-green-600 mr-2" />
                                <span className="font-medium text-green-700 group-hover:text-green-800">View Applications</span>
                            </button>
                            <button className="flex items-center justify-center p-4 bg-orange-50 rounded-lg hover:bg-orange-100 transition-colors group">
                                <Activity className="w-6 h-6 text-orange-600 mr-2" />
                                <span className="font-medium text-orange-700 group-hover:text-orange-800">View Logs</span>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};
