import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, FileText, Settings, CheckCircle } from 'lucide-react';
import { GraphQLSchemaExplorer } from '../components/GraphQLSchemaExplorer';
import { ApplicationService } from '../services/applicationService';
import { RegistrationSuccess } from '../components/RegistrationSuccess';
import type { ApplicationRegistration as ApplicationRegistrationData, SelectedField } from '../types/applications';

interface ApplicationRegistrationProps {
    memberId: string;
}

// Sample SDL for demonstration purposes
const sampleSDL = `directive @deprecated(
    reason: String = "No longer supported"
) on FIELD_DEFINITION | ENUM_VALUE

directive @sourceInfo(
    providerKey: String!
    schemaId: String!
    providerField: String!
) on FIELD_DEFINITION

directive @sourceInfoArgList(
    providerArgs: [SourceInfoInput!]
) on ARGUMENT_DEFINITION

type Query {
    personInfo(nic: String!): PersonInfo
    vehicle: VehicleInfo
}

input SourceInfoInput {
    providerKey: String!
    providerField: String!
}

type PersonInfo {
    fullName: String @sourceInfo(providerKey: "drp", schemaId: "drp-schema-v1", providerField: "person.fullName")
    name: String @sourceInfo(providerKey: "rgd", schemaId: "abc-212", providerField: "getPersonInfo.name")
    otherNames: String @sourceInfo(providerKey: "drp", schemaId: "drp-schema-v1", providerField: "person.otherNames")
    address: String @sourceInfo(providerKey: "drp", schemaId: "drp-schema-v1", providerField: "person.permanentAddress")
    profession: String @sourceInfo(providerKey: "drp", schemaId: "drp-schema-v1", providerField: "person.profession")
    dateOfBirth: String @sourceInfo(providerKey: "rgd", schemaId: "abc-212", providerField: "getPersonInfo.birthDate")
    sex: String @sourceInfo(providerKey: "rgd", schemaId: "abc-212", providerField: "getPersonInfo.sex")
    birthInfo: BirthInfo
    ownedVehicles: [VehicleInfo] @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicles")
}

type VehicleInfo {
    regNo: String @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.registrationNumber")
    make: String @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.make")
    model: String @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.model")
    year: Int @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.yearOfManufacture")
    class: [VehicleClass] @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.classes")
}

type VehicleClass {
    className: String @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.classes.className")
    classCode: String @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.classes.classCode")
}

type BirthInfo {
    birthRegistrationNumber: String @sourceInfo(providerKey: "rgd", schemaId: "abc-212", providerField: "getPersonInfo.brNo")
    birthPlace: String @sourceInfo(providerKey: "rgd", schemaId: "abc-212", providerField: "getPersonInfo.birthPlace")
    district: String @sourceInfo(providerKey: "rgd", schemaId: "abc-212", providerField: "getPersonInfo.district")
}`;


