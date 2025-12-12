import React from 'react';
import { AlertCircle } from 'lucide-react';
import { useConsent } from '../ConsentContext';
import UserHeader from '../components/UserHeader';
import {useAsgardeo} from "@asgardeo/react";

const ErrorPage: React.FC = () => {
  const { error } = useConsent();

  const {signIn, signOut, user} = useAsgardeo()

  console.log(user)

  return (
    <div className="min-h-screen bg-gradient-to-br from-red-50 to-pink-100 flex items-center justify-center p-4 relative">
      <UserHeader userName={user} onSignIn={() => signIn()} onSignOut={() => signOut()} />
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
};

export default ErrorPage;