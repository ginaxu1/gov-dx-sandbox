import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthContext } from "@asgardeo/auth-react";
import { Navbar } from "./components/Navbar";
import { Schemas } from './pages/Schemas';
import { Logs } from "./pages/Logs";
import { Applications } from "./pages/Applications";
import { useEffect, useState } from "react";
import { Shield } from 'lucide-react';
import { Dashboard } from './pages/Dashboard';
import { Members } from './pages/Members';

function App() {
  const { state, signIn, signOut, getBasicUserInfo } = useAuthContext();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchEntityInfoFromDB = async (entityId: string) => {
    try {
      // fetch entity info from API
      const response = await fetch(`${window.configs.apiUrl}/entities/${entityId}`);
      if (!response.ok) {
        throw new Error('Failed to fetch entity info');
      }
      const data = await response.json();
      return data;
    } catch (error) {
      console.error('Error fetching entity info:', error);
      return null;
    }
  };

  useEffect(() => {
    const fetchEntityInfo = async () => {
      if (!state.isAuthenticated) {
        return;
      }

      setLoading(true);
      
      try {
        // Fetch fresh data from API
        const userBasicInfo = await getBasicUserInfo();
        console.log('Fetching entity info from user attributes:', userBasicInfo);

        const entityId = userBasicInfo.memberId;
        if (!entityId) {
          setError('User does not have a memberId attribute');
        }

        const fetchedEntityInfoFromDB = await fetchEntityInfoFromDB(entityId);
        if ( !fetchedEntityInfoFromDB || fetchedEntityInfoFromDB.entityType !== 'admin') {
          setError('Failed to fetch valid entity info from DB or user is not an admin');
        } else {
          const entityInfo = fetchedEntityInfoFromDB;
          console.log('Fetched entity info from DB:', entityInfo);
        }
      } catch (error) {
        console.error('Failed to fetch entity info:', error);
        setError('An error occurred while fetching entity info');
      } finally {
        setLoading(false);
      }
    };

    fetchEntityInfo();
  }, [state.isAuthenticated]);

  const handleSignIn = () => {
    signIn();
  };

  const handleSignOut = () => {
    signOut();
  };

  // Show login screen if not authenticated
  if (!state.isAuthenticated) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4 relative">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6 text-center">
          <Shield className="h-12 w-12 text-blue-500 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-gray-800 mb-4">Admin Portal</h1>
          <p className="text-gray-600 mb-4">
            Sign in to access the OpenDIF Admin Portal.
          </p>
          <button 
            onClick={handleSignIn} 
            className="bg-blue-500 hover:bg-blue-600 text-white px-6 py-3 rounded-lg font-medium transition-colors"
          >
            Sign In to Continue
          </button>
        </div>
      </div>
    );
  }

  // Show loading while fetching entity data
  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center relative">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading entity information...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4 relative">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6 text-center">
          <h1 className="text-2xl font-bold text-red-600 mb-4">Error</h1>
          <p className="text-gray-600 mb-4">{error}</p>
          <button 
            onClick={handleSignOut} 
            className="bg-red-500 hover:bg-red-600 text-white px-6 py-3 rounded-lg font-medium transition-colors"
          >
            Sign Out
          </button>
        </div>
      </div>
    );
  }

  return (
    <Router>
      <div className="App">
        <Navbar 
          onSignOut={handleSignOut}
        />
        <Routes>
          <Route path="/" element={<Dashboard/>} />
          <Route path="/members" element={<Members />} />
          <Route path="/schemas" element={<Schemas />} />
          <Route path="/logs" element={<Logs />} />
          <Route path="/applications" element={<Applications />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;