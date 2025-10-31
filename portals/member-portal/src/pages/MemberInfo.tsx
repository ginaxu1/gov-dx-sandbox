import React, { useState } from "react";

interface MemberInfoProps {
    name: string;
    email: string;
    phoneNumber: string;
    createdAt: string; // ISO date string
    updatedAt: string; // ISO date string
}

const MemberInfo: React.FC<MemberInfoProps> = ({
    name,
    email,
    phoneNumber,
    createdAt,
    updatedAt
}) => {
    const [editMode, setEditMode] = useState({
        name: false,
        email: false,
        phoneNumber: false
    });
    
    const [editValues, setEditValues] = useState({
        name,
        email,
        phoneNumber
    });

    const handleEditToggle = (field: 'name' | 'email' | 'phoneNumber') => {
        setEditMode(prev => ({
            ...prev,
            [field]: !prev[field]
        }));
        // Reset edit value to current value when canceling edit
        if (editMode[field]) {
            setEditValues(prev => ({
                ...prev,
                [field]: field === 'name' ? name : field === 'email' ? email : phoneNumber
            }));
        }
    };

    const handleSubmitEdit = (field: 'name' | 'email' | 'phoneNumber') => {
        console.log(`Changes submitted for admin approval: ${field} = ${editValues[field]}`);
        setEditMode(prev => ({
            ...prev,
            [field]: false
        }));
    };

    return (
        <div className="max-w-4xl mx-auto p-4 sm:p-6 lg:p-8">
            <div className="bg-white rounded-xl shadow-lg border border-gray-100 overflow-hidden">
                {/* Header */}
                <div className="bg-gradient-to-r from-blue-600 to-indigo-600 px-6 py-8 sm:px-8">
                    <h2 className="text-2xl sm:text-3xl font-bold text-white mb-2">Member Information</h2>
                    <p className="text-blue-100">Manage your account details and roles</p>
                </div>

                <div className="p-6 sm:p-8">
                    {/* Main Info Grid */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
                        {/* Personal Information Card */}
                        <div className="bg-gray-50 rounded-lg p-6">
                            <h3 className="text-lg font-semibold text-gray-800 mb-4 flex items-center">
                                <svg className="w-5 h-5 mr-2 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                                </svg>
                                Personal Details
                            </h3>
                            
                            {/* Name Field */}
                            <div className="mb-4">
                                <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                                <div className="flex items-center gap-3">
                                    {editMode.name ? (
                                        <div className="flex-1 flex gap-2">
                                            <input
                                                type="text"
                                                value={editValues.name}
                                                onChange={(e) => setEditValues(prev => ({ ...prev, name: e.target.value }))}
                                                className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                            />
                                            <button
                                                onClick={() => handleSubmitEdit('name')}
                                                className="px-3 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors text-sm"
                                            >
                                                Save
                                            </button>
                                            <button
                                                onClick={() => handleEditToggle('name')}
                                                className="px-3 py-2 bg-gray-500 text-white rounded-md hover:bg-gray-600 transition-colors text-sm"
                                            >
                                                Cancel
                                            </button>
                                        </div>
                                    ) : (
                                        <>
                                            <span className="flex-1 text-gray-900">{name}</span>
                                            <button
                                                onClick={() => handleEditToggle('name')}
                                                className="px-3 py-1 text-blue-600 hover:bg-blue-50 rounded-md transition-colors text-sm font-medium"
                                            >
                                                Edit
                                            </button>
                                        </>
                                    )}
                                </div>
                            </div>

                            {/* Email Field */}
                            <div className="mb-4">
                                <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
                                <div className="flex items-center gap-3">
                                    {editMode.email ? (
                                        <div className="flex-1 flex gap-2">
                                            <input
                                                type="email"
                                                value={editValues.email}
                                                onChange={(e) => setEditValues(prev => ({ ...prev, email: e.target.value }))}
                                                className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                            />
                                            <button
                                                onClick={() => handleSubmitEdit('email')}
                                                className="px-3 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors text-sm"
                                            >
                                                Save
                                            </button>
                                            <button
                                                onClick={() => handleEditToggle('email')}
                                                className="px-3 py-2 bg-gray-500 text-white rounded-md hover:bg-gray-600 transition-colors text-sm"
                                            >
                                                Cancel
                                            </button>
                                        </div>
                                    ) : (
                                        <>
                                            <span className="flex-1 text-gray-900">{email}</span>
                                            <button
                                                onClick={() => handleEditToggle('email')}
                                                className="px-3 py-1 text-blue-600 hover:bg-blue-50 rounded-md transition-colors text-sm font-medium"
                                            >
                                                Edit
                                            </button>
                                        </>
                                    )}
                                </div>
                            </div>

                            {/* Phone Number Field */}
                            <div className="mb-4">
                                <label className="block text-sm font-medium text-gray-700 mb-1">Phone Number</label>
                                <div className="flex items-center gap-3">
                                    {editMode.phoneNumber ? (
                                        <div className="flex-1 flex gap-2">
                                            <input
                                                type="tel"
                                                value={editValues.phoneNumber}
                                                onChange={(e) => setEditValues(prev => ({ ...prev, phoneNumber: e.target.value }))}
                                                className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                            />
                                            <button
                                                onClick={() => handleSubmitEdit('phoneNumber')}
                                                className="px-3 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors text-sm"
                                            >
                                                Save
                                            </button>
                                            <button
                                                onClick={() => handleEditToggle('phoneNumber')}
                                                className="px-3 py-2 bg-gray-500 text-white rounded-md hover:bg-gray-600 transition-colors text-sm"
                                            >
                                                Cancel
                                            </button>
                                        </div>
                                    ) : (
                                        <>
                                            <span className="flex-1 text-gray-900">{phoneNumber}</span>
                                            <button
                                                onClick={() => handleEditToggle('phoneNumber')}
                                                className="px-3 py-1 text-blue-600 hover:bg-blue-50 rounded-md transition-colors text-sm font-medium"
                                            >
                                                Edit
                                            </button>
                                        </>
                                    )}
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Account Information */}
                    <div className="bg-gray-50 rounded-lg p-6">
                        <h3 className="text-lg font-semibold text-gray-800 mb-4 flex items-center">
                            <svg className="w-5 h-5 mr-2 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            Account Information
                        </h3>
                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">Account Created</label>
                                <p className="text-gray-900">{new Date(createdAt).toLocaleDateString()}</p>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">Last Updated</label>
                                <p className="text-gray-900">{new Date(updatedAt).toLocaleDateString()}</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default MemberInfo;