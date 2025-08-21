// App.tsx
import React, { useState, useEffect } from 'react';
import { ApolloClient, InMemoryCache, createHttpLink, ApolloProvider } from '@apollo/client';
import { useAuthContext } from '@asgardeo/auth-react';
import PassportForm from './PassportForm';
import './styles.css';

// Extend Window interface to include configs
declare global {
  interface Window {
    configs?: {
      apiUrl?: string;
    };
  }
}

window.configs = {
    apiUrl: '/choreo-apis/gov-dx-sandbox/graphql-resolver/v1'
};
const apiUrl = window?.configs?.apiUrl ? window.configs.apiUrl : "/";


// Apollo Client setup
const httpLink = createHttpLink({
  uri: apiUrl
});

const client = new ApolloClient({
  link: httpLink,
  cache: new InMemoryCache(),
});

function AppContent() {
  const { state, signIn, signOut, getBasicUserInfo } = useAuthContext();
  const [userInfo, setUserInfo] = useState(null);
  const [showPassportForm, setShowPassportForm] = useState(false);

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
  );
}

// Root component with providers
export default function App() {
  return (
    <ApolloProvider client={client}>
      <AppContent />
    </ApolloProvider>
  );
}