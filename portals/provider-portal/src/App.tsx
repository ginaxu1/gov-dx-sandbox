import fetchSchema from "./services/graphql_introspection";
import { useState, useEffect} from "react";
import { printSchema} from "graphql";

function App() {
  const [schema, setSchema] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      const result = await fetchSchema("http://localhost:9091");
      // To fetch the schema
      // You need to start the mock-drp service (/samples/lk-passport-application/mock-services/mock-drp) + drp service (/old/provider-wrappers/drp)
      setSchema(printSchema(result));
    };
    fetchData();
  }, []);

  return (
    <div className="flex flex-col items-center justify-center min-h-screen py-2">
      <h1 className="text-5xl font-bold">Welcome to the Provider Portal</h1>
      <div className="flex flex-col items-center">
        <h2 className="text-2xl font-bold">Schema</h2>
        <pre className="whitespace-pre-wrap">{schema}</pre>
      </div>
    </div>
  )
}

export default App
