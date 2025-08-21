// App.tsx
import React, { useState, useEffect } from 'react';
import { ApolloClient, InMemoryCache, createHttpLink, ApolloProvider } from '@apollo/client';
import { useAuthContext } from '@asgardeo/auth-react';
import PassportForm from './PassportForm';
import './styles.css';

declare global {
  interface Window {
    configs?: {
      apiUrl?: string;
      ndxApiKey?: string; 
    };
  }
}


window.configs = {
    apiUrl: '/choreo-apis/gov-dx-sandbox/graphql-resolver/v1'
};
const apiUrl = window?.configs?.apiUrl ? window.configs.apiUrl : "/";
const ndx_api_key = window?.configs?.ndxApiKey;

// Apollo Client setup
const httpLink = createHttpLink({
  uri: apiUrl,
  headers: {
    // Use the API key from the Choreo injected config
    "Test-Key": ndx_api_key || "eyJraWQiOiJnYXRld2F5X2NlcnRpZmljYXRlX2FsaWFzIiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJkMThhMmVmMi03ZmMxLTRjMzItOTY4My03NjA1YjFkNWQ5NzlAY2FyYm9uLnN1cGVyIiwiYXVkIjoiY2hvcmVvOmRlcGxveW1lbnQ6c2FuZGJveCIsIm9yZ2FuaXphdGlvbiI6eyJ1dWlkIjoiNDEyMDBhYTEtNDEwNi00ZTZjLWJhYmYtMzExZGNlMzdjMDRhIn0sImlzcyI6Imh0dHBzOlwvXC9zdHMuY2hvcmVvLmRldjo0NDNcL2FwaVwvYW1cL3B1Ymxpc2hlclwvdjJcL2FwaXNcL2ludGVybmFsLWtleSIsImtleXR5cGUiOiJTQU5EQk9YIiwic3Vic2NyaWJlZEFQSXMiOlt7InN1YnNjcmliZXJUZW5hbnREb21haW4iOm51bGwsIm5hbWUiOiJncmFwaHFsLXJlc29sdmVyIC0gZ3JhcGhxbC1yZXNvbHZlciIsImNvbnRleHQiOiJcLzQxMjAwYWExLTQxMDYtNGU2Yy1iYWJmLTMxMWRjZTM3YzA0YVwvZ292LWR4LXNhbmRib3hcL2dyYXBocWwtcmVzb2x2ZXJcL3YxLjAiLCJwdWJsaXNoZXIiOiJjaG9yZW9fcHJvZF9hcGltX2FkbWluIiwidmVyc2lvbiI6InYxLjAiLCJzdWJzY3JpcHRpb25UaWVyIjpudWxsfV0sImV4cCI6MTc1NTc2NTYxMiwidG9rZW5fdHlwZSI6IkludGVybmFsS2V5IiwiaWF0IjoxNzU1NzY1MDEyLCJqdGkiOiIzMjg0YTA3ZC04NTk3LTQwNzAtYTVjOS02NTIyYWQwNmM5MmIifQ.UqsBQ8LR7xgdUcMr5QxILxgIHlJHCXj0mIj0zBIo1yrVtGF_pqEh9Y7RFytaVJdQl6QWtVUN_xiFn0iNWYFhXWU2xWkYl3JkNJ5EdcUN_V-oyGZx6kvNL4H_ZmQDFbDnfP_6zqjdpSY2o1HfpahD4WNyUpi6JGvVJP78UxnijRXSTedv86TYKXe4WerV68wIJniXdAUYfeLsTDYkokGkkX1K8iBG4aBSF8XajnKS9YNuQgRK6LM0cwISwmBvTk91x--ST5YFZ05Jwz1aO-OlN1hB_qcv8XXcyBSg9Iitw89cLQ5fLFS0JayHG_bWZWiuAUefRoyrDVDQC6ZjTEfSvVpH5eNy8Yo9zvQpEKghNHG_GK03yi7PsSZbnMK0Deo4xMGqf1MWilRr-iq0Zh6qxObWtnI1wjN2pmuZC7TK-N6zndwrQXJCtRjunGhMmPSx3HezzHvMggUGbL1fSXG2NebPMHa-t37VSltcpCs8egSbnQZ3-ksQ5rUIEp9-cvmKAPFC5NAiffmDO0JnCYBpk4_Jj8KJdlzE0NI1Y1xXxx6HaExP5fZeaPs7xKkuT7vo4BHNwL-lLXJF2qH5arWHTEXOrfVJRRs56OlSeVAAL6tOoLWk5Su4rofOCR9eGJihwJ4JN5S-iGdSqI25fCuQ-Lsl6-UDza-Otg5YDu9AfbY",
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
          <h2 className="welcome-heading">Welcome {userInfo?.given_name || 'User'}!</h2>
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