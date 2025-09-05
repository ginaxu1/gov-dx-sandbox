import React from 'react';

export const LoadingSpinner = ({ text }: { text: string }) => (
    <div className="flex flex-col items-center justify-center p-8 text-center">
        <svg className="animate-spin h-8 w-8 text-blue-500 mb-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <p className="text-gray-500">{text}</p>
    </div>
);

export const ErrorMessage = ({ message, onRetry }: { message: string; onRetry?: () => void }) => (
  <div className="text-center p-6 bg-red-100 text-red-800 rounded-lg">
    <h3 className="text-xl font-semibold">An Error Occurred</h3>
    <p>{message}</p>
    {onRetry && (
      <button onClick={onRetry} className="mt-4 px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700">
        Try Again
      </button>
    )}
  </div>
);

export const SuccessMessage = () => (
    <div className="text-center p-6 bg-green-100 text-green-800 rounded-lg">
        <h3 className="text-xl font-semibold">Schema Submitted!</h3>
        <p>Your schema has been submitted for review.</p>
        <p className="mt-2">Go to the **Admin** view to approve it.</p>
    </div>
);
