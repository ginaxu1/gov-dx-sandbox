import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthContext } from "@asgardeo/auth-react";
import { Navbar } from "./components/Navbar";
import { Dashboard } from './pages/Dashboard';
import { SchemasPage } from './pages/Schemas';
import { SchemaRegistrationPage } from "./pages/SchemaRegistrationPage";
import { Logs } from "./pages/Logs";
import { ApplicationsPage as Applications } from "./pages/Applications";
import { useEffect, useState } from "react";
import { ApplicationRegistration } from './pages/ApplicationRegistration';
import { Shield } from 'lucide-react';

interface EntityProps {
  id: string;
  name: string;
  userName: string;
  providerId?: string;
  consumerId?: string;
}

function App() {
  const [view, setView] = useState<'provider' | 'consumer' | null>(null);
  const [entityData, setEntityData] = useState<EntityProps | null>(null);
  // const [userName, setUserName] = useState<string>('');
  // const [userEmail, setUserEmail] = useState<string>('');

  const { state, signIn, signOut, getBasicUserInfo } = useAuthContext();

  // Fetch user info when authenticated
  // const fetchUserInfo = async () => {
  //   try {
  //     const userBasicInfo = await getBasicUserInfo();
  //     console.log('User Basic Info:', userBasicInfo);
      
  //     if (userBasicInfo) {
  //       setUserName(userBasicInfo.name || '');
  //       setUserEmail(userBasicInfo.email || '');
  //     }
  //   } catch (error) {
  //     console.error('Failed to fetch user info:', error);
  //   }
  // };

  // Save entity state to localStorage to persist through auth redirects
  const saveEntityStateToStorage = (entityInfo: EntityProps, viewType: 'provider' | 'consumer' | null) => {
    localStorage.setItem('entity_data', JSON.stringify(entityInfo));
    if (viewType) {
      localStorage.setItem('entity_view', viewType);
    }
  };

  // Get entity state from localStorage
  const getEntityStateFromStorage = (): { entityData: EntityProps | null, view: 'provider' | 'consumer' | null } => {
    try {
      const entityDataStr = localStorage.getItem('entity_data');
      const viewStr = localStorage.getItem('entity_view');
      
      return {
        entityData: entityDataStr ? JSON.parse(entityDataStr) : null,
        view: viewStr as 'provider' | 'consumer' | null
      };
    } catch (error) {
      console.error('Failed to parse stored entity data:', error);
      return { entityData: null, view: null };
    }
  };

  // Clear entity state from localStorage
  const clearEntityStateFromStorage = () => {
    localStorage.removeItem('entity_data');
    localStorage.removeItem('entity_view');
  };

  useEffect(() => {
    const fetchEntityInfo = async () => {
      try {
        const userBasicInfo = await getBasicUserInfo();
        console.log('Fetching entity info from user attributes:', userBasicInfo);
        
        if (userBasicInfo) {
          const entityInfo: EntityProps = {
            id: userBasicInfo.userid || '', // User ID -> id
            name: userBasicInfo.displayName || userBasicInfo.name || '', // Last Name -> name
            userName: userBasicInfo.username || '', // Email -> userEmail
            providerId: userBasicInfo.providerId || '', // Provider ID -> providerId
            consumerId: userBasicInfo.consumerId || '', // ConsumerID -> consumerId
          };
          
          console.log('Parsed entity info:', entityInfo);
          setEntityData(entityInfo);
          
          // Determine initial view based on available IDs
          let initialView: 'provider' | 'consumer' | null = null;
          if (entityInfo.providerId) {
            initialView = 'provider';
          } else if (entityInfo.consumerId) {
            initialView = 'consumer';
          }
          setView(initialView);
          
          // Save to localStorage for auth redirect recovery
          saveEntityStateToStorage(entityInfo, initialView);
        }
      } catch (error) {
        console.error('Failed to fetch entity info:', error);
        // Fallback to empty entity data if fetch fails
        setEntityData({
          id: '',
          name: '',
          userName: '',
          providerId: undefined,
          consumerId: undefined,
        });
      }
    };

    // Check if we have stored entity data (after auth redirect)
    if (state.isAuthenticated && !entityData) {
      const storedState = getEntityStateFromStorage();
      if (storedState.entityData) {
        setEntityData(storedState.entityData);
        setView(storedState.view);
      } else {
        fetchEntityInfo();
      }
    }
  }, [state.isAuthenticated]);

  // Remove this useEffect - it's causing the circular dependency
  // useEffect(() => {
  //   if (entityData) {
  //     if (entityData.providerId) {
  //       setView('provider');
  //     } else if (entityData.consumerId) {
  //       setView('consumer');
  //     } else {
  //       setView(null);
  //     }
  //   }
  // }, [entityData]);

  // Update user info when authentication state changes
  // useEffect(() => {
  //   if (state.isAuthenticated) {
  //     fetchUserInfo();
  //   }
  // }, [state.isAuthenticated]);

  const canSwitchView = () => {
    return entityData?.providerId && entityData?.consumerId;
  };

  const handleViewChange = (newView: 'provider' | 'consumer') => {
    if (!canSwitchView() && newView !== view) {
      alert(`Cannot switch to ${newView} view. You're not registered as a ${newView === 'provider' ? 'provider' : 'consumer'}.`);
      return;
    }
    setView(newView);
    
    // Update stored view
    if (entityData) {
      saveEntityStateToStorage(entityData, newView);
    }
  };

  // Handle sign in with state preservation (like consent-portal)
  const handleSignIn = () => {
    // Ensure entity state is saved before redirect
    if (entityData) {
      saveEntityStateToStorage(entityData, view);
    }
    signIn();
  };

  // Handle sign out with state cleanup (like consent-portal)
  const handleSignOut = () => {
    clearEntityStateFromStorage();
    signOut();
  };

  // Show login screen if not authenticated
  if (!state.isAuthenticated) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4 relative">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6 text-center">
          <Shield className="h-12 w-12 text-blue-500 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-gray-800 mb-4">Entity Portal</h1>
          <p className="text-gray-600 mb-4">
            Sign in to access the OpenDIF Entity Portal.
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
  if (!entityData) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center relative">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading entity information...</p>
        </div>
      </div>
    );
  }

  return (
    <Router>
      <div className="App">
        <Navbar 
          onViewChange={handleViewChange} 
          providerId={entityData?.providerId} 
          consumerId={entityData?.consumerId}
          currentView={view}
          userName={entityData.name}
          onSignOut={handleSignOut}
        />
        <Routes>
          {view === 'provider' ? (
            <>
              <Route path="/" element={<Dashboard />} />
              <Route path="/provider/schemas" element={<SchemasPage />} />
              <Route 
                path="/provider/schemas/new" 
                element={
                  <SchemaRegistrationPage 
                    providerId={entityData?.providerId || ''}
                    providerName={entityData?.name || ''}
                  />
                } 
              />
              <Route path="/provider/logs" element={<Logs />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </>
          ) : (
            <>
              <Route path="/" element={<Dashboard />} />
              <Route path="/consumer/applications" element={<Applications />} />
              <Route 
                path="/consumer/applications/new" 
                element={
                  <ApplicationRegistration 
                    consumerId={entityData?.consumerId || ''}
                  />
                } 
              />
              <Route path="*" element={<Navigate to="/" replace />} />
            </>
          )}
        </Routes>
      </div>
    </Router>
  );
}

export default App;