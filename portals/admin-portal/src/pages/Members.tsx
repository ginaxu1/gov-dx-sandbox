import React, { useState, useEffect, useCallback } from 'react';
import {
    Search,
    RefreshCw,
    Building2,
    Mail,
    Phone,
    Calendar,
    User,
    Edit,
    ChevronDown,
    ChevronRight,
    X,
    Save,
    Plus
} from 'lucide-react';
import { MemberService } from '../services/memberService';
import type { Member } from '../services/memberService';

interface FilterOptions {
    searchByName?: string;
}

interface CreateMemberFormData {
    name: string;
    email: string;
    phoneNumber: string;
}

interface UpdateMemberFormData {
    name: string;
    phoneNumber: string;
}


export const Members: React.FC = () => {
    const [members, setMembers] = useState<Member[]>([]);
    const [filteredMembers, setFilteredMembers] = useState<Member[]>([]);
    const [filters, setFilters] = useState<FilterOptions>({
        searchByName: '',
    });
    const [loading, setLoading] = useState(true);
    const [expandedCards, setExpandedCards] = useState<Set<string>>(new Set());
    const [showAddForm, setShowAddForm] = useState(false);
    const [showEditForm, setShowEditForm] = useState(false);
    const [editingMember, setEditingMember] = useState<Member | null>(null);
    const [createMemberFormData, setCreateMemberFormData] = useState<CreateMemberFormData>({
        name: '',
        email: '',
        phoneNumber: ''
    });
    const [updateMemberFormData, setUpdateMemberFormData] = useState<UpdateMemberFormData>({
        name: '',
        phoneNumber: ''
    });

    // Helper function to update filters
    const updateFilter = <K extends keyof FilterOptions>(key: K, value: FilterOptions[K]) => {
        setFilters(prev => ({ ...prev, [key]: value }));
    };

    // Helper function to clear all filters
    const clearAllFilters = () => {
        setFilters({
            searchByName: '',
        });
    };

    const fetchMembers = useCallback(async () => {
        setLoading(true);
        try {
            const data: Member[] = await MemberService.fetchMembers();
            setMembers(data);
            setFilteredMembers(data);
        } catch (error) {
            console.error('Error fetching members:', error);
            // Optionally show user-friendly error message
            setMembers([]);
            setFilteredMembers([]);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchMembers();
    }, [fetchMembers]);

    useEffect(() => {
        let filtered = members;

        // Filter by search term (name)
        if (filters.searchByName) {
            filtered = filtered.filter(member =>
                member.name.toLowerCase().includes(filters.searchByName!.toLowerCase())
            );
        }

        setFilteredMembers(filtered);
    }, [members, filters]);

    const formatTimestamp = (timestamp: string) => {
        const date = new Date(timestamp);
        return date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    const handleRefresh = () => {
        fetchMembers();
    };

    const toggleCardExpansion = (memberId: string) => {
        setExpandedCards(prev => {
            const newSet = new Set(prev);
            if (newSet.has(memberId)) {
                newSet.delete(memberId);
            } else {
                newSet.add(memberId);
            }
            return newSet;
        });
    };

    const resetForm = () => {
        setCreateMemberFormData({
            name: '',
            email: '',
            phoneNumber: '',
        });
        setUpdateMemberFormData({
            name: '',
            phoneNumber: '',
        });
    };

    const handleAddMember = () => {
        resetForm();
        setShowAddForm(true);
        setEditingMember(null);
    };

    const handleEditMember = (member: Member) => {
        setUpdateMemberFormData({
            name: member.name,
            phoneNumber: member.phoneNumber
        });
        setEditingMember(member);
        setShowEditForm(true);
    };

    const handleCloseForm = () => {
        setShowAddForm(false);
        setShowEditForm(false);
        setEditingMember(null);
        resetForm();
    };

    const handleFormSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            if (editingMember) {
                // Update existing member
                await MemberService.updateMember(editingMember.memberId, updateMemberFormData);
            } else {
                // Create new member
                await MemberService.createMember(createMemberFormData);
            }
            await fetchMembers();
            handleCloseForm();
        } catch (error) {
            console.error('Error saving member:', error);
            alert('Error saving member. Please try again.');
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
                        <div className="flex flex-col sm:flex-row gap-3">
                            <button
                                onClick={handleRefresh}
                                className="flex items-center justify-center space-x-2 px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                            >
                                <RefreshCw className="w-4 h-4" />
                                <span>Refresh</span>
                            </button>
                            <button
                                className="flex items-center justify-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                                onClick={handleAddMember}
                            >
                                <Plus className="w-4 h-4" />
                                <span>Add Member</span>
                            </button>
                        </div>
                    </div>
                </div>

                {/* Filters */}
                <div className="bg-white rounded-xl shadow-lg p-6 mb-8">
                    <div className="space-y-4">
                        {/* Search and Member Type Row */}
                        <div className="flex flex-col lg:flex-row gap-4">
                            <div className="flex-1">
                                <div className="relative">
                                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                                    <input
                                        type="text"
                                        placeholder="Search by member name..."
                                        value={filters.searchByName || ''}
                                        onChange={(e) => updateFilter('searchByName', e.target.value)}
                                        className="pl-10 pr-4 py-2 w-full border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                    />
                                </div>
                            </div>
                        </div>
                        {/* Clear Filters Button */}
                        <div className="flex justify-end">
                            <button
                                onClick={clearAllFilters}
                                className="px-4 py-2 text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
                            >
                                Clear Filter
                            </button>
                        </div>
                    </div>
                </div>

                {/* Member Statistics */}
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
                    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                        <div className="flex items-center space-x-4">
                            <div className="p-3 bg-blue-100 rounded-full">
                                <Building2 className="w-6 h-6 text-blue-600" />
                            </div>
                            <div>
                                <p className="text-2xl font-bold text-gray-900">{members.length}</p>
                                <p className="text-gray-600">Total Members</p>
                            </div>
                        </div>
                    </div>
                    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                        <div className="flex items-center space-x-4">
                            <div className="p-3 bg-green-100 rounded-full">
                                <Search className="w-6 h-6 text-green-600" />
                            </div>
                            <div>
                                <p className="text-2xl font-bold text-gray-900">{filteredMembers.length}</p>
                                <p className="text-gray-600">{filters.searchByName ? 'Search Results' : 'Showing All'}</p>
                            </div>
                        </div>
                    </div>
                    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                        <div className="flex items-center space-x-4">
                            <div className="p-3 bg-purple-100 rounded-full">
                                <ChevronDown className="w-6 h-6 text-purple-600" />
                            </div>
                            <div>
                                <p className="text-2xl font-bold text-gray-900">{expandedCards.size}</p>
                                <p className="text-gray-600">Expanded</p>
                            </div>
                        </div>
                    </div>
                </div>


                {/* Members List */}
                <div className="bg-white rounded-xl shadow-lg overflow-hidden mb-16">
                    <div className="bg-gradient-to-r from-blue-600 to-blue-700 px-6 py-4">
                        <h2 className="text-xl font-semibold text-white flex items-center space-x-2">
                            <User className="w-5 h-5" />
                            <span>Member List</span>
                            <span className="ml-auto bg-blue-500 text-blue-100 px-3 py-1 rounded-full text-sm">
                                {filteredMembers.length} {filteredMembers.length === 1 ? 'member' : 'members'}
                            </span>
                        </h2>
                    </div>

                    <div className="divide-y divide-gray-100">
                        {filteredMembers.map((member) => {
                            const isExpanded = expandedCards.has(member.memberId);
                            return (
                                <div key={member.memberId} className="group hover:bg-gray-50 transition-all duration-200">
                                    {/* Main Card Content - Always Visible */}
                                    <div
                                        className="p-6 cursor-pointer"
                                        onClick={() => toggleCardExpansion(member.memberId)}
                                    >
                                        <div className="flex items-center justify-between">
                                            <div className="flex items-center space-x-4 flex-1">
                                                <div className="flex-shrink-0 bg-blue-50 rounded-full p-2 group-hover:bg-blue-100 transition-colors">
                                                    {isExpanded ? (
                                                        <ChevronDown className="w-4 h-4 text-blue-600" />
                                                    ) : (
                                                        <ChevronRight className="w-4 h-4 text-blue-600" />
                                                    )}
                                                </div>

                                                <div className="flex-1 min-w-0">
                                                    <div className="flex flex-col sm:flex-row sm:items-center sm:space-x-6 space-y-2 sm:space-y-0">
                                                        <div className="flex-shrink-0">
                                                            <h3 className="text-lg font-semibold text-gray-900 group-hover:text-blue-700 transition-colors">
                                                                {member.name}
                                                            </h3>
                                                        </div>
                                                        <div className="flex items-center space-x-2 text-gray-600">
                                                            <div className="bg-gray-100 rounded-full p-1">
                                                                <Mail className="w-3 h-3" />
                                                            </div>
                                                            <span className="text-sm font-medium truncate">{member.email}</span>
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>

                                            <div className="flex items-center space-x-2 ml-4">
                                                <button
                                                    onClick={(e) => {
                                                        e.stopPropagation();
                                                        handleEditMember(member);
                                                    }}
                                                    className="p-2.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-all duration-200 shadow-sm hover:shadow-md"
                                                    title="Edit member"
                                                >
                                                    <Edit className="w-4 h-4" />
                                                </button>
                                            </div>
                                        </div>
                                    </div>

                                    {/* Expanded Content */}
                                    {isExpanded && (
                                        <div className="px-6 pb-6 bg-gradient-to-br from-gray-50 to-slate-50 border-t border-gray-100">
                                            <div className="pt-4">
                                                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                                                    <div className="space-y-4">
                                                        <div className="bg-white rounded-lg p-4 shadow-sm">
                                                            <div className="flex items-start space-x-3">
                                                                <div className="bg-blue-100 rounded-full p-2">
                                                                    <User className="w-4 h-4 text-blue-600" />
                                                                </div>
                                                                <div className="flex-1 min-w-0">
                                                                    <p className="text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">
                                                                        Member ID
                                                                    </p>
                                                                    <p className="text-sm font-semibold text-gray-900 break-all">
                                                                        {member.memberId}
                                                                    </p>
                                                                </div>
                                                            </div>
                                                        </div>

                                                        <div className="bg-white rounded-lg p-4 shadow-sm">
                                                            <div className="flex items-start space-x-3">
                                                                <div className="bg-green-100 rounded-full p-2">
                                                                    <User className="w-4 h-4 text-green-600" />
                                                                </div>
                                                                <div className="flex-1 min-w-0">
                                                                    <p className="text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">
                                                                        IDP User ID
                                                                    </p>
                                                                    <p className="text-sm font-semibold text-gray-900 break-all">
                                                                        {member.idpUserId}
                                                                    </p>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <div className="space-y-4">
                                                        <div className="bg-white rounded-lg p-4 shadow-sm">
                                                            <div className="flex items-start space-x-3">
                                                                <div className="bg-purple-100 rounded-full p-2">
                                                                    <Phone className="w-4 h-4 text-purple-600" />
                                                                </div>
                                                                <div className="flex-1 min-w-0">
                                                                    <p className="text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">
                                                                        Phone Number
                                                                    </p>
                                                                    <p className="text-sm font-semibold text-gray-900">
                                                                        {member.phoneNumber}
                                                                    </p>
                                                                </div>
                                                            </div>
                                                        </div>

                                                        <div className="bg-white rounded-lg p-4 shadow-sm">
                                                            <div className="flex items-start space-x-3">
                                                                <div className="bg-orange-100 rounded-full p-2">
                                                                    <Calendar className="w-4 h-4 text-orange-600" />
                                                                </div>
                                                                <div className="flex-1 min-w-0">
                                                                    <p className="text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">
                                                                        Member Since
                                                                    </p>
                                                                    <p className="text-sm font-semibold text-gray-900">
                                                                        {formatTimestamp(member.createdAt)}
                                                                    </p>
                                                                    <p className="text-xs text-gray-500 mt-1">
                                                                        Updated: {formatTimestamp(member.updatedAt)}
                                                                    </p>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                    </div>

                    {filteredMembers.length === 0 && (
                        <div className="text-center py-16 px-6">
                            <div className="max-w-sm mx-auto">
                                <div className="bg-gray-100 rounded-full w-20 h-20 flex items-center justify-center mx-auto mb-6">
                                    <Building2 className="w-10 h-10 text-gray-400" />
                                </div>
                                <h3 className="text-lg font-semibold text-gray-900 mb-2">No members found</h3>
                                <p className="text-gray-600 mb-4">
                                    {filters.searchByName
                                        ? "No members match your search criteria. Try adjusting your search terms."
                                        : "Get started by adding your first member to the system."
                                    }
                                </p>
                                <button
                                    onClick={handleAddMember}
                                    className="inline-flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                                >
                                    <Plus className="w-4 h-4" />
                                    <span>Add First Member</span>
                                </button>
                            </div>
                        </div>
                    )}
                </div>

                {/* Add Member Form Modal */}
                {showAddForm && (
                    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
                        <div className="bg-white rounded-lg shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto">
                            <div className="p-6">
                                <div className="flex items-center justify-between mb-6">
                                    <h2 className="text-xl font-semibold text-gray-900">
                                        {editingMember ? 'Edit Member' : 'Add New Member'}
                                    </h2>
                                    <button
                                        onClick={handleCloseForm}
                                        className="text-gray-400 hover:text-gray-600 transition-colors"
                                    >
                                        <X className="w-6 h-6" />
                                    </button>
                                </div>

                                <form onSubmit={handleFormSubmit} className="space-y-4">
                                    <div>
                                        <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                                            Name *
                                        </label>
                                        <input
                                            type="text"
                                            id="name"
                                            required
                                            value={createMemberFormData.name}
                                            onChange={(e) => setCreateMemberFormData(prev => ({ ...prev, name: e.target.value }))}
                                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                            placeholder="Enter member name"
                                        />
                                    </div>

                                    <div>
                                        <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
                                            Email *
                                        </label>
                                        <input
                                            type="email"
                                            id="email"
                                            required
                                            value={createMemberFormData.email}
                                            onChange={(e) => setCreateMemberFormData(prev => ({ ...prev, email: e.target.value }))}
                                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                            placeholder="Enter email address"
                                        />
                                    </div>

                                    <div>
                                        <label htmlFor="phoneNumber" className="block text-sm font-medium text-gray-700 mb-1">
                                            Phone Number *
                                        </label>
                                        <input
                                            type="tel"
                                            id="phoneNumber"
                                            required
                                            value={createMemberFormData.phoneNumber}
                                            onChange={(e) => setCreateMemberFormData(prev => ({ ...prev, phoneNumber: e.target.value }))}
                                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                            placeholder="Enter phone number"
                                        />
                                    </div>
                                    <div className="flex justify-end space-x-3 pt-4">
                                        <button
                                            type="button"
                                            onClick={handleCloseForm}
                                            className="px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
                                        >
                                            Cancel
                                        </button>
                                        <button
                                            type="submit"
                                            className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                                        >
                                            <Save className="w-4 h-4" />
                                            <span>Create Member</span>
                                        </button>
                                    </div>
                                </form>
                            </div>
                        </div>
                    </div>
                )}

                {/* Edit Member Form Modal */}
                {showEditForm && (
                    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
                        <div className="bg-white rounded-lg shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto">
                            <div className="p-6">
                                <div className="flex items-center justify-between mb-6">
                                    <h2 className="text-xl font-semibold text-gray-900">
                                        Edit Member
                                    </h2>
                                    <button
                                        onClick={handleCloseForm}
                                        className="text-gray-400 hover:text-gray-600 transition-colors"
                                    >
                                        <X className="w-6 h-6" />
                                    </button>
                                </div>

                                <form onSubmit={handleFormSubmit} className="space-y-4">
                                    <div>
                                        <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                                            Name *
                                        </label>
                                        <input
                                            type="text"
                                            id="name"
                                            required
                                            value={updateMemberFormData.name}
                                            onChange={(e) => setUpdateMemberFormData(prev => ({ ...prev, name: e.target.value }))}
                                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                            placeholder="Enter member name"
                                        />
                                    </div>

                                    <div>
                                        <label htmlFor="phoneNumber" className="block text-sm font-medium text-gray-700 mb-1">
                                            Phone Number *
                                        </label>
                                        <input
                                            type="tel"
                                            id="phoneNumber"
                                            required
                                            value={updateMemberFormData.phoneNumber}
                                            onChange={(e) => setUpdateMemberFormData(prev => ({ ...prev, phoneNumber: e.target.value }))}
                                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                            placeholder="Enter phone number"
                                        />
                                    </div>
                                    <div className="flex justify-end space-x-3 pt-4">
                                        <button
                                            type="button"
                                            onClick={handleCloseForm}
                                            className="px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
                                        >
                                            Cancel
                                        </button>
                                        <button
                                            type="submit"
                                            className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                                        >
                                            <Save className="w-4 h-4" />
                                            <span>Update Member</span>
                                        </button>
                                    </div>
                                </form>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};