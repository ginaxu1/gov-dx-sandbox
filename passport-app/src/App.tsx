import React, { useState, useEffect } from 'react';
import { ApolloClient, InMemoryCache, createHttpLink, ApolloProvider, gql, useLazyQuery } from '@apollo/client';
import { AuthProvider, useAuthContext } from '@asgardeo/auth-react';
import PassportForm from './PassportForm';
import './styles.css';

// Apollo Client setup
const httpLink = createHttpLink({
  uri: import.meta.env.GRAPHQL_ENDPOINT,
  headers: {
    "Test-Key": TEST_KEY 
  },
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
            Log in with Asgardeo
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="app-container">
      {/* Add w-full to make it take up the full width, allowing inner content to be centered */}
      <div className="main-content-wrapper w-full"> 
        <header className="app-header">
          <h1 className="app-title">Apply for Passport</h1>
          <button onClick={() => signOut()} className="logout-button">
            Logout
          </button>
        </header>

        <div className="welcome-section">
          <h2 className="welcome-heading">Welcome, {userInfo?.given_name || 'User'}!</h2>
          <p className="welcome-text">
            Click the button below to start your passport application. Your personal data will be retrieved automatically from the Exchange.
          </p>
          <button onClick={() => setShowPassportForm(true)} className="apply-button">
            Apply for Passport
          </button>
        </div>
      </div>

      {showPassportForm && (
        <PassportForm 
          onClose={() => setShowPassportForm(false)} 
          nic="199512345678" // Pass NIC as a prop
          userInfo={userInfo}
        />
      )}
    </div>
  );
}

// Root component with providers
export default function App() {
  const config = {
    signInRedirectURL: "http://localhost:5173",
    signOutRedirectURL: window.location.origin,
    clientID: import.meta.env.VITE_ASGARDEO_CLIENT_ID,
    baseUrl: import.meta.env.VITE_ASGARDEO_BASE_URL,
    scope: [import.meta.env.VITE_ASGARDEO_SCOPE],
  };

  return (
    <AuthProvider config={config}>
      <ApolloProvider client={client}>
        <AppContent />
      </ApolloProvider>
    </AuthProvider>
  );
}