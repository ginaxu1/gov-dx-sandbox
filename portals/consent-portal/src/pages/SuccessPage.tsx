import React from 'react';
import { CheckCircle } from 'lucide-react';
import UserHeader from '../components/UserHeader';
import {useAsgardeo} from "@asgardeo/react";

const SuccessPage: React.FC = () => {
  const { signIn, signOut, user } = useAsgardeo();

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-50 to-emerald-100 flex items-center justify-center p-4 relative">
      {user && <UserHeader userName={user.givenName} onSignIn={() => signIn()} onSignOut={() => signOut()} />}
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
};

export default SuccessPage;