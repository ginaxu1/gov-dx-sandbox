import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthContext } from "@asgardeo/auth-react";
import { Navbar } from "./components/Navbar";
import { SchemasPage } from './pages/Schemas';
import { SchemaRegistrationPage } from "./pages/SchemaRegistrationPage";
import { Logs } from "./pages/Logs";
import { ApplicationsPage as Applications } from "./pages/Applications";
import { useEffect, useState } from "react";
import { ApplicationRegistration } from './pages/ApplicationRegistration';
import { Shield } from 'lucide-react';
import MemberInfo  from './pages/MemberInfo';

interface EntityProps {
  entityId: string;
  name: string;
  email: string;
  entityType: 'gov' | 'business' | '';
  phoneNumber: string;
  providerId?: string;
  consumerId?: string;
  createdAt: string;
  updatedAt: string;
  roles: Array<'provider' | 'consumer'>;
}

function App() {
  const [view, setView] = useState<'provider' | 'consumer' | null>(null);
  const [entityData, setEntityData] = useState<EntityProps | null>(null);
  const { state, signIn, signOut, getBasicUserInfo } = useAuthContext();
  const [loading, setLoading] = useState(false);

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
        // First check localStorage for existing data
        const storedState = getEntityStateFromStorage();
        if (storedState.entityData) {
          console.log('Loading entity data from storage');
          setEntityData(storedState.entityData);
          setView(storedState.view);
          return;
        }

        // Fetch fresh data from API
        const userBasicInfo = await getBasicUserInfo();
        console.log('Fetching entity info from user attributes:', userBasicInfo);

        const entityId = userBasicInfo.memberId;
        if (!entityId) {
          throw new Error('User does not have a memberId attribute');
        }

        const fetchedEntityInfoFromDB = await fetchEntityInfoFromDB(entityId);
        if (fetchedEntityInfoFromDB) {
          const entityInfo: EntityProps = {
            entityId: fetchedEntityInfoFromDB.entityId || '',
            name: fetchedEntityInfoFromDB.name || '',
            email: fetchedEntityInfoFromDB.email || '',
            entityType: fetchedEntityInfoFromDB.entityType || '',
            phoneNumber: fetchedEntityInfoFromDB.phoneNumber || '',
            providerId: fetchedEntityInfoFromDB.providerId || '',
            consumerId: fetchedEntityInfoFromDB.consumerId || '',
            createdAt: fetchedEntityInfoFromDB.createdAt || '',
            updatedAt: fetchedEntityInfoFromDB.updatedAt || '',
            roles: []
          };
          
          // Determine roles based on IDs
          if (entityInfo.consumerId !== '') {
            entityInfo.roles.push('consumer');
          }
          if (entityInfo.providerId !== '') {
            entityInfo.roles.push('provider');
          }
          
          console.log('Parsed entity info from DB:', entityInfo);
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
        } else {
          // Fallback to empty entity data if fetch fails
          const emptyEntityData: EntityProps = {
            entityId: '',
            name: '',
            email: '',
            phoneNumber: '',
            entityType: '',
            createdAt: '',
            updatedAt: '',
            roles: [],
            providerId: undefined,
            consumerId: undefined,
          };
          setEntityData(emptyEntityData);
        }
      } catch (error) {
        console.error('Failed to fetch entity info:', error);
        // Fallback to empty entity data if fetch fails
        const emptyEntityData: EntityProps = {
          entityId: '',
          name: '',
          email: '',
          phoneNumber: '',
          entityType: '',
          createdAt: '',
          updatedAt: '',
          roles: [],
          providerId: undefined,
          consumerId: undefined,
        };
        setEntityData(emptyEntityData);
      } finally {
        setLoading(false);
      }
    };

    fetchEntityInfo();
  }, [state.isAuthenticated]);

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

  const handleSignIn = () => {
    signIn();
  };

  const handleSignOut = () => {
    clearEntityStateFromStorage();
    setEntityData(null);
    setView(null);
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
  if (loading || !entityData) {
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
              <Route path="/" element={<MemberInfo 
                name={entityData.name}
                email={entityData.email}
                phoneNumber={entityData.phoneNumber}
                entityType={entityData.entityType}
                roles={entityData.roles}
                createdAt={entityData.createdAt}
                updatedAt={entityData.updatedAt}
              />} />
              <Route path="/provider/schemas" element={<SchemasPage providerId={entityData?.providerId || ''} />} />
              <Route 
                path="/provider/schemas/new" 
                element={
                  <SchemaRegistrationPage 
                    providerId={entityData?.providerId || ''}
                  />
                } 
              />
              <Route path="/provider/logs" element={<Logs role="provider" providerId={entityData?.providerId || ''} />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </>
          ) : (
            <>
              <Route path="/" element={<MemberInfo
                name={entityData.name}
                email={entityData.email}
                phoneNumber={entityData.phoneNumber}
                entityType={entityData.entityType}
                roles={entityData.roles}
                createdAt={entityData.createdAt}
                updatedAt={entityData.updatedAt}
              />} />
              <Route path="/consumer/applications" element={<Applications consumerId={entityData?.consumerId || ''} />} />
              <Route 
                path="/consumer/applications/new" 
                element={
                  <ApplicationRegistration 
                    consumerId={entityData?.consumerId || ''}
                  />
                } 
              />
              <Route path="/consumer/logs" element={<Logs role="consumer" consumerId={entityData?.consumerId || ''} />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </>
          )}
        </Routes>
      </div>
    </Router>
  );
}

export default App;