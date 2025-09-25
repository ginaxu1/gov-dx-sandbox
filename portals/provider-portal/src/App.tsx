// App.tsx
import { Navbar } from "./components/Navbar";
import { Dashboard } from './pages/Dashboard';
import { SchemasPage as Schemas } from './pages/Schemas';
import { SchemaRegistrationPage } from "./pages/SchemaRegistrationPage";
import { Logs } from "./pages/Logs";
import { Router, type Route } from "./Router";
import { useEffect, useState } from "react";

interface EntityProps {
  id: string;
  name: string;
  providerId?: string;
  consumerId?: string;
}

function App() {
  const [view, setView] = useState<'provider' | 'consumer'>('provider');
  const [entityData, setEntityData] = useState<any>(null);

  // const providerId = "prov_bd7fa213a556e7105677313c";
  // const providerName = "Department Registrar of Persons";

  useEffect(() => {
    // Fetch entity type and ID from an API or configuration
    const fetchEntityInfo = async () => {
      // Simulate an API call
      // Replace this with actual API call to fetch entity info
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
  }

  const SchemaRegistrationRoute: React.FC = () => (
    <SchemaRegistrationPage providerId={entityData?.providerId} providerName={entityData?.name} />
  );

  // Fixed route order - more specific routes first
  const routes: Route[] = [
    { path: "/schemas/new", component: SchemaRegistrationRoute, exact: true },
    { path: "/schemas", component: Schemas, exact: true },
    { path: "/logs", component: Logs, exact: true },
    { path: "/", component: Dashboard, exact: true },
  ];

  return (
    <div className="App">
      <Navbar onViewChange={handleViewChange} providerId={entityData?.providerId} consumerId={entityData?.consumerId} />
      <Router routes={routes} />
    </div>
  );
}

export default App;