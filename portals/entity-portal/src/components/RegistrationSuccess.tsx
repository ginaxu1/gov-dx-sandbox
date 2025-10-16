import React, { useEffect } from 'react';
import { CheckCircle, ArrowRight, Clock } from 'lucide-react';

interface RegistrationSuccessProps {
  type: 'schema' | 'application';
  title: string;
  onRedirect: () => void;
}

export const RegistrationSuccess: React.FC<RegistrationSuccessProps> = ({
  type,
  title,
  onRedirect
}) => {
  useEffect(() => {
    const timer = setTimeout(() => {
      onRedirect();
    }, 20000);

    return () => clearTimeout(timer);
  }, [onRedirect]);

  const redirectPath = type === 'schema' ? 'provider/schemas' : 'consumer/applications';

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-50 to-emerald-100 flex items-center justify-center">
      <div className="max-w-md w-full mx-4">
        <div className="bg-white shadow-2xl rounded-2xl overflow-hidden">
          {/* Header */}
          <div className="bg-gradient-to-r from-green-600 to-green-700 px-8 py-6 text-center">
            <div className="flex justify-center mb-4">
              <div className="w-16 h-16 bg-white rounded-full flex items-center justify-center">
                <CheckCircle className="w-8 h-8 text-green-600" />
              </div>
            </div>
            <h1 className="text-2xl font-bold text-white mb-2">
              Registration Successful!
            </h1>
            <p className="text-green-100">
              Your {type} has been submitted successfully
            </p>
          </div>

          {/* Content */}
          <div className="px-8 py-6">
            <div className="text-center mb-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-2">
                "{title}"
              </h2>
              <p className="text-gray-600">
                Your {type} registration has been submitted for review. 
                You'll be notified once it's approved.
              </p>
            </div>

            {/* Redirect Notice */}
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
              <div className="flex items-center">
                <Clock className="w-5 h-5 text-blue-600 mr-3 flex-shrink-0" />
                <div className="text-sm">
                  <p className="text-blue-800 font-medium mb-1">
                    Redirecting automatically...
                  </p>
                  <p className="text-blue-600">
                    You'll be redirected to {redirectPath} in a few seconds
                  </p>
                </div>
              </div>
            </div>

            {/* Manual Redirect Button */}
            <button
              onClick={() => {
                console.log('Manual redirect button clicked');
                onRedirect();
              }}
              className="w-full bg-gradient-to-r from-blue-600 to-blue-700 text-white py-3 px-4 rounded-lg hover:from-blue-700 hover:to-blue-800 transition-all duration-200 font-semibold shadow-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 flex items-center justify-center"
            >
              Go to {type === 'schema' ? 'Schemas' : 'Applications'}
              <ArrowRight className="w-4 h-4 ml-2" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};