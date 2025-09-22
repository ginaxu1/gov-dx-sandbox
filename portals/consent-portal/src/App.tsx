import React, { useState, useEffect } from 'react';
import { useAuthContext } from "@asgardeo/auth-react";
import { Shield, CheckCircle, X, AlertCircle } from 'lucide-react';

// Extend Window interface to include config
declare global {
  interface Window {
    configs: {
      apiUrl: string;
      VITE_CLIENT_ID: string;
      VITE_BASE_URL: string;
      VITE_SCOPE: string;
      signInRedirectURL: string;
      signOutRedirectURL: string;
    };
  }
}

// Types
interface ConsentRecord {
  consent_id: string;
  owner_id: string;  
  owner_name: string;
  owner_email: string;
  data_consumer: string;
  status: 'pending' | 'approved' | 'rejected' | 'expired' | 'revoked';
  type?: string;
  created_at: string;
  updated_at: string;
  expires_at: string;
  fields: string[];
  session_id: string;
  redirect_url: string;
  app_display_name: string;
}

interface ConsentGatewayProps {}

const ConsentGateway: React.FC<ConsentGatewayProps> = () => {
  // State management
  const [currentStep, setCurrentStep] = useState<'loading' | 'consent' | 'success' | 'error' | 'statusInfo' | 'unauthorized'>('loading');
  const [consentRecord, setConsentRecord] = useState<ConsentRecord | null>(null);
  const [error, setError] = useState('');
  const [consentId, setConsentId] = useState<string | null>(null);
  const [userEmail, setUserEmail] = useState<string>('');
  const [userName, setName] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { state, signIn, signOut, getBasicUserInfo, getAccessToken } = useAuthContext();

  // Base API path from environment variable
  const CONSENT_ENGINE_PATH = window?.configs?.apiUrl;

  // Helper function to get headers with JWT token
  const getAuthHeaders = async (): Promise<HeadersInit> => {
    const accessToken = await getAccessToken();
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      // 'X-Requested-With': 'XMLHttpRequest', // Mark as frontend request for hybrid auth
      // 'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36' // Browser user agent for frontend detection
    };
    
    if (accessToken) {
      console.log("Using access token for API call");
      console.log("Access Token:", accessToken);
      headers['Authorization'] = `Bearer ${accessToken}`;
      console.log("Headers set for frontend request:", headers);
    } else {
      console.warn("No access token available - request may fail for frontend authentication");
      console.log("Headers set without Authorization:", headers);
    }
    
    return headers;
  };

  // Step 1: Check for consent_id in URL parameters
  const getConsentIdFromUrl = (): string | null => {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get('consent_id');
  };

  // Step 2: Get authenticated user info
  const fetchUserInfo = async () => {
    try {
      const userBasicInfo = await getBasicUserInfo();
      console.log('User Basic Info:', userBasicInfo);
      
      if (userBasicInfo) {
        setName(userBasicInfo.name || '');
        setUserEmail(userBasicInfo.email || '');
      }
    } catch (error) {
      console.error('Failed to fetch user info:', error);
    }
  };

  // Step 3: Fetch consent data
  const fetchConsentData = async (consentUuid: string) => {
    try {
      const headers = await getAuthHeaders();
      const response = await fetch(`${CONSENT_ENGINE_PATH}/consents/${consentUuid}`, {
        headers
      });

      // Handle different error responses
      if (!response.ok) {
        let errorMessage = '';
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || '';
        } catch (parseError) {
          console.warn('Failed to parse error response:', parseError);
        }
        
        if (response.status === 401) {
          throw new Error(errorMessage || 'Unauthorized: Please sign in to access this consent');
        } else if (response.status === 403) {
          // throw new Error(errorMessage || 'Forbidden: You do not have permission to access this consent');
          setCurrentStep('unauthorized');
          setError(errorMessage || 'Forbidden: You do not have permission to access this consent');
          return null;
        } else if (response.status === 404) {
          throw new Error(errorMessage || 'Consent not found');
        } else {
          throw new Error(errorMessage || `Failed to fetch consent data: ${response.status} ${response.statusText}`);
        }
      }
      const data: ConsentRecord = await response.json();
      setConsentRecord({
        ...data,
        consent_id: consentUuid,
      });
      
      return data;
    } catch (err) {
      console.error('Failed to fetch consent data:', err);
      setError(err instanceof Error ? err.message : 'Failed to load consent information. Please try again.');
      setCurrentStep('error');
      return null;
    }
  };

  // Step 4: Handle consent decision (approve/deny)
  const handleConsentDecision = async (decision: 'approved' | 'rejected') => {
    if (!consentRecord) {
      setError('Missing consent data');
      return;
    }

    setIsSubmitting(true);

    try {
      const payload = {
        ...consentRecord,
        status: decision,
        updated_by: userEmail || 'self-service',
        reason: decision === 'rejected' ? 'User denied the consent request via self-service portal' : 'User approved the consent request via self-service portal'
      };

      const headers = await getAuthHeaders();
      const response = await fetch(`${CONSENT_ENGINE_PATH}/consents/${consentRecord.consent_id}`, {
        method: 'PUT',
        headers,
        body: JSON.stringify(payload)
      });

      if (!response.ok) {
        let errorMessage = '';
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || '';
        } catch (parseError) {
          console.warn('Failed to parse error response:', parseError);
        }
        
        if (response.status === 401) {
          throw new Error(errorMessage || 'Unauthorized: Please sign in to update this consent');
        } else if (response.status === 403) {
          throw new Error(errorMessage || 'Forbidden: You do not have permission to update this consent');
        } else if (response.status === 404) {
          throw new Error(errorMessage || 'Consent not found');
        } else {
          throw new Error(errorMessage || `Failed to update consent: ${response.status} ${response.statusText}`);
        }
      }

      setCurrentStep('success');
      
      // Clear consent_id from storage since the flow is complete
      clearConsentIdFromStorage();
      
      // Redirect after success
      setTimeout(() => {
        if (window.opener) {
          window.opener.postMessage("consent_granted", "*");
        }
        if (consentRecord.redirect_url) {
          window.location.href = consentRecord.redirect_url;
        } else {
          window.close();
        }
      }, 3000);

    } catch (err) {
      console.error('Failed to process consent decision:', err);
      setError(err instanceof Error ? err.message : 'Failed to process your consent decision. Please try again.');
      setCurrentStep('error');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Format field names for display
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

  // Format date for display
  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleString();
  };

  // Save consent_id to localStorage to persist through auth redirects
  const saveConsentIdToStorage = (id: string) => {
    localStorage.setItem('consent_id', id);
  };

  // Get consent_id from localStorage
  const getConsentIdFromStorage = (): string | null => {
    return localStorage.getItem('consent_id');
  };

  // Clear consent_id from localStorage
  const clearConsentIdFromStorage = () => {
    localStorage.removeItem('consent_id');
  };

  // Main initialization effect
  useEffect(() => {
    // Step 1: Check for consent_id in URL first, then fallback to localStorage
    let urlConsentId = getConsentIdFromUrl();
    
    if (!urlConsentId) {
      // Try to get from localStorage (after auth redirect)
      urlConsentId = getConsentIdFromStorage();
    }
    
    if (!urlConsentId) {
      setError('URL missing the consent_id parameter');
      setCurrentStep('error');
      return;
    }
    
    // Save to localStorage and update URL if needed
    saveConsentIdToStorage(urlConsentId);
    
    // Update URL to include consent_id if it's missing
    const currentUrl = new URL(window.location.href);
    if (!currentUrl.searchParams.get('consent_id')) {
      currentUrl.searchParams.set('consent_id', urlConsentId);
      window.history.replaceState({}, '', currentUrl.toString());
    }
    
    setConsentId(urlConsentId);
  }, []);

  // Handle authentication state changes and restore flow after auth redirect
  useEffect(() => {
    // When user becomes authenticated, check if we have a stored consent_id
    if (state.isAuthenticated && !consentId) {
      const storedConsentId = getConsentIdFromStorage();
      if (storedConsentId) {
        setConsentId(storedConsentId);
        // Update URL to include consent_id
        const currentUrl = new URL(window.location.href);
        if (!currentUrl.searchParams.get('consent_id')) {
          currentUrl.searchParams.set('consent_id', storedConsentId);
          window.history.replaceState({}, '', currentUrl.toString());
        }
      }
    }
  }, [state.isAuthenticated, consentId]);

  // Handle authentication and consent data fetching
  useEffect(() => {
    if (!consentId) return;

    // Step 2: Check authentication
    if (!state.isAuthenticated) {
      console.log('User not authenticated, waiting for login');
      return;
    }

    // Step 3: Get user info and fetch consent data
    const initializeConsentFlow = async () => {
      // await fetchUserInfo();
      const consent = await fetchConsentData(consentId);
      
      if (consent) {
        // Step 4: Check owner email authorization
        // console.log('Consent Owner Email:', consent.owner_email);
        // console.log('Authenticated User Email:', userEmail);
        // if (consent.owner_email !== userEmail) {
        //   setError('Unauthorized access: You are not the data owner for this consent request');
        //   setCurrentStep('unauthorized');
        //   return;
        // }
        
        // Step 5: Set appropriate step based on status
        if (consent.status === 'pending') {
          setCurrentStep('consent');
        } else {
          setCurrentStep('statusInfo');
        }
      }
    };

    initializeConsentFlow();
  }, [state.isAuthenticated, consentId, userEmail]);

  // Update user email when authentication state changes
  useEffect(() => {
    if (state.isAuthenticated) {
      fetchUserInfo();
    }
  }, [state.isAuthenticated]);

  // Handle sign in with consent_id preservation
  const handleSignIn = () => {
    // Ensure consent_id is saved before redirect
    const currentConsentId = consentId || getConsentIdFromUrl() || getConsentIdFromStorage();
    if (currentConsentId) {
      saveConsentIdToStorage(currentConsentId);
    }
    signIn();
  };

  // Handle sign out with consent_id cleanup
  const handleSignOut = () => {
    clearConsentIdFromStorage();
    signOut();
  };

  // User header component
  const UserHeader = () => {
    if (!state.isAuthenticated) {
      return (
        <div className="absolute top-4 right-4 flex items-center space-x-4 bg-white rounded-lg shadow-md px-4 py-2">
          <div className="text-sm text-gray-600">
            Please sign in to continue.
          </div>
          <button
            onClick={handleSignIn}
            className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg"
          >
            Sign In
          </button>
        </div>
      );
    }
    
    return (
      <div className="absolute top-4 right-4 flex items-center space-x-4 bg-white rounded-lg shadow-md px-4 py-2">
        <div className="text-sm text-gray-600">
          Welcome, <span className="font-medium text-gray-800">{userName}</span>
        </div>
        <button
          onClick={handleSignOut}
          className="text-red-600 hover:text-red-800 text-sm font-medium transition-colors"
        >
          Sign Out
        </button>
      </div>
    );
  };

  // Error state - including URL missing consent_id
  if (currentStep === 'error') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-red-50 to-pink-100 flex items-center justify-center p-4 relative">
        <UserHeader />
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

  // Unauthorized access state
  if (currentStep === 'unauthorized') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-orange-50 to-red-100 flex items-center justify-center p-4 relative">
        <UserHeader />
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6 text-center">
          <X className="h-12 w-12 text-orange-500 mx-auto mb-4" />
          <h1 className="text-xl font-bold text-gray-800 mb-2">Unauthorized Access</h1>
          <p className="text-gray-600 mb-4">{error}</p>
          <p className="text-sm text-gray-500 mb-4">
            Your email: <span className="font-mono text-blue-600">{userEmail}</span>
          </p>
          <button 
            onClick={handleSignOut}
            className="bg-orange-600 text-white px-6 py-2 rounded-lg hover:bg-orange-700 transition-colors"
          >
            Sign Out
          </button>
        </div>
      </div>
    );
  }

  // Login required state
  if (!state.isAuthenticated && consentId) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4 relative">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6 text-center">
          <Shield className="h-12 w-12 text-blue-500 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-gray-800 mb-4">Consent Portal</h1>
          <p className="text-gray-600 mb-4">
            You need to sign in to process your consent request.
          </p>
          <p className="text-sm text-gray-500 mb-6">
            Consent ID: <span className="font-mono text-blue-600">{consentId}</span>
          </p>
          <button 
            onClick={handleSignIn} 
            className="bg-blue-500 hover:bg-blue-600 text-white px-6 py-3 rounded-lg font-medium transition-colors"
          >
            Sign In to Continue
          </button>
        </div>
      </div>
    );
  }

  // Loading state
  if (currentStep === 'loading') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center relative">
        <UserHeader />
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading consent information...</p>
        </div>
      </div>
    );
  }

  // Success state
  if (currentStep === 'success') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-green-50 to-emerald-100 flex items-center justify-center p-4 relative">
        <UserHeader />
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6 text-center">
          <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-4" />
          <h1 className="text-xl font-bold text-gray-800 mb-2">Success!</h1>
          <p className="text-gray-600 mb-4">
            Your consent has been processed successfully.
          </p>
          <p className="text-sm text-gray-500">
            Redirecting you back to the application...
          </p>
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
      <div className={`min-h-screen bg-gradient-to-br ${statusInfo.bgColor} flex items-center justify-center p-4 relative`}>
        <UserHeader />
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
                  <span className="ml-2 text-gray-800">{consentRecord.app_display_name}</span>
                </div>
                <div>
                  <span className="font-medium text-gray-600">Owner Email:</span>
                  <span className="ml-2 text-gray-800">{consentRecord.owner_email}</span>
                </div>
                <div>
                  <span className="font-medium text-gray-600">Status:</span>
                  <span className={`ml-2 px-2 py-1 rounded-full text-xs font-medium ${
                    consentRecord.status === 'approved' ? 'bg-green-100 text-green-800' :
                    consentRecord.status === 'rejected' ? 'bg-red-100 text-red-800' :
                    consentRecord.status === 'expired' ? 'bg-orange-100 text-orange-800' :
                    'bg-gray-100 text-gray-800'
                  }`}>
                    {consentRecord.status.charAt(0).toUpperCase() + consentRecord.status.slice(1)}
                  </span>
                </div>
                <div>
                  <span className="font-medium text-gray-600">Created:</span>
                  <span className="ml-2 text-gray-800">{formatDate(consentRecord.created_at)}</span>
                </div>
              </div>
            </div>

            {/* Data Fields */}
            <div className="mb-6 text-left">
              <h3 className="text-lg font-semibold text-gray-800 mb-3">Data Fields</h3>
              <div className="space-y-2">
                {consentRecord.fields.map((field, index) => (
                  <div key={index} className="flex items-center p-2 bg-white border border-gray-200 rounded">
                    <div className="h-2 w-2 bg-indigo-400 rounded-full mr-3"></div>
                    <span className="text-gray-700">{formatFieldName(field)}</span>
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

  // Main consent approval state (status === "pending")
  if (currentStep === 'consent' && consentRecord) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 p-4 relative">
        <UserHeader />
        <div className="max-w-2xl mx-auto py-8">
          <div className="bg-white rounded-lg shadow-lg overflow-hidden">
            {/* Header */}
            <div className="bg-indigo-600 text-white p-6">
              <div className="flex items-center">
                <Shield className="h-8 w-8 mr-3" />
                <div>
                  <h1 className="text-2xl font-bold">Consent Request</h1>
                  <p className="text-indigo-100">Review and approve data sharing</p>
                </div>
              </div>
            </div>

            {/* Content */}
            <div className="p-6">
              {/* Application Info */}
              <div className="mb-6 p-4 bg-blue-50 rounded-lg">
                <h3 className="text-lg font-semibold text-gray-800 mb-2">Application Request</h3>
                <p className="text-gray-600">
                  <span className="font-medium">{consentRecord.app_display_name}</span> is requesting access to the following data fields:
                </p>
              </div>

              {/* Data Fields */}
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

              {/* Consent Details */}
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

              {/* Action Buttons */}
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
  }

  // Default fallback (should not reach here)  
  return null;
};

export default ConsentGateway;