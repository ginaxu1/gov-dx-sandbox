import express from "express";
import consumerRoutes from "./routes/consumerRoutes";
import providerRoutes from "./routes/providerRoutes";

// EXPRESS APPLICATION SETUP
const app = express();
const PORT = 3000;

// Middleware to parse JSON bodies
app.use(express.json());

// API Routes
app.use(consumerRoutes);
app.use(providerRoutes);

// SERVER START
// This conditional logic prevents the server from starting during tests
if (process.env.NODE_ENV !== "test") {
  app.listen(PORT, () => {
    console.log(`Backend server is running on http://localhost:${PORT}`);
  });
}

export default app;
