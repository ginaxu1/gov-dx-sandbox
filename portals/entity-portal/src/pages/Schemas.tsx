import { useEffect, useState } from "react";
import { useNavigate } from 'react-router-dom';

interface SchemaProps {
    id: number;
    name: string;
}

export const SchemasPage: React.FC = () => {
    const navigate = useNavigate();
    const [registeredSchemas, setRegisteredSchemas] = useState<SchemaProps[]>([]);
    const [pendingSchemas, setPendingSchemas] = useState<SchemaProps[]>([]);

    useEffect(() => {
        // Fetch registered and pending schemas from the API
        const fetchSchemas = async () => {
            try {
                // Simulate API call
                // Replace this with actual API call
                const fetchedRegisteredSchemas = [
                    { id: 1, name: 'Schema 1' },
                    { id: 2, name: 'Schema 2' },
                    { id: 3, name: 'Schema 3' },
                ];
                const fetchedPendingSchemas = [
                    { id: 4, name: 'Schema 4' },
                    { id: 5, name: 'Schema 5' },
                ];

                setRegisteredSchemas(fetchedRegisteredSchemas);
                setPendingSchemas(fetchedPendingSchemas);
            } catch (error) {
                console.error('Error fetching schemas:', error);

            }
        };

        fetchSchemas();
    }, []);

    const handleCreateNewSchema = () => {
        navigate('/provider/schemas/new');
    };

    return (
        <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
            <div className="px-4 py-6 sm:px-0">
                <h1 className="text-3xl font-bold text-gray-900 mb-6">Schemas</h1>
                <p className="text-gray-600 mb-8">You can view schema info here</p>
                
                <div className="bg-white shadow rounded-lg p-6 mb-6">
                    <h2 className="text-xl font-semibold text-gray-900 mb-4">Registered Schemas</h2>
                    <ul className="space-y-2">
                        {registeredSchemas.map(schema => (
                            <li key={schema.id} className="p-2 bg-gray-50 rounded">{schema.name}</li>
                        ))}
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
                        {pendingSchemas.map(schema => (
                            <li key={schema.id} className="p-2 bg-gray-50 rounded">{schema.name}</li>
                        ))}
                    </ul>
                </div>
            </div>
        </div>
    );
};