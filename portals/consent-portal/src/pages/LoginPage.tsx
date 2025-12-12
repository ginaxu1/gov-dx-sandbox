import React, {useEffect} from 'react';
import { Shield } from 'lucide-react';
import {useAsgardeo} from "@asgardeo/react";
import {useNavigate} from "react-router-dom";

const LoginPage: React.FC = () => {

   const navigate = useNavigate();

  const { isSignedIn, signIn } = useAsgardeo()

  useEffect(() => {
    if (isSignedIn) {
      navigate('/');
    }
  }, [isSignedIn, navigate]);

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4 relative">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6 text-center">
        <Shield className="h-12 w-12 text-blue-500 mx-auto mb-4" />
        <h1 className="text-2xl font-bold text-gray-800 mb-4">Consent Portal</h1>
        <p className="text-gray-600 mb-4">
          You need to sign in to process your consent request.
        </p>
        <button
          onClick={() => signIn()}
          className="bg-blue-500 hover:bg-blue-600 text-white px-6 py-3 rounded-lg font-medium transition-colors"
        >
          Sign In to Continue
        </button>
      </div>
    </div>
  );
};

export default LoginPage;