export const ApplicationRegistration: React.FC<ApplicationRegistrationProps> = ({ 
    memberId
}) => {
    const navigate = useNavigate();
    const [applicationName, setApplicationName] = useState('');
    const [description, setDescription] = useState('');
    const [selectedFields, setSelectedFields] = useState<SelectedField[]>([]);
    const [showSuccess, setShowSuccess] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<string>('');

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        
        if (isSubmitting) return;
        
        setIsSubmitting(true);
        setError('');
        
        try {
            const applicationData: ApplicationRegistrationData = {
                applicationName: applicationName,
                applicationDescription: description,
                selectedFields: selectedFields,
                memberId: memberId,
            };

            await ApplicationService.registerApplication(applicationData);

            // Show success page on successful registration
            setShowSuccess(true);
        } catch (error) {
            console.error('Error registering application:', error);
            const errorMessage = error instanceof Error ? error.message : 'Failed to register application';
            setError(errorMessage);
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleBack = () => {
        navigate('/applications');
    };

    const handleSuccessRedirect = () => {
        navigate('/applications');
    };

    // Show success page after successful registration
    if (showSuccess) {
        return (
            <RegistrationSuccess 
                type="application"
                title={applicationName}
                onRedirect={handleSuccessRedirect}
            />
        );
    }

    return (
        <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
            <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                {/* Header Section */}
                <div className="mb-8">
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between">
                        <div className="mb-4 sm:mb-0">
                            <button
                                onClick={handleBack}
                                className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4 transition-colors"
                            >
                                <ArrowLeft className="w-4 h-4 mr-2" />
                                Back to Applications
                            </button>
                            <h1 className="text-4xl font-bold text-gray-900 mb-2">
                                Register New Application
                            </h1>
                            <p className="text-lg text-gray-600">
                                Configure your application's access to government data services
                            </p>
                        </div>
                        <div className="hidden sm:flex items-center space-x-2 text-sm text-gray-500">
                            <div className="flex items-center">
                                <FileText className="w-4 h-4 mr-1" />
                                Step 1 of 1
                            </div>
                        </div>
                    </div>
                </div>

                {/* Error Alert */}
                {error && (
                    <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
                        <div className="flex items-center">
                            <div className="flex-shrink-0">
                                <svg className="w-5 h-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
                                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                                </svg>
                            </div>
                            <div className="ml-3">
                                <h3 className="text-sm font-medium text-red-800">Registration Error</h3>
                                <div className="mt-2 text-sm text-red-700">{error}</div>
                            </div>
                        </div>
                    </div>
                )}
                
                <form onSubmit={handleSubmit} className="space-y-8">
                    {/* Application Details Section */}
                    <div className="bg-white shadow-xl rounded-2xl overflow-hidden">
                        <div className="bg-gradient-to-r from-blue-600 to-blue-700 px-6 py-4">
                            <div className="flex items-center">
                                <Settings className="w-6 h-6 text-white mr-3" />
                                <h2 className="text-xl font-semibold text-white">Application Details</h2>
                            </div>
                        </div>
                        <div className="p-6 sm:p-8">
                            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                                <div className="lg:col-span-1">
                                    <label className="block text-sm font-medium text-gray-700 mb-2" htmlFor="applicationName">
                                        Application Name *
                                    </label>
                                    <input
                                        type="text"
                                        id="applicationName"
                                        value={applicationName}
                                        onChange={(e) => setApplicationName(e.target.value)}
                                        placeholder="Enter a descriptive name for your application"
                                        className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 text-gray-900 placeholder-gray-500"
                                        required
                                    />
                                    <p className="mt-2 text-sm text-gray-500">
                                        Choose a clear, descriptive name that identifies your application's purpose
                                    </p>
                                </div>
                                <div className="lg:col-span-2">
                                    <label className="block text-sm font-medium text-gray-700 mb-2" htmlFor="description">
                                        Description
                                    </label>
                                    <textarea
                                        id="description"
                                        value={description}
                                        onChange={(e) => setDescription(e.target.value)}
                                        placeholder="Describe what your application does and how it will use the data..."
                                        className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 text-gray-900 placeholder-gray-500 resize-none"
                                        rows={4}
                                    />
                                    <p className="mt-2 text-sm text-gray-500">
                                        Provide details about your application's functionality and data usage
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* GraphQL Schema Explorer Section */}
                    <div className="bg-white shadow-xl rounded-2xl overflow-hidden">
                        <div className="bg-gradient-to-r from-green-600 to-green-700 px-6 py-4">
                            <div className="flex items-center">
                                <CheckCircle className="w-6 h-6 text-white mr-3" />
                                <h2 className="text-xl font-semibold text-white">Data Access Configuration</h2>
                            </div>
                        </div>
                        <div className="p-6 sm:p-8">
                            <div className="mb-6">
                                <h3 className="text-lg font-medium text-gray-900 mb-2">Required Fields Selection</h3>
                                <p className="text-gray-600">
                                    Select only the data fields your application needs. This helps maintain data privacy and improves performance.
                                </p>
                            </div>
                            <GraphQLSchemaExplorer 
                                sdl={sampleSDL} // TODO: Replace with actual SDL fetch call from OrchestrationEngine
                                onSelectionChange={setSelectedFields}
                            />
                            
                            {/* Selection Summary */}
                            {selectedFields.length > 0 && (
                                <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
                                    <h4 className="text-sm font-medium text-blue-900 mb-2">Access Summary</h4>
                                    <p className="text-sm text-blue-700">
                                        Your application will have access to <strong>{selectedFields.length}</strong> data field{selectedFields.length !== 1 ? 's' : ''}.
                                        This includes personal information, vehicle data, and birth records as selected.
                                    </p>
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Submit Section */}
                    <div className="bg-white shadow-xl rounded-2xl p-6 sm:p-8">
                        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-4 sm:space-y-0">
                            <div>
                                <h3 className="text-lg font-medium text-gray-900 mb-1">Ready to Register?</h3>
                                <p className="text-gray-600">
                                    Review your configuration and submit your application for approval.
                                </p>
                            </div>
                            <div className="flex flex-col sm:flex-row space-y-3 sm:space-y-0 sm:space-x-4">
                                <button
                                    type="button"
                                    onClick={handleBack}
                                    className="px-6 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors font-medium focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    disabled={!applicationName.trim() || selectedFields.length === 0 || isSubmitting}
                                    className="px-8 py-3 bg-gradient-to-r from-blue-600 to-blue-700 text-white rounded-lg hover:from-blue-700 hover:to-blue-800 disabled:from-gray-400 disabled:to-gray-500 disabled:cursor-not-allowed transition-all duration-200 font-semibold shadow-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                                >
                                    {isSubmitting ? 'Registering...' : 'Register Application'}
                                </button>
                            </div>
                        </div>
                        
                        {(!applicationName.trim() || selectedFields.length === 0) && (
                            <div className="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
                                <p className="text-sm text-yellow-800">
                                    <strong>Required:</strong> Please provide an application name and select at least one data field to proceed.
                                </p>
                            </div>
                        )}
                    </div>
                </form>
            </div>
        </div>
    );
}