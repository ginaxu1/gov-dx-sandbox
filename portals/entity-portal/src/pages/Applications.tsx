import React, { useEffect, useState } from "react";
import { useNavigate } from 'react-router-dom';
import { FileText, Plus, Clock, CheckCircle, Search } from 'lucide-react';

interface ApplicationProps {
    id: number;
    name: string;
}

export const ApplicationsPage: React.FC = () => {
    const navigate = useNavigate();
    const [registeredApplications, setRegisteredApplications] = useState<ApplicationProps[]>([]);
    const [pendingApplications, setPendingApplications] = useState<ApplicationProps[]>([]);
    const [searchTerm, setSearchTerm] = useState('');
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchApplications = async () => {
            try {
                setLoading(true);
                // Simulate API call delay
                await new Promise(resolve => setTimeout(resolve, 1000));
                
                const fetchedRegisteredApplications = [
                    { id: 1, name: 'Government Data Portal' },
                    { id: 2, name: 'Citizen Services App' },
                    { id: 3, name: 'Healthcare Management System' },
                ];
                const fetchedPendingApplications = [
                    { id: 4, name: 'Education Portal' },
                    { id: 5, name: 'Tax Management System' },
                ];

                setRegisteredApplications(fetchedRegisteredApplications);
                setPendingApplications(fetchedPendingApplications);
            } catch (error) {
                console.error('Error fetching applications:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchApplications();
    }, []);

    const handleCreateNewApplication = () => {
        navigate('/consumer/applications/new');
    };

    const filteredRegisteredApps = registeredApplications.filter(app =>
        app.name.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const filteredPendingApps = pendingApplications.filter(app =>
        app.name.toLowerCase().includes(searchTerm.toLowerCase())
    );

    if (loading) {
        return (
            <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
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
        <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                {/* Header Section */}
                <div className="mb-8">
                    <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between">
                        <div className="mb-6 lg:mb-0">
                            <h1 className="text-4xl font-bold text-gray-900 mb-2">
                                My Applications
                            </h1>
                            <p className="text-lg text-gray-600">
                                Manage your registered applications and track pending requests
                            </p>
                        </div>
                        <div className="flex flex-col sm:flex-row gap-4">
                            <div className="relative">
                                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                                <input
                                    type="text"
                                    placeholder="Search applications..."
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    className="pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 w-full sm:w-64"
                                />
                            </div>
                            <button
                                onClick={handleCreateNewApplication}
                                className="bg-gradient-to-r from-blue-600 to-blue-700 text-white px-6 py-2 rounded-lg hover:from-blue-700 hover:to-blue-800 transition-all duration-200 font-medium shadow-lg hover:shadow-xl flex items-center space-x-2"
                            >
                                <Plus className="w-5 h-5" />
                                <span>New Application</span>
                            </button>
                        </div>
                    </div>
                </div>

                {/* Statistics Cards */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center">
                            <div className="p-3 bg-green-100 rounded-full">
                                <CheckCircle className="w-6 h-6 text-green-600" />
                            </div>
                            <div className="ml-4">
                                <p className="text-sm font-medium text-gray-600">Active Applications</p>
                                <p className="text-2xl font-bold text-gray-900">{registeredApplications.length}</p>
                            </div>
                        </div>
                    </div>
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center">
                            <div className="p-3 bg-yellow-100 rounded-full">
                                <Clock className="w-6 h-6 text-yellow-600" />
                            </div>
                            <div className="ml-4">
                                <p className="text-sm font-medium text-gray-600">Pending Approval</p>
                                <p className="text-2xl font-bold text-gray-900">{pendingApplications.length}</p>
                            </div>
                        </div>
                    </div>
                    <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
                        <div className="flex items-center">
                            <div className="p-3 bg-blue-100 rounded-full">
                                <FileText className="w-6 h-6 text-blue-600" />
                            </div>
                            <div className="ml-4">
                                <p className="text-sm font-medium text-gray-600">Total Applications</p>
                                <p className="text-2xl font-bold text-gray-900">{registeredApplications.length + pendingApplications.length}</p>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Registered Applications */}
                <div className="bg-white rounded-xl shadow-lg mb-8 overflow-hidden">
                    <div className="bg-gradient-to-r from-green-600 to-green-700 px-6 py-4">
                        <div className="flex items-center">
                            <CheckCircle className="w-6 h-6 text-white mr-3" />
                            <h2 className="text-xl font-semibold text-white">Active Applications</h2>
                        </div>
                    </div>
                    <div className="p-6">
                        {filteredRegisteredApps.length === 0 ? (
                            <div className="text-center py-12">
                                <FileText className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                                <p className="text-gray-500 text-lg">
                                    {searchTerm ? 'No applications match your search' : 'No registered applications yet'}
                                </p>
                                {!searchTerm && (
                                    <button
                                        onClick={handleCreateNewApplication}
                                        className="mt-4 text-blue-600 hover:text-blue-700 font-medium"
                                    >
                                        Register your first application
                                    </button>
                                )}
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                                {filteredRegisteredApps.map(application => (
                                    <div key={application.id} className="bg-gray-50 rounded-lg p-4 hover:bg-gray-100 transition-colors border border-gray-200">
                                        <div className="flex items-start justify-between">
                                            <div className="flex-1">
                                                <h3 className="font-semibold text-gray-900 mb-2">{application.name}</h3>
                                                <div className="flex items-center text-sm text-green-600">
                                                    <CheckCircle className="w-4 h-4 mr-1" />
                                                    Active
                                                </div>
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

                {/* Pending Applications */}
                <div className="bg-white rounded-xl shadow-lg overflow-hidden">
                    <div className="bg-gradient-to-r from-yellow-600 to-yellow-700 px-6 py-4">
                        <div className="flex items-center">
                            <Clock className="w-6 h-6 text-white mr-3" />
                            <h2 className="text-xl font-semibold text-white">Pending Approval</h2>
                        </div>
                    </div>
                    <div className="p-6">
                        {filteredPendingApps.length === 0 ? (
                            <div className="text-center py-12">
                                <Clock className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                                <p className="text-gray-500 text-lg">
                                    {searchTerm ? 'No pending applications match your search' : 'No pending applications'}
                                </p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                                {filteredPendingApps.map(application => (
                                    <div key={application.id} className="bg-gray-50 rounded-lg p-4 hover:bg-gray-100 transition-colors border border-gray-200">
                                        <div className="flex items-start justify-between">
                                            <div className="flex-1">
                                                <h3 className="font-semibold text-gray-900 mb-2">{application.name}</h3>
                                                <div className="flex items-center text-sm text-yellow-600">
                                                    <Clock className="w-4 h-4 mr-1" />
                                                    Pending Review
                                                </div>
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