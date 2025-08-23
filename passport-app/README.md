# Passport Application Portal

This is a web application built with React and Vite, designed to streamline the passport application process by automatically fetching and pre-filling personal data from an OpenDIF Exchange via a GraphQL API. It provides a user-friendly interface for both pre-filled data and manual entry of supplementary information.

### Technologies Used

  * **React:** A JavaScript library for building user interfaces.
  * **Vite:** A fast build tool for a lightning-fast development experience.
  * **Apollo Client:** A comprehensive state management library for GraphQL.
  * **Asgardeo/Auth-React:** A React SDK for integrating with Asgardeo, an open-source Identity and Access Management (IAM) solution.

### Getting Started

Prerequisites: Node.js, npm, or Yarn

1.  Clone the repository.

2.  Install Dependencies from `/passport-app` with `npm install`

### Configuration

This application uses environment variables for both local development and Choreo deployments.

#### Local Development (`.env.local`)

Create a file named **`.env.local`** in the root of your project and fill in the following environment variables. This file should **not** be committed to your repository. A sample file named `.env.example` is provided for reference.

```
VITE_ASGARDEO_CLIENT_ID="<YOUR_ASGARDEO_CLIENT_ID>"
VITE_ASGARDEO_BASE_URL="<YOUR_ASGARDEO_BASE_URL>" # e.g., https://api.asgardeo.io/t/<your_tenant_name>
VITE_ASGARDEO_SCOPE="openid profile"

VITE_NDX_URL="/choreo-apis/gov-dx-sandbox/graphql-resolver/v1.0"
VITE_NDX_API_KEY="<YOUR_CHOREO_API_KEY>"
```

**Important:** For local development, ensure `http://localhost:5173` is registered as a valid **Allowed Redirect URL** in your Asgardeo application settings.

#### Choreo Deployment

When deploying to Choreo, you must add these variables in the **Manage Configs and Secrets** section of your frontend component. The values for `VITE_NDX_URL` and `VITE_NDX_API_KEY` will be provided by your deployed Choreo Gateway.

### Running the Application

  * **To start the development server:**

    ```bash
    npm run dev
    ```

    The application will be accessible at `http://localhost:5173`.

  * **To create a production-ready build:**

    ```bash
    npm run build
    ```

    The optimized static assets will be generated in the `dist` directory.