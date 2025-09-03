import { getIntrospectionQuery, buildClientSchema } from "graphql";

async function fetchSchema(endpoint: string) {
  const res = await fetch(endpoint, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query: getIntrospectionQuery() }),
  });

  const json = await res.json();
  const { data } = json;
  const schema = buildClientSchema(data);
  
//   console.log(schema); // GraphQLSchema object

//   console.log(printSchema(schema)); // SDL format
  return schema;
}

export default fetchSchema;