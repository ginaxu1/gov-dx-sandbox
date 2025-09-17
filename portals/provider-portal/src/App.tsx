// App.tsx
// import { useState } from "react";
import { Navbar } from "./components/Navbar";
import { Dashboard } from './pages/Dashboard';
import { Schemas } from './pages/Schemas';
import { Consumers } from './pages/Consumers';
import { SchemaRegistrationPage } from "./pages/SchemaRegistrationPage";
import { Router, type Route } from "./Router";


function App() {
  // const [providerId, setProviderId] = useState<string>("prov_bd7fa213a556e7105677313c");
  // const [providerName, setProviderName] = useState<string>("Department Registrar of Persons");
  
  // const handleProviderIdChange = (id: string) => {
  //   setProviderId(id);
  // };

  // const handleProviderNameChange = (name: string) => {
  //   setProviderName(name);
  // };

  const providerId = "prov_bd7fa213a556e7105677313c";
  const providerName = "Department Registrar of Persons";
  const previousSchemaId = "";

  const SchemaRegistrationRoute: React.FC = () => (
    <SchemaRegistrationPage providerId={providerId} providerName={providerName} previous_schema_id={previousSchemaId} />
  );

  // Fixed route order - more specific routes first
  const routes: Route[] = [
    { path: "/schemas/new", component: SchemaRegistrationRoute, exact: true },
    { path: "/schemas", component: Schemas, exact: true },
    { path: "/consumers", component: Consumers, exact: true },
    { path: "/", component: Dashboard, exact: true },
  ];

  return (
    <div className="App">
      <Navbar />
      <Router routes={routes} />
    </div>
  );
}

export default App;