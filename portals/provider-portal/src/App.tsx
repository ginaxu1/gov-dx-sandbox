import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Navbar } from "./components/Navbar";
import { Dashboard } from './pages/Dashboard';
import { SchemasPage } from './pages/Schemas';
import { SchemaRegistrationPage } from "./pages/SchemaRegistrationPage";
import { Logs } from "./pages/Logs";
import { Applications } from "./pages/Applications";
import { useEffect, useState } from "react";

interface EntityProps {
  id: string;
  name: string;
  providerId?: string;
  consumerId?: string;
}

function App() {
  const [view, setView] = useState<'provider' | 'consumer'>('provider');
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

  const handleViewChange = (newView: 'provider' | 'consumer') => {
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
                    providerId={entityData?.providerId} 
                    providerName={entityData?.name} 
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
              <Route path="*" element={<Navigate to="/" replace />} />
            </>
          )}
        </Routes>
      </div>
    </Router>
  );
}

export default App;