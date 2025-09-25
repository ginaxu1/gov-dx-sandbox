import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Navbar } from "./components/Navbar";
import { Dashboard } from './pages/Dashboard';
import { SchemasPage } from './pages/Schemas';
import { SchemaRegistrationPage } from "./pages/SchemaRegistrationPage";
import { Logs } from "./pages/Logs";
import { ApplicationsPage as Applications } from "./pages/Applications";
import { useEffect, useState } from "react";
import { ApplicationRegistration } from './pages/ApplicationRegistration';

interface EntityProps {
  id: string;
  name: string;
  providerId?: string;
  consumerId?: string;
}

function App() {
  const [view, setView] = useState<'provider' | 'consumer' | null>(null);
  const [entityData, setEntityData] = useState<EntityProps | null>(null);

  useEffect(() => {
    const fetchEntityInfo = async () => {
      const entityInfo: EntityProps = {
        id: "prov_bd7fa213a556e7105677313c",
        name: "Department Registrar of Persons",
        providerId: "prov_bd7fa213a556e7105677313c",
        consumerId: "cons_1234567890abcdef",
      };
      setEntityData(entityInfo);
    };

    fetchEntityInfo();
  }, []);

  useEffect(() => {
    if (entityData) {
      if (entityData.providerId) {
        setView('provider');
      } else if (entityData.consumerId) {
        setView('consumer');
      } else {
        setView(null);
      }
    }
  }, [entityData]);

  const canSwitchView = () => {
    return entityData?.providerId && entityData?.consumerId;
  };

  const handleViewChange = (newView: 'provider' | 'consumer') => {
    if (!canSwitchView() && newView !== view) {
      alert(`Cannot switch to ${newView} view. You're not registered as a ${newView === 'provider' ? 'provider' : 'consumer'}.`);
      return;
    }
    setView(newView);
  };

  return (
    <Router>
      <div className="App">
        <Navbar 
          onViewChange={handleViewChange} 
          providerId={entityData?.providerId} 
          consumerId={entityData?.consumerId}
          currentView={view}
        />
        <Routes>
          {view === 'provider' ? (
            <>
              <Route path="/" element={<Dashboard />} />
              <Route path="/provider/schemas" element={<SchemasPage />} />
              <Route 
                path="/provider/schemas/new" 
                element={
                  <SchemaRegistrationPage 
                    providerId={entityData?.providerId || ''}
                    providerName={entityData?.name || ''}
                  />
                } 
              />
              <Route path="/provider/logs" element={<Logs />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </>
          ) : (
            <>
              <Route path="/" element={<Dashboard />} />
              <Route path="/consumer/applications" element={<Applications />} />
              <Route path="/consumer/applications/new" element={<ApplicationRegistration 
              />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </>
          )}
        </Routes>
      </div>
    </Router>
  );
}

export default App;