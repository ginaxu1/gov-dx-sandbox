import { useEffect, useState } from "react";
import { useNavigate } from 'react-router-dom';

interface ApplicationProps {
    id: number;
    name: string;
}

export const ApplicationsPage: React.FC = () => {
    const navigate = useNavigate();
    const [registeredApplications, setRegisteredApplications] = useState<ApplicationProps[]>([]);
    const [pendingApplications, setPendingApplications] = useState<ApplicationProps[]>([]);

    useEffect(() => {
        // Fetch registered and pending applications from the API
        const fetchApplications = async () => {
            try {
                // Simulate API call
                // Replace this with actual API call
                const fetchedRegisteredApplications = [
                    { id: 1, name: 'Application 1' },
                    { id: 2, name: 'Application 2' },
                    { id: 3, name: 'Application 3' },
                ];
                const fetchedPendingApplications = [
                    { id: 4, name: 'Application 4' },
                    { id: 5, name: 'Application 5' },
                ];

                setRegisteredApplications(fetchedRegisteredApplications);
                setPendingApplications(fetchedPendingApplications);
            } catch (error) {
                console.error('Error fetching applications:', error);

            }
        };

        fetchApplications();
    }, []);

    const handleCreateNewApplication = () => {
        navigate('/consumer/applications/new');
    };

    return (
        <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
            <div className="px-4 py-6 sm:px-0">
                <h1 className="text-3xl font-bold text-gray-900 mb-6">Applications</h1>
                <p className="text-gray-600 mb-8">You can view application info here</p>

                <div className="bg-white shadow rounded-lg p-6 mb-6">
                    <h2 className="text-xl font-semibold text-gray-900 mb-4">Registered Applications</h2>
                    <ul className="space-y-2">
                        {registeredApplications.map(application => (
                            <li key={application.id} className="p-2 bg-gray-50 rounded">{application.name}</li>
                        ))}
                    </ul>
                    <div className="mt-4">
                        <button
                            type="button"
                            onClick={handleCreateNewApplication}
                            className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
                        >
                            Register New Application
                        </button>
                    </div>
                </div>

                <div className="bg-white shadow rounded-lg p-6 mb-6">
                    <h2 className="text-xl font-semibold text-gray-900 mb-4">Pending Registration</h2>
                    <ul className="space-y-2">
                        {pendingApplications.map(application => (
                            <li key={application.id} className="p-2 bg-gray-50 rounded">{application.name}</li>
                        ))}
                    </ul>
                </div>
            </div>
        </div>
    );
};