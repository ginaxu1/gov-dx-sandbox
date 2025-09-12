import React, { useState, useEffect } from 'react';
import { Shield, Check, X, Lock, AlertCircle, CheckCircle } from 'lucide-react';

// Types
interface ConsentRecord {
  consent_uuid: string;
  owner_id: string;
  data_consumer: string;
  status: 'pending' | 'approved' | 'denied' | 'rejected';
  type?: string;
  created_at: string;
  updated_at: string;
  expires_at: string;
  fields: string[];
  session_id: string;
  redirect_url: string;
}

interface ConsentGatewayProps {}

const ConsentGateway: React.FC<ConsentGatewayProps> = () => {
  const [currentStep, setCurrentStep] = useState<'loading' | 'consent' | 'otp' | 'success' | 'error'>('loading');
  const [consentRecord, setConsentRecord] = useState<ConsentRecord | null>(null);
  const [userDecision, setUserDecision] = useState<'approved' | 'rejected' | null>(null);
  const [otp, setOtp] = useState('');
  const [otpError, setOtpError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  // Base API path from environment variable
  const BASE_PATH = import.meta.env.VITE_BASE_PATH || 'http://localhost:3000'; 
  // For demonstration, using a placeholder. Replace with actual API base path.

  // Get consent_uuid from URL params
  const getConsentUuid = (): string | null => {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get('consent') ? urlParams.get('consent') : 'consent_abc123'; // Placeholder for testing
  };

  // Fetch consent data
  const fetchConsentData = async (consentUuid: string) => {
    // try {
    //   const response = await fetch(`${BASE_PATH}/consents/${consentUuid}`);
    //   if (!response.ok) {
    //     throw new Error('Failed to fetch consent data');
    //   }
    //   const data: ConsentRecord = await response.json();
    //   setConsentRecord(data);
    //   setCurrentStep('consent');
    // } catch (err) {
    //   setError('Failed to load consent information. Please try again.');
    //   setCurrentStep('error');
    // }
    // Mocked data for demonstration
    const mockedData: ConsentRecord = {
      consent_uuid: consentUuid,
      owner_id: 'user_12345',
      data_consumer: 'Example App',
      status: 'pending',
      type: 'Standard',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(), // 7 days later
      fields: ['profile.name', 'profile.email', 'address.street', 'address.city'],
      session_id: 'session_67890',
      redirect_url: 'http://localhost:4000/redirect' // Placeholder redirect URL
    };
    setConsentRecord(mockedData);
    setCurrentStep('consent');
  };

  // Initialize component
  useEffect(() => {
    const consentUuid = getConsentUuid();
    if (!consentUuid) {
      setError('Invalid consent link. Missing consent ID.');
      setCurrentStep('error');
      return;
    }
    
    fetchConsentData(consentUuid);
  }, []);

  // Handle consent decision
  const handleConsentDecision = (decision: 'approved' | 'rejected') => {
    setUserDecision(decision);
    setCurrentStep('otp');
    setOtp('');
    setOtpError('');
  };

  // Validate OTP
  const validateOtp = (inputOtp: string): boolean => {
    return inputOtp === '123456'; // Hardcoded as requested
  };

  // Submit consent decision
  const submitConsentDecision = async () => {
    if (!validateOtp(otp)) {
      setOtpError('Invalid OTP. Please enter the correct code.');
      return;
    }

    if (!consentRecord || !userDecision) {
      setError('Missing consent data or decision');
      return;
    }

    setIsSubmitting(true);
    setOtpError('');

    try {
      const payload = {
        ...consentRecord,
        status: userDecision
      };

      const response = await fetch(`${BASE_PATH}/consents/${consentRecord.consent_uuid}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload)
      });

      if (!response.ok) {
        throw new Error('Failed to update consent');
      }

      setCurrentStep('success');
      
      // Redirect after 3 seconds
      setTimeout(() => {
        if (consentRecord.redirect_url) {
          window.location.href = consentRecord.redirect_url;
        }
      }, 3000);

    } catch (err) {
      setError('Failed to process your consent decision. Please try again.');
      setCurrentStep('error');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Format field names for display
  const formatFieldName = (field: string): string => {
    return field.split('.').map(part => 
      part.charAt(0).toUpperCase() + part.slice(1)
    ).join(' - ');
  };

  // Format date for display
  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleString();
  };

  // Loading state
  if (currentStep === 'loading') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading consent information...</p>
        </div>
      </div>
    );
  }

  // Error state
  if (currentStep === 'error') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-red-50 to-pink-100 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6 text-center">
          <AlertCircle className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h1 className="text-xl font-bold text-gray-800 mb-2">Error</h1>
          <p className="text-gray-600 mb-4">{error}</p>
          <button 
            onClick={() => window.location.reload()} 
            className="bg-red-600 text-white px-6 py-2 rounded-lg hover:bg-red-700 transition-colors"
          >
            Try Again
          </button>
        </div>
      </div>
    );
  }

  // Success state
  if (currentStep === 'success') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-green-50 to-emerald-100 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6 text-center">
          <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-4" />
          <h1 className="text-xl font-bold text-gray-800 mb-2">Success!</h1>
          <p className="text-gray-600 mb-4">
            Your consent has been {userDecision === 'approved' ? 'approved' : 'denied'} successfully.
          </p>
          <p className="text-sm text-gray-500">
            Redirecting you back to the application...
          </p>
        </div>
      </div>
    );
  }

  // OTP verification state
  if (currentStep === 'otp') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6">
          <div className="text-center mb-6">
            <Lock className="h-12 w-12 text-indigo-600 mx-auto mb-4" />
            <h1 className="text-xl font-bold text-gray-800 mb-2">Verify Your Decision</h1>
            <p className="text-gray-600">
              Please enter the OTP to confirm your decision to {' '}
              <span className={`font-semibold ${userDecision === 'approved' ? 'text-green-600' : 'text-red-600'}`}>
                {userDecision === 'approved' ? 'approve' : 'deny'}
              </span>
              {' '}the consent request.
            </p>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Enter OTP
              </label>
              <input
                type="text"
                value={otp}
                onChange={(e) => {
                  setOtp(e.target.value);
                  setOtpError('');
                }}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder="Enter 6-digit OTP"
                maxLength={6}
              />
              {otpError && (
                <p className="mt-1 text-sm text-red-600">{otpError}</p>
              )}
            </div>

            <div className="flex space-x-3">
              <button
                onClick={() => setCurrentStep('consent')}
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
                disabled={isSubmitting}
              >
                Back
              </button>
              <button
                onClick={submitConsentDecision}
                disabled={!otp || isSubmitting}
                className={`flex-1 px-4 py-2 rounded-lg text-white transition-colors ${
                  userDecision === 'approved' 
                    ? 'bg-green-600 hover:bg-green-700 disabled:bg-green-400' 
                    : 'bg-red-600 hover:bg-red-700 disabled:bg-red-400'
                }`}
              >
                {isSubmitting ? 'Processing...' : 'Confirm'}
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Main consent approval state
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 p-4">
      <div className="max-w-2xl mx-auto py-8">
        <div className="bg-white rounded-lg shadow-lg overflow-hidden">
          {/* Header */}
          <div className="bg-indigo-600 text-white p-6">
            <div className="flex items-center">
              <Shield className="h-8 w-8 mr-3" />
              <div>
                <h1 className="text-2xl font-bold">Consent Request</h1>
                <p className="text-indigo-100">Please review and approve data access</p>
              </div>
            </div>
          </div>

          {/* Content */}
          <div className="p-6">
            {consentRecord && (
              <>
                {/* Application Info */}
                <div className="mb-6 p-4 bg-gray-50 rounded-lg">
                  <h2 className="text-lg font-semibold text-gray-800 mb-3">Application Request</h2>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
                    <div>
                      <span className="font-medium text-gray-600">Application:</span>
                      <span className="ml-2 text-gray-800">{consentRecord.data_consumer}</span>
                    </div>
                    <div>
                      <span className="font-medium text-gray-600">Owner ID:</span>
                      <span className="ml-2 text-gray-800">{consentRecord.owner_id}</span>
                    </div>
                    <div>
                      <span className="font-medium text-gray-600">Request Type:</span>
                      <span className="ml-2 text-gray-800">{consentRecord.type || 'Standard'}</span>
                    </div>
                    <div>
                      <span className="font-medium text-gray-600">Expires:</span>
                      <span className="ml-2 text-gray-800">{formatDate(consentRecord.expires_at)}</span>
                    </div>
                  </div>
                </div>

                {/* Fields Requested */}
                <div className="mb-6">
                  <h2 className="text-lg font-semibold text-gray-800 mb-3">Data Fields Requested</h2>
                  <div className="space-y-2">
                    {consentRecord.fields.map((field, index) => (
                      <div key={index} className="flex items-center p-3 bg-blue-50 rounded-lg">
                        <div className="w-2 h-2 bg-blue-500 rounded-full mr-3"></div>
                        <span className="text-gray-800 font-medium">{formatFieldName(field)}</span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Warning */}
                <div className="mb-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
                  <div className="flex items-start">
                    <AlertCircle className="h-5 w-5 text-yellow-600 mr-2 mt-0.5" />
                    <div className="text-sm text-yellow-800">
                      <p className="font-medium mb-1">Important:</p>
                      <p>By approving this request, you are granting {consentRecord.data_consumer} access to the specified data fields. This access will remain valid until {formatDate(consentRecord.expires_at)}.</p>
                    </div>
                  </div>
                </div>

                {/* Action Buttons */}
                <div className="flex space-x-4">
                  <button
                    onClick={() => handleConsentDecision('approved')}
                    className="flex-1 bg-green-600 hover:bg-green-700 text-white py-3 px-6 rounded-lg font-medium transition-colors flex items-center justify-center"
                  >
                    <Check className="h-5 w-5 mr-2" />
                    Approve Consent
                  </button>
                  <button
                    onClick={() => handleConsentDecision('rejected')}
                    className="flex-1 bg-red-600 hover:bg-red-700 text-white py-3 px-6 rounded-lg font-medium transition-colors flex items-center justify-center"
                  >
                    <X className="h-5 w-5 mr-2" />
                    Deny Consent
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ConsentGateway;