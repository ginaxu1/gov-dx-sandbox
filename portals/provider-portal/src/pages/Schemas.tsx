import React from 'react';

export const Schemas: React.FC = () => {
    const handleCreateNewSchema = () => {
        // Use the global navigate function instead of window.location.href
        if ((window as any).navigate) {
            (window as any).navigate('/schemas/new');
        } else {
            // Fallback if navigate function is not available
            window.history.pushState({}, '', '/schemas/new');
            window.dispatchEvent(new Event('navigate'));
        }
    };

    return (
        <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
            <div className="px-4 py-6 sm:px-0">
                <h1 className="text-3xl font-bold text-gray-900 mb-6">Schemas</h1>
                <p className="text-gray-600 mb-8">You can view schema info here</p>
                
                <div className="bg-white shadow rounded-lg p-6 mb-6">
                    <h2 className="text-xl font-semibold text-gray-900 mb-4">Registered Schemas</h2>
                    <ul className="space-y-2">
                        <li className="p-2 bg-gray-50 rounded">Schema 1</li>
                        <li className="p-2 bg-gray-50 rounded">Schema 2</li>
                        <li className="p-2 bg-gray-50 rounded">Schema 3</li>
                    </ul>
                    <div className="mt-4">
                        <button
                            type="button"
                            onClick={handleCreateNewSchema}
                            className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
                        >
                            Register New Schema
                        </button>
                    </div>
                </div>

                <div className="bg-white shadow rounded-lg p-6 mb-6">
                    <h2 className="text-xl font-semibold text-gray-900 mb-4">Pending Registration</h2>
                    <ul className="space-y-2">
                        <li className="p-2 bg-gray-50 rounded">Schema 1</li>
                        <li className="p-2 bg-gray-50 rounded">Schema 2</li>
                        <li className="p-2 bg-gray-50 rounded">Schema 3</li>
                    </ul>
                </div>

            </div>
        </div>
    );
};