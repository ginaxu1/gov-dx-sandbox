# Passport Application Portal
This is a web application built with React and Vite, designed to streamline the passport application process by automatically fetching and pre-filling personal data, including Full Name, NIC, Date of Birth, Address, Parent Info, and Photo, from an OpenDIF Exchange via a GraphQL API. Additionally, it provides fields for manual entry of supplementary information such as Email Address and Emergency Contact, all within a responsive and user-friendly interface.

## Technologies Used
- React: A JavaScript library for building user interfaces
- Vite: A fast build tool that provides a lightning-fast development experience for modern web projects
- Apollo Client: A comprehensive state management library for GraphQL, enabling efficient data fetching and caching
- Asgardeo/Auth-React: A React SDK for integrating with Asgardeo, an open-source identity and access management (IAM) solution

## Getting Started
Prerequisites: Node.js, npm or Yarn

1. Clone the repository

2. Install Dependencies from `/passport-app`
npm install # or yarn install

3. Configure Asgardeo Authentication
This application uses [Asgardeo](https://wso2.com/asgardeo/) for authentication. Create a .env file in the root of your project and add the following:
```
VITE_ASGARDEO_CLIENT_ID="<YOUR_ASGARDEO_CLIENT_ID>"
VITE_ASGARDEO_BASE_URL="<YOUR_ASGARDEO_BASE_URL>" # e.g., https://api.asgardeo.io/t/<your_tenant_name>
VITE_ASGARDEO_SCOPE="<YOUR_ASGARDEO_SCOPE>" # equired scopes for your Asgardeo application, e.g., "openid profile"
```
Important: In the Asgardeo console, ensure the Asgardeo application's `signInRedirectURL` is set to `http://localhost:5173` (or your chosen development port) and `signOutRedirectURL` to `window.location.origin`, like configured in App.tsx

4. Configure GraphQL API Key: the application interacts with a GraphQL API that requires a Test-Key for authorization

App.tsx (Relevant snippet for API Key):

```
const httpLink = createHttpLink({
  uri: "https://41200aa1-4106-4e6c-babf-311dce37c04a-dev.e1-us-east-azure.choreoapis.dev/gov-dx-sandbox/graphql-resolver-or/v1.0/",
  headers: {
    "Test-Key": "your-actual-test-key", // Replace with your actual key
  },
});
```

Note: Ensure the Test-Key value is replaced with the correct key for your deployed service

5. Running the Application
To start the development server:

npm run dev # or yarn dev

The application will be accessible at http://localhost:5173

6. To create a production-ready build:

npm run build # or yarn build

The optimized static assets will be generated in the dist directory
