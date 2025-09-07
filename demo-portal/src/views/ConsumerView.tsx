import ConsumerRegistrationForm from '../components/ConsumerRegistrationForm';

interface ConsumerViewProps {
    logApiCall: (message: string, status: number | string, response: object) => void;
    consumerState: {
        isSubmitted: boolean;
        appId: string | null;
    };
    handleSubmitApp: (payload: { appId: string, requiredFields: object }) => Promise<void>;
}

export default function ConsumerView({
    logApiCall,
    consumerState,
    handleSubmitApp,
}: ConsumerViewProps) {
    return (
        <div className="space-y-4">
            <h2 className="text-2xl font-semibold">Data Consumer</h2>
            <p className="text-gray-600">Submit an application to access data</p>
            <ConsumerRegistrationForm logApiCall={logApiCall} consumerState={consumerState} handleSubmitApp={handleSubmitApp}/>
        </div>
    );
}