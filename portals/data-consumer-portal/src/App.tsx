import { useState } from 'react';
import './index.css';
import ConsumerRegistrationForm from './components/RegistrationForm';

const App = () => {
    const [showRegistrationForm, setShowRegistrationForm] = useState(false);

    return (
        <div className="flex flex-col items-center justify-center min-h-screen p-4 bg-background text-foreground">
            <div className="w-full max-w-xl p-8 space-y-6 bg-card rounded-xl shadow-2xl text-center">
                <div className="space-y-2">
                    <h1 className="text-4xl font-bold text-card-foreground">Data Consumer Portal</h1>
                    <p className="text-muted-foreground">
                        Register your application to access data through the exchange
                    </p>
                </div>

                {showRegistrationForm ? (
                    <ConsumerRegistrationForm />
                ) : (
                    <div className="pt-4">
                        <button
                            onClick={() => setShowRegistrationForm(true)}
                            className="inline-flex items-center justify-center w-full sm:w-auto px-8 py-3 bg-primary text-primary-foreground font-semibold rounded-lg shadow-md hover:opacity-90 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-opacity-75 transform hover:-translate-y-1 transition-all duration-300"
                        >
                            Register a New Application
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
};

export default App;
