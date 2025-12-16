
import { Route, Routes } from 'react-router-dom';
import ConsentPage from "./pages/ConsentPage.tsx";
import ErrorPage from "./pages/ErrorPage.tsx";
import LoginPage from "./pages/LoginPage.tsx";
import SuccessPage from "./pages/SuccessPage.tsx";
import UnauthorizedPage from "./pages/UnauthorizedPage.tsx";

// Extend Window interface to include config
declare global {
  interface Window {
    configs: {
      apiUrl: string;
      VITE_CLIENT_ID: string;
      VITE_BASE_URL: string;
      VITE_SCOPE: string;
      signInRedirectURL: string;
      signOutRedirectURL: string;
      organizationHandle: string;
    };
  }
}

function App() {
  return (
    <Routes>
      <Route path="/" element={<ConsentPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/error" element={<ErrorPage />} />
      <Route path="/success" element={<SuccessPage />} />
      <Route path="/unauthorized" element={<UnauthorizedPage />} />
    </Routes>
  )
}

export default App;