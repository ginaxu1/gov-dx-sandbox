import React, { useEffect, useState } from 'react';

export type Route = {
  path: string;
  component: React.ComponentType;
  exact?: boolean;
};

// Simple client-side router implementation
export const Router: React.FC<{ routes: Route[] }> = ({ routes }) => {
  const [currentPath, setCurrentPath] = useState(window.location.pathname);

  useEffect(() => {
    const handlePopState = () => {
      setCurrentPath(window.location.pathname);
    };

    // Listen for custom navigation events from Navbar
    const handleNavigationEvent = () => {
      setCurrentPath(window.location.pathname);
    };

    window.addEventListener('popstate', handlePopState);
    window.addEventListener('navigate', handleNavigationEvent);
    
    return () => {
      window.removeEventListener('popstate', handlePopState);
      window.removeEventListener('navigate', handleNavigationEvent);
    };
  }, []);

  const navigate = (path: string) => {
    if (path === currentPath) return;
    
    window.history.pushState({}, '', path);
    setCurrentPath(path);
    
    // Dispatch custom event for other components to listen to
    window.dispatchEvent(new Event('navigate'));
  };

  // Provide navigation context to child components
  React.useEffect(() => {
    (window as any).navigate = navigate;
  }, []);

  // Improved route matching - check exact matches first, then partial matches
  const findMatchingRoute = () => {
    // First, try to find exact matches
    const exactMatch = routes.find(route => 
      route.exact && route.path === currentPath
    );
    
    if (exactMatch) return exactMatch;
    
    // Then try partial matches, prioritizing longer paths first
    const partialMatches = routes
      .filter(route => !route.exact && currentPath.startsWith(route.path))
      .sort((a, b) => b.path.length - a.path.length);
    
    return partialMatches[0];
  };

  const matchingRoute = findMatchingRoute();

  if (matchingRoute) {
    const Component = matchingRoute.component;
    return <Component />;
  }

  // 404 fallback
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">404</h1>
        <p className="text-gray-600 mb-8">Page not found</p>
        <button
          onClick={() => navigate('/')}
          className="bg-blue-500 text-white px-4 py-2 rounded-md hover:bg-blue-600 transition-colors"
        >
          Go Home
        </button>
      </div>
    </div>
  );
};
