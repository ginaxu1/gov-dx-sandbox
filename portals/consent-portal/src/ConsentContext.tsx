import React, {createContext, useContext, useState, useEffect, type ReactNode, useCallback} from 'react';
import {useAsgardeo} from "@asgardeo/react";
import {useNavigate} from 'react-router-dom';
import type {ConsentRecord} from "./types.ts";

interface ConsentContextType {
  consentRecord: ConsentRecord | null;
  handleConsentFetch: () => Promise<void>;
  error: string;
  isSubmitting: boolean;
  consentId: string | null;
  updateConsentId: (consentId: string) => void;
  handleConsentDecision: (decision: 'approved' | 'rejected') => Promise<void>;
}

const ConsentContext = createContext<ConsentContextType | undefined>(undefined);

export const useConsent = () => {
  const context = useContext(ConsentContext);
  if (!context) {
    throw new Error('useConsent must be used within ConsentProvider');
  }
  return context;
};

export const ConsentProvider: React.FC<{ children: ReactNode }> = ({children}) => {
  const navigate = useNavigate();
  const [consentRecord, setConsentRecord] = useState<ConsentRecord | null>(null);
  const [error, setError] = useState('');
  const [consentId, setConsentId] = useState<string | null>(null);
  const [userEmail, setUserEmail] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {isSignedIn, user, getAccessToken} = useAsgardeo();

  const CONSENT_ENGINE_PATH = window?.configs?.apiUrl;
  if (!CONSENT_ENGINE_PATH) {
    throw new Error(
      "Consent Engine API URL is not defined. Please ensure window.configs.apiUrl is set before loading the Consent Portal."
    );
  }

  const getAuthHeaders = async (): Promise<HeadersInit> => {
    const accessToken = await getAccessToken();
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };

    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`;
    }

    return headers;
  };

  const fetchUserInfo = useCallback(() => {
    if (user) {
      setUserEmail(user.email || '');
    }
  }, [user]);

  const fetchConsentData = async (consentUuid: string) => {
    try {
      const headers = await getAuthHeaders();
      const response = await fetch(`${CONSENT_ENGINE_PATH}/consents/${consentUuid}`, {headers});

      if (!response.ok) {
        let errorMessage = '';
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || '';
        } catch { /* empty */
        }

        if (response.status === 403) {
          setError(errorMessage || 'Forbidden: You do not have permission to access this consent');
          navigate('/unauthorized');
          return null;
        } else {
          throw new Error(errorMessage || `Failed to fetch consent data: ${response.status}`);
        }
      }

      const data: ConsentRecord = await response.json();
      setConsentRecord({...data, consent_id: consentUuid});
      return data;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load consent information. Please try again.');
      console.error(err);
      navigate('/error');
      return null;
    }
  };

  const handleConsentFetch = async () => {
    const consentUuid = consentId || localStorage.getItem('consent_id') || '';

    if (!consentUuid) {
      setError('No consent ID provided.');
      navigate('/error');
      return;
    }
    await fetchConsentData(consentUuid);
  };

  const handleConsentDecision = async (decision: 'approved' | 'rejected') => {
    if (!consentRecord) return;

    setIsSubmitting(true);

    try {
      const payload = {
        ...consentRecord,
        status: decision,
        updated_by: userEmail || 'self-service',
        reason: decision === 'rejected'
          ? 'User denied the consent request via self-service portal'
          : 'User approved the consent request via self-service portal'
      };

      const headers = await getAuthHeaders();
      const response = await fetch(`${CONSENT_ENGINE_PATH}/consents/${consentRecord.consent_id}`, {
        method: 'PUT',
        headers,
        body: JSON.stringify(payload)
      });

      if (!response.ok) {
        throw new Error('Failed to update consent');
      }

      navigate('/success');
      localStorage.removeItem('consent_id');

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
      setError(err instanceof Error ? err.message : 'Failed to process your consent decision.');
      navigate('/error');
    } finally {
      setIsSubmitting(false);
    }
  };

  const updateConsentId = (consentId: string) => {
    setConsentId(consentId);
    // Store the consent_id in localStorage for persistence across redirects
    localStorage.setItem('consent_id', consentId);
  }

  useEffect(() => {
    if (isSignedIn) {
      fetchUserInfo();
    }
  }, [fetchUserInfo, isSignedIn, user]);

  return (
    <ConsentContext.Provider
      value={{
        consentRecord,
        updateConsentId,
        error,
        isSubmitting,
        consentId,
        handleConsentDecision,
        handleConsentFetch
      }}
    >
      {children}
    </ConsentContext.Provider>
  );
};
