// App.tsx
import React, { useState, useEffect } from 'react';
import { ApolloClient, InMemoryCache, createHttpLink, ApolloProvider, from } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';
import { useAuthContext } from '@asgardeo/auth-react';
import PassportForm from './PassportForm';
import './styles.css';

// Extend Window interface to include configs
declare global {
  interface Window {
    configs?: {
      apiUrl?: string;
      ndxApiKey?: string;
    };
  }
}

// Choreo injects environment variables into the window.configs object
const apiUrl = window?.configs?.apiUrl;
const ndx_api_key = window?.configs?.ndxApiKey;

function AppContent() {
  const { state, signIn, signOut, getBasicUserInfo, getAccessToken } = useAuthContext();
  const [userInfo, setUserInfo] = useState(null);
  const [showPassportForm, setShowPassportForm] = useState(false);

  console.log(">>> Apollo Client URI:", apiUrl);
  console.log(">>> Ndx_API Key:", ndx_api_key);  

  // Set up Apollo Client within the component to access getAccessToken
  const authLink = setContext(async (_, { headers }) => {
    try {
      const accessToken = await getAccessToken();
      return {
        headers: {
          ...headers,
          "Authorization": `Bearer ${accessToken}`,
          "Test-Key": ndx_api_key, 
        }
      };
    } catch (error) {
      console.error("Error getting access token:", error);
      return { headers };
    }
  });

  const httpLink = createHttpLink({
    uri: apiUrl,
  });

  const client = new ApolloClient({
    link: from([authLink, httpLink]),
    cache: new InMemoryCache(),
  });

  useEffect(() => {
    if (state.isAuthenticated) {
      getBasicUserInfo()
        .then(info => {
          setUserInfo(info);
        })
        .catch(err => console.error("Error fetching basic user info:", err));
    }
  }, [state.isAuthenticated, getBasicUserInfo]);

  if (!state.isAuthenticated) {
    return (
      <div className="login-container">
        <div className="login-card">
          <h1 className="login-heading">Passport Application Portal</h1>
          <p className="login-text">Please log in to apply for a passport.</p>
          <button onClick={() => signIn()} className="login-button">
            Log in
          </button>
        </div>
      </div>
    );
  }

return (
    <ApolloProvider client={client}>
      <div className="app-container">
        <div className="main-content-wrapper">
          <header className="app-header">
            <h1 className="app-title">Apply for Passport</h1>
            <button onClick={() => signOut()} className="logout-button">
              Logout
            </button>
          </header>

          <div className="welcome-section">
            <h2 className="welcome-heading">Welcome, {userInfo?.given_name || 'User'}!</h2>
            <p className="welcome-text">
              Click the button below to start your passport application. Your personal data will be retrieved automatically.
            </p>
            <button onClick={() => setShowPassportForm(true)} className="apply-button">
              Apply for Passport
            </button>
          </div>
        </div>

        {showPassportForm && (
          <PassportForm
            onClose={() => setShowPassportForm(false)}
            nic="199512345678"
            userInfo={userInfo}
          />
        )}
      </div>
    </ApolloProvider>
  );
}
// Root component with providers
export default function App() {
  const { state } = useAuthContext();

  // If not logged in, show the login button and ApolloProvider will be initialized after login
  if (!state.isAuthenticated) {
    return (
      <div className="login-container">
        <div className="login-card">
          <h1 className="login-heading">Passport Application Portal</h1>
          <p className="login-text">Please log in to apply for a passport</p>
          <button onClick={() => useAuthContext().signIn()} className="login-button">
            Log in
          </button>
        </div>
      </div>
    );
  }
  return (
    <AppContent />
  );
}