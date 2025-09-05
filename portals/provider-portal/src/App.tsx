import { useState} from "react";
import { SchemaRegistrationPage } from "./pages/SchemaRegistrationPage";

function App() {
  const [providerId, setProviderId] = useState<string>("prov_bd7fa213a556e7105677313c");
  const [providerName, setProviderName] = useState<string>("Department Registrar of Persons");
  const handleProviderIdChange = (id: string) => {
    setProviderId(id);
  };

  const handleProviderNameChange = (name: string) => {
    setProviderName(name);
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen py-2 w-full">
      <h1 className="text-5xl font-bold">Welcome to the Provider Portal</h1>
      <SchemaRegistrationPage providerId={providerId} providerName={providerName} />
    </div>
  )
}

export default App
