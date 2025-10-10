import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, FileText, Settings, CheckCircle } from 'lucide-react';
import { GraphQLSchemaExplorer } from '../components/GraphQLSchemaExplorer';
import { ApplicationService } from '../services/applicationService';
import { RegistrationSuccess } from '../components/RegistrationSuccess';
import type { ApplicationRegistration as ApplicationRegistrationData } from '../types/applications';

interface ApplicationRegistrationProps {
    consumerId: string;
}

// Sample SDL for demonstration purposes
const sampleSDL = `directive @deprecated(
    reason: String = "No longer supported"
) on FIELD_DEFINITION | ENUM_VALUE

directive @sourceInfo(
    providerKey: String!
    providerField: String!
) on FIELD_DEFINITION

directive @sourceInfoArgList(
    providerArgs: [SourceInfoInput!]
) on ARGUMENT_DEFINITION

directive @description(
    text: String!
) on FIELD_DEFINITION

type Query {
    personInfo(nic: String!): PersonInfo @description(text: "Retrieve comprehensive personal information using National Identity Card number")
    vehicle: VehicleInfo @description(text: "Get vehicle information and registration details")
}

input SourceInfoInput {
    providerKey: String!
    providerField: String!
}

type PersonInfo {
    fullName: String @sourceInfo(providerKey: "drp", providerField: "person.fullName") @description(text: "Complete full name as registered in official documents")
    name: String @sourceInfo(providerKey: "rgd", providerField: "getPersonInfo.name") @description(text: "Primary name used for identification")
    otherNames: String @sourceInfo(providerKey: "drp", providerField: "person.otherNames") @description(text: "Alternative names or aliases")
    address: String @sourceInfo(providerKey: "drp", providerField: "person.permanentAddress") @description(text: "Current permanent residential address")
    profession: String @sourceInfo(providerKey: "drp", providerField: "person.profession")
    dateOfBirth: String @sourceInfo(providerKey: "rgd", providerField: "getPersonInfo.birthDate") @description(text: "Date of birth in YYYY-MM-DD format")
    sex: String @sourceInfo(providerKey: "rgd", providerField: "getPersonInfo.sex") @description(text: "Gender as recorded in birth certificate")
    birthInfo: BirthInfo @description(text: "Detailed birth registration information")
    ownedVehicles: [VehicleInfo] @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data") @description(text: "List of all vehicles registered under this person")
}

type VehicleInfo {
    regNo: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.registrationNumber") @description(text: "Official vehicle registration number")
    make: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.make") @description(text: "Vehicle manufacturer brand")
    model: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.model") @description(text: "Specific model name of the vehicle")
    year: Int @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.yearOfManufacture") @description(text: "Year the vehicle was manufactured")
}

type BirthInfo {
    birthRegistrationNumber: String @sourceInfo(providerKey: "rgd", providerField: "getPersonInfo.brNo") @description(text: "Unique birth certificate registration number")
    birthPlace: String @sourceInfo(providerKey: "rgd", providerField: "getPersonInfo.birthPlace") @description(text: "Location where the person was born")
    district: String @sourceInfo(providerKey: "rgd", providerField: "getPersonInfo.district") @description(text: "Administrative district of birth")
}`;


export const ApplicationRegistration: React.FC<ApplicationRegistrationProps> = ({ 
    consumerId
}) => {
    const navigate = useNavigate();
    const [applicationName, setApplicationName] = useState('');
    const [description, setDescription] = useState('');
    const [selectedFields, setSelectedFields] = useState<string[]>([]);
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
            };
            
            await ApplicationService.registerApplication(consumerId, applicationData);
            
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
        navigate('/consumer/applications');
    };

    const handleSuccessRedirect = () => {
        navigate('/consumer/applications');
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
                                {/* <div className="lg:col-span-1">
                                    <label className="block text-sm font-medium text-gray-700 mb-2" htmlFor="consumerId">
                                        Consumer ID
                                    </label>
                                    <input
                                        type="text"
                                        id="consumerId"
                                        value={consumerId}
                                        readOnly
                                        className="w-full px-4 py-3 border border-gray-300 rounded-lg bg-gray-50 text-gray-600 cursor-not-allowed transition-all duration-200"
                                    />
                                    <p className="mt-2 text-sm text-gray-500">
                                        Your unique consumer identifier (automatically assigned)
                                    </p>
                                </div> */}
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
                                sdl={sampleSDL}
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