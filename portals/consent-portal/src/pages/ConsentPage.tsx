import React, {useEffect} from 'react';
import { Shield } from 'lucide-react';
import { useConsent } from '../ConsentContext';
import {useAsgardeo} from "@asgardeo/react";
import {useNavigate, useSearchParams} from "react-router-dom";

const ConsentPage: React.FC = () => {
  const { consentRecord, isSubmitting, handleConsentDecision, handleConsentFetch, updateConsentId } = useConsent()
  const navigate = useNavigate()

  const { signIn, isLoading, isSignedIn } = useAsgardeo();

  const [searchParams] = useSearchParams()

  useEffect(() => {
    // get the consentId from the query params
    const consentId = searchParams.get('consent_id');

    if (consentId) {
      updateConsentId(consentId);
    }

    if (!consentRecord && !isLoading && isSignedIn) {
      handleConsentFetch();
    }

    if (!isSignedIn && !isLoading) {
      navigate('/login');
    }
  }, [consentRecord, handleConsentFetch, isLoading, isSignedIn, searchParams, signIn, updateConsentId, navigate]);

  if (!consentRecord) {

    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center relative">
        {/*<UserHeader userName={user.userName.givenName} onSignIn={() => signIn()} onSignOut={() => signOut()} />*/}
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading consent details...</p>
        </div>
      </div>
    );
  }

  const formatFieldName = (field: string): string => {
    const lastField = field ? field.split('.').at(-1) : '';
    if (!lastField) return field;

    const words = lastField
      .replace(/([a-z])([A-Z])/g, '$1 $2')
      .split(/[_\s]+/)
      .filter(word => word.length > 0);

    return words
      .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(' ');
  };

  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleString();
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 p-4 relative">
      {/*<UserHeader userName={user} onSignIn={() => signIn()} onSignOut={() => signOut()} />*/}
      <div className="max-w-2xl mx-auto py-8">
        <div className="bg-white rounded-lg shadow-lg overflow-hidden">
          <div className="bg-indigo-600 text-white p-6">
            <div className="flex items-center">
              <Shield className="h-8 w-8 mr-3" />
              <div>
                <h1 className="text-2xl font-bold">Consent Request</h1>
                <p className="text-indigo-100">Review and approve data sharing</p>
              </div>
            </div>
          </div>

          <div className="p-6">
            <div className="mb-6 p-4 bg-blue-50 rounded-lg">
              <h3 className="text-lg font-semibold text-gray-800 mb-2">Application Request</h3>
              <p className="text-gray-600">
                <span className="font-medium">{consentRecord.app_display_name}</span> is requesting access to the following data fields:
              </p>
            </div>

            <div className="mb-6">
              <h3 className="text-lg font-semibold text-gray-800 mb-3">Data Fields</h3>
              <div className="space-y-2">
                {consentRecord.fields.map((field, index) => (
                  <div key={index} className="flex items-center p-3 bg-gray-50 border border-gray-200 rounded">
                    <div className="h-2 w-2 bg-indigo-400 rounded-full mr-3"></div>
                    <span className="text-gray-700 font-medium">{formatFieldName(field)}</span>
                  </div>
                ))}
              </div>
            </div>

            <div className="mb-6 p-4 bg-gray-50 rounded-lg">
              <h3 className="text-lg font-semibold text-gray-800 mb-3">Consent Details</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
                <div>
                  <span className="font-medium text-gray-600">Owner Name:</span>
                  <span className="ml-2 text-gray-800">{consentRecord.owner_name}</span>
                </div>
                <div>
                  <span className="font-medium text-gray-600">Owner Email:</span>
                  <span className="ml-2 text-gray-800">{consentRecord.owner_email}</span>
                </div>
                <div>
                  <span className="font-medium text-gray-600">Expires:</span>
                  <span className="ml-2 text-gray-800">{formatDate(consentRecord.expires_at)}</span>
                </div>
                <div>
                  <span className="font-medium text-gray-600">Created:</span>
                  <span className="ml-2 text-gray-800">{formatDate(consentRecord.created_at)}</span>
                </div>
              </div>
            </div>

            <div className="flex space-x-4">
              <button
                onClick={() => handleConsentDecision('rejected')}
                disabled={isSubmitting}
                className="flex-1 px-6 py-3 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:bg-red-400 transition-colors font-medium"
              >
                {isSubmitting ? 'Processing...' : 'Deny'}
              </button>
              <button
                onClick={() => handleConsentDecision('approved')}
                disabled={isSubmitting}
                className="flex-1 px-6 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:bg-green-400 transition-colors font-medium"
              >
                {isSubmitting ? 'Processing...' : 'Approve'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ConsentPage;
