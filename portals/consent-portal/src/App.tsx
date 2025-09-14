import React, { useState, useEffect } from 'react';
import { Shield, Check, X, Lock, AlertCircle, CheckCircle } from 'lucide-react';

// Types
interface ConsentRecord {
  consent_uuid: string;
  owner_id: string;
  data_consumer: string;
  status: 'pending' | 'approved' | 'rejected' | 'expired' | 'revoked';
  type?: string;
  created_at: string;
  updated_at: string;
  expires_at: string;
  fields: string[];
  session_id: string;
  redirect_url: string;
}

interface OwnerInfo {
  owner_id: string;
  email: string;
  contact_number: string;
  name?: string;
}

interface ConsentGatewayProps {}

const ConsentGateway: React.FC<ConsentGatewayProps> = () => {
  const [currentStep, setCurrentStep] = useState<'loading' | 'consent' | 'otp' | 'success' | 'error' | 'statusInfo'>('loading');
  const [consentRecord, setConsentRecord] = useState<ConsentRecord | null>(null);
  const [ownerInfo, setOwnerInfo] = useState<OwnerInfo | null>(null);
  const [userDecision, setUserDecision] = useState<'approved' | 'rejected' | null>(null);
  const [otp, setOtp] = useState('');
  const [otpError, setOtpError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [otpExpiryTime, setOtpExpiryTime] = useState<Date | null>(null);
  const [deliveryMethod, setDeliveryMethod] = useState<'email' | 'sms'>('email');
  const [isResendingOtp, setIsResendingOtp] = useState(false);
  const [otpSent, setOtpSent] = useState(false);
  const [currentTime, setCurrentTime] = useState(new Date());

  // Base API path from environment variable
  const BASE_PATH = import.meta.env.VITE_BASE_PATH || 'http://localhost:3000';
  const CONSENT_ENGINE_PATH = import.meta.env.VITE_CONSENT_ENGINE_PATH || 'http://localhost:8081';
  // For demonstration, using a placeholder. Replace with actual API base path.

  // Get consent_id from URL params
  const getConsentId = (): string | null => {
    const urlParams = new URLSearchParams(window.location.search);
    console.log('URL Params:', urlParams.toString());
    return urlParams.get('consent_id') || urlParams.get('consent') || 'consent_5df473bd'; // Fallback for testing
  };

  // Fetch consent data
  const fetchConsentData = async (consentUuid: string) => {
    try {
      const response = await fetch(`${CONSENT_ENGINE_PATH}/consents/${consentUuid}`);
      if (!response.ok) {
        throw new Error('Failed to fetch consent data');
      }
      const data: ConsentRecord = await response.json();
      setConsentRecord(data);
      
      // Check consent status and set appropriate step
      if (data.status === 'pending') {
        setCurrentStep('consent');
      } else {
        setCurrentStep('statusInfo');
      }
      await fetchOwnerInfo(consentRecord ? consentRecord.owner_id : 'user_12345');

    } catch (err) {
      setError('Failed to load consent information. Please try again.');
      setCurrentStep('error');
    }
    // Mocked data for demonstration
    // const mockedData: ConsentRecord = {
    //   consent_uuid: consentUuid,
    //   owner_id: 'user_12345',
    //   data_consumer: 'Example App',
    //   status: 'pending', // Change this to test different statuses: 'approved', 'rejected', 'expired', 'revoked'
    //   type: 'Standard',
    //   created_at: new Date().toISOString(),
    //   updated_at: new Date().toISOString(),
    //   expires_at: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(), // 7 days later
    //   fields: ['profile.name', 'profile.email', 'address.street', 'address.city'],
    //   session_id: 'session_67890',
    //   redirect_url: 'http://localhost:4000/redirect' // Placeholder redirect URL
    // };
    // setConsentRecord(mockedData);
    
    // Fetch owner information after getting consent data
    // await fetchOwnerInfo(consentRecord ? consentRecord.owner_id : 'user_12345');
    
    // setCurrentStep('consent');
  };

  // Fetch owner information
  const fetchOwnerInfo = async (ownerId: string) => {
    try {
      // const response = await fetch(`${BASE_PATH}/owners/${ownerId}`);
      // if (!response.ok) {
      //   throw new Error('Failed to fetch owner information');
      // }
      // const data: OwnerInfo = await response.json();
      // setOwnerInfo(data);
      
      // Mocked data for demonstration
      const mockedOwnerData: OwnerInfo = {
        owner_id: ownerId,
        email: 'user@example.com',
        contact_number: '+1234567890',
        name: 'John Doe'
      };
      setOwnerInfo(mockedOwnerData);
    } catch (err) {
      console.error('Failed to fetch owner information:', err);
      // Continue without owner info - this is optional
    }
  };

  // Send OTP
  const sendOTP = async (method: 'email' | 'sms') => {
    if (!ownerInfo || !consentRecord) {
      setError('Missing owner or consent information');
      return false;
    }

    try {
      const payload = {
        owner_id: ownerInfo.owner_id,
        consent_uuid: consentRecord.consent_uuid,
        delivery_method: method,
        email: method === 'email' ? ownerInfo.email : undefined,
        contact_number: method === 'sms' ? ownerInfo.contact_number : undefined,
        decision: userDecision
      };

      // const response = await fetch(`${BASE_PATH}/otp/send`, {
      //   method: 'POST',
      //   headers: {
      //     'Content-Type': 'application/json',
      //   },
      //   body: JSON.stringify(payload)
      // });

      // if (!response.ok) {
      //   throw new Error('Failed to send OTP');
      // }

      // For demonstration, simulate API call
      console.log('Sending OTP via', method, 'to', method === 'email' ? ownerInfo.email : ownerInfo.contact_number);
      
      // Set OTP expiry time (5 minutes from now)
      const expiryTime = new Date();
      expiryTime.setMinutes(expiryTime.getMinutes() + 5);
      setOtpExpiryTime(expiryTime);
      setOtpSent(true);
      
      return true;
    } catch (err) {
      console.error('Failed to send OTP:', err);
      setError('Failed to send OTP. Please try again.');
      return false;
    }
  };

  // Initialize component
  useEffect(() => {
    const consentId = getConsentId();
    if (!consentId) {
      setError('Invalid consent link. Missing consent ID.');
      setCurrentStep('error');
      return;
    }

    fetchConsentData(consentId);
  }, []);

  // Timer for OTP expiry countdown
  useEffect(() => {
    let interval: NodeJS.Timeout;
    
    if (currentStep === 'otp' && otpExpiryTime) {
      interval = setInterval(() => {
        setCurrentTime(new Date()); // Update current time to trigger re-render
        if (isOtpExpired()) {
          setOtpError('OTP has expired. Please request a new one.');
        }
      }, 1000);
    }
    
    return () => {
      if (interval) {
        clearInterval(interval);
      }
    };
  }, [currentStep, otpExpiryTime]);

  // Handle consent decision
  const handleConsentDecision = async (decision: 'approved' | 'rejected') => {
    setUserDecision(decision);
    setOtp('');
    setOtpError('');
    setError('');
    
    if (decision === 'approved') {
      // For approved decisions, send OTP and proceed to OTP verification
      const otpSent = await sendOTP(deliveryMethod);
      if (otpSent) {
        setCurrentStep('otp');
      } else {
        setCurrentStep('error');
      }
    } else {
      // For rejected decisions, submit immediately without OTP
      await submitConsentDecision();
    }
  };

  // Resend OTP
  const resendOTP = async () => {
    setIsResendingOtp(true);
    setOtpError('');
    
    const otpSent = await sendOTP(deliveryMethod);
    if (!otpSent) {
      setOtpError('Failed to resend OTP. Please try again.');
    }
    
    setIsResendingOtp(false);
  };

  // Check if OTP is expired
  const isOtpExpired = (): boolean => {
    if (!otpExpiryTime) return false;
    return new Date() > otpExpiryTime;
  };

  // Get remaining time for OTP expiry
  const getRemainingTime = (): string => {
    if (!otpExpiryTime) return '';
    
    const now = currentTime; // Use currentTime state instead of new Date()
    const remaining = otpExpiryTime.getTime() - now.getTime();
    
    if (remaining <= 0) return 'Expired';
    
    const minutes = Math.floor(remaining / 60000);
    const seconds = Math.floor((remaining % 60000) / 1000);
    
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  };

  // Validate OTP
  const validateOtp = (inputOtp: string): boolean => {
    // Check if OTP is expired first
    if (isOtpExpired()) {
      setOtpError('OTP has expired. Please request a new one.');
      return false;
    }
    
    // Check OTP value (hardcoded as requested)
    if (inputOtp !== '123456') {
      setOtpError('Invalid OTP. Please enter the correct code.');
      return false;
    }
    
    return true;
  };

  // Submit consent decision
  const submitConsentDecision = async () => {
    // Only validate OTP for approved decisions
    if (userDecision === 'approved' && !validateOtp(otp)) {
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
        status: userDecision,
        owner_id: consentRecord.owner_id,
        message: userDecision === 'approved' ? 'User approved consent via portal' : 'User rejected consent via portal'
      };

      const response = await fetch(`${CONSENT_ENGINE_PATH}/consents/${consentRecord.consent_uuid}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload)
      });

      if (!response.ok) {
        throw new Error('Failed to update consent');
      }

      setCurrentStep('success');
      
      // Notify parent window of consent completion
      if (window.opener) {
        window.opener.postMessage({
          type: 'consent-complete',
          status: userDecision,
          consentId: consentRecord.consent_uuid
        }, window.location.origin);
      }
      
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
              Please enter the OTP sent to your {deliveryMethod === 'email' ? 'email' : 'phone'} to confirm your decision to {' '}
              <span className={`font-semibold ${userDecision === 'approved' ? 'text-green-600' : 'text-red-600'}`}>
                {userDecision === 'approved' ? 'approve' : 'deny'}
              </span>
              {' '}the consent request.
            </p>
            {ownerInfo && (
              <p className="text-sm text-gray-500 mt-2">
                OTP sent to: {deliveryMethod === 'email' ? ownerInfo.email : ownerInfo.contact_number}
              </p>
            )}
          </div>

          {/* Delivery Method Selection */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Delivery Method
            </label>
            <div className="flex space-x-2">
              <button
                onClick={() => setDeliveryMethod('email')}
                className={`flex-1 px-3 py-2 text-sm rounded-lg border transition-colors ${
                  deliveryMethod === 'email'
                    ? 'bg-indigo-50 border-indigo-300 text-indigo-700'
                    : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
                }`}
                disabled={isSubmitting || isResendingOtp}
              >
                Email
              </button>
              <button
                onClick={() => setDeliveryMethod('sms')}
                className={`flex-1 px-3 py-2 text-sm rounded-lg border transition-colors ${
                  deliveryMethod === 'sms'
                    ? 'bg-indigo-50 border-indigo-300 text-indigo-700'
                    : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
                }`}
                disabled={isSubmitting || isResendingOtp}
              >
                SMS
              </button>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <div className="flex justify-between items-center mb-1">
                <label className="block text-sm font-medium text-gray-700">
                  Enter OTP
                </label>
                {otpExpiryTime && (
                  <span className={`text-xs ${isOtpExpired() ? 'text-red-600' : 'text-gray-500'}`}>
                    {isOtpExpired() ? 'Expired' : `Expires in ${getRemainingTime()}`}
                  </span>
                )}
              </div>
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
                disabled={isOtpExpired()}
              />
              {otpError && (
                <p className="mt-1 text-sm text-red-600">{otpError}</p>
              )}
            </div>

            {/* Resend OTP */}
            <div className="text-center">
              <button
                onClick={resendOTP}
                disabled={isResendingOtp || isSubmitting}
                className="text-sm text-indigo-600 hover:text-indigo-700 disabled:text-gray-400 transition-colors"
              >
                {isResendingOtp ? 'Resending...' : 'Resend OTP'}
              </button>
            </div>

            <div className="flex space-x-3">
              <button
                onClick={() => setCurrentStep('consent')}
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
                disabled={isSubmitting || isResendingOtp}
              >
                Back
              </button>
              <button
                onClick={submitConsentDecision}
                disabled={!otp || isSubmitting || isResendingOtp || isOtpExpired()}
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

  // Status information state (for non-pending consents)
  if (currentStep === 'statusInfo' && consentRecord) {
    const getStatusIcon = () => {
      switch (consentRecord.status) {
        case 'approved':
          return <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-4" />;
        case 'rejected':
          return <X className="h-12 w-12 text-red-500 mx-auto mb-4" />;
        case 'expired':
          return <AlertCircle className="h-12 w-12 text-orange-500 mx-auto mb-4" />;
        case 'revoked':
          return <X className="h-12 w-12 text-gray-500 mx-auto mb-4" />;
        default:
          return <AlertCircle className="h-12 w-12 text-gray-500 mx-auto mb-4" />;
      }
    };

    const getStatusMessage = () => {
      switch (consentRecord.status) {
        case 'approved':
          return {
            title: 'Consent Already Approved',
            message: `This consent request has already been approved on ${formatDate(consentRecord.updated_at)}.`,
            bgColor: 'from-green-50 to-emerald-100'
          };
        case 'rejected':
          return {
            title: 'Consent Already Rejected',
            message: `This consent request was rejected on ${formatDate(consentRecord.updated_at)}.`,
            bgColor: 'from-red-50 to-pink-100'
          };
        case 'expired':
          return {
            title: 'Consent Expired',
            message: `This consent request expired on ${formatDate(consentRecord.expires_at)} and is no longer valid.`,
            bgColor: 'from-orange-50 to-yellow-100'
          };
        case 'revoked':
          return {
            title: 'Consent Revoked',
            message: `This consent was revoked on ${formatDate(consentRecord.updated_at)} and is no longer active.`,
            bgColor: 'from-gray-50 to-slate-100'
          };
        default:
          return {
            title: 'Consent Status',
            message: `This consent is currently ${consentRecord.status}.`,
            bgColor: 'from-gray-50 to-slate-100'
          };
      }
    };

    const statusInfo = getStatusMessage();

    return (
      <div className={`min-h-screen bg-gradient-to-br ${statusInfo.bgColor} flex items-center justify-center p-4`}>
        <div className="max-w-2xl w-full bg-white rounded-lg shadow-lg overflow-hidden">
          {/* Header */}
          <div className="bg-indigo-600 text-white p-6">
            <div className="flex items-center">
              <Shield className="h-8 w-8 mr-3" />
              <div>
                <h1 className="text-2xl font-bold">Consent Status</h1>
                <p className="text-indigo-100">Information about your consent request</p>
              </div>
            </div>
          </div>

          {/* Status Content */}
          <div className="p-6 text-center">
            {getStatusIcon()}
            <h2 className="text-xl font-bold text-gray-800 mb-2">{statusInfo.title}</h2>
            <p className="text-gray-600 mb-6">{statusInfo.message}</p>

            {/* Consent Details */}
            <div className="mb-6 p-4 bg-gray-50 rounded-lg text-left">
              <h3 className="text-lg font-semibold text-gray-800 mb-3">Consent Details</h3>
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
                  <span className="font-medium text-gray-600">Status:</span>
                  <span className={`ml-2 font-medium capitalize ${
                    consentRecord.status === 'approved' ? 'text-green-600' :
                    consentRecord.status === 'rejected' ? 'text-red-600' :
                    consentRecord.status === 'expired' ? 'text-orange-600' :
                    'text-gray-600'
                  }`}>
                    {consentRecord.status}
                  </span>
                </div>
                <div>
                  <span className="font-medium text-gray-600">Last Updated:</span>
                  <span className="ml-2 text-gray-800">{formatDate(consentRecord.updated_at)}</span>
                </div>
              </div>
            </div>

            {/* Data Fields */}
            <div className="mb-6 text-left">
              <h3 className="text-lg font-semibold text-gray-800 mb-3">Data Fields</h3>
              <div className="space-y-2">
                {consentRecord.fields.map((field, index) => (
                  <div key={index} className="flex items-center p-3 bg-blue-50 rounded-lg">
                    <div className="w-2 h-2 bg-blue-500 rounded-full mr-3"></div>
                    <span className="text-gray-800 font-medium">{formatFieldName(field)}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Redirect Button */}
            <button 
              onClick={() => {
                if (consentRecord.redirect_url) {
                  window.location.href = consentRecord.redirect_url;
                } else {
                  window.close();
                }
              }}
              className="bg-indigo-600 text-white px-6 py-2 rounded-lg hover:bg-indigo-700 transition-colors"
            >
              Return to Application
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Main consent approval state
  if (currentStep === 'consent') {
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

                {/* OTP Delivery Method Selection */}
                {ownerInfo && (
                  <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
                    <h3 className="text-sm font-medium text-gray-800 mb-3">OTP Delivery Method</h3>
                    <div className="space-y-2">
                      <label className="flex items-center">
                        <input
                          type="radio"
                          name="deliveryMethod"
                          value="email"
                          checked={deliveryMethod === 'email'}
                          onChange={(e) => setDeliveryMethod(e.target.value as 'email' | 'sms')}
                          className="mr-2"
                        />
                        <span className="text-sm text-gray-700">Email: {ownerInfo.email}</span>
                      </label>
                      <label className="flex items-center">
                        <input
                          type="radio"
                          name="deliveryMethod"
                          value="sms"
                          checked={deliveryMethod === 'sms'}
                          onChange={(e) => setDeliveryMethod(e.target.value as 'email' | 'sms')}
                          className="mr-2"
                        />
                        <span className="text-sm text-gray-700">SMS: {ownerInfo.contact_number}</span>
                      </label>
                    </div>
                  </div>
                )}

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
  }

  // Default fallback (should not reach here)  
  return null;
};

export default ConsentGateway;