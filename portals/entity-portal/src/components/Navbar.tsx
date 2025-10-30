// components/Navbar.tsx
import React, {useEffect, useState} from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthContext } from "@asgardeo/auth-react";

interface NavItem {
  label: string;
  path: string;
  icon?: React.ReactNode;
}

const MemberNavItems: NavItem[] = [
  {
    label: 'Schemas',
    path: '/schemas',
    icon: (
      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
    ),
  },
  {
    label: 'Schemas Logs',
    path: '/schemas/logs',
    icon: (
      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1M4 8V7a3 3 0 013-3h10a3 3 0 013 3v1m-16 0a3 3 0 003 3h10a3 3 0 003-3m-16 8a3 3 0 003 3h10a3 3 0 003-3m-16-8a3 3 0 003-3h10a3 3 0 003 3" />
      </svg>
    ),
  },
    {
        label: 'Applications',
        path: '/applications',
        icon: (
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
        ),
    },
    {
        label: 'Applications Logs',
        path: '/applications/logs',
        icon: (
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1M4 8V7a3 3 0 013-3h10a3 3 0 013 3v1m-16 0a3 3 0 003 3h10a3 3 0 003-3m-16 8a3 3 0 003 3h10a3 3 0 003-3m-16-8a3 3 0 003-3h10a3 3 0 003 3" />
            </svg>
        ),
    },
];

interface SideNavbarProps {
    onSignOut: () => void;
}

export const SideNavbar: React.FC<SideNavbarProps> = (
  {
    onSignOut
  }) => {
    const navigate = useNavigate();
    const location = useLocation();
    const [userDropdownOpen, setUserDropdownOpen] = useState(false);
    const [sidebarExpanded, setSidebarExpanded] = useState(true);
    const { state, getBasicUserInfo } = useAuthContext();
    const [ userName, setUserName ] = useState<string | null>(null);

    const isActive = (path: string) => {
        if (path === '/') {
            return location.pathname === '/';
        }
        return location.pathname === path;
    };

    const handleSignOut = () => {
        setUserDropdownOpen(false);
        onSignOut();
    };

    const handleNavItemClick = (path: string) => {
        navigate(path);
    };

    useEffect(() => {
        const fetchUserName = async () => {
            const userInfo = await getBasicUserInfo();
            console.log('User Info:', userInfo);
            if (userInfo.displayName) {
                setUserName(userInfo.displayName);
            }
        };

        fetchUserName();
    }, [state.isAuthenticated]);

    return (
        <>
            {/* Left Sidebar */}
            <div className={`bg-white shadow-xl border-r border-gray-200 h-screen transition-all duration-300 ease-in-out ${
                sidebarExpanded ? 'w-64' : 'w-16'
            }`}>
                {/* Navigation Items */}
                <nav className="pt-20 p-4 space-y-2">
                    {MemberNavItems.map((item) => (
                        <button
                            key={item.path}
                            onClick={() => handleNavItemClick(item.path)}
                            className={`w-full flex items-center ${sidebarExpanded ? 'space-x-3 px-4' : 'justify-center px-2'} py-3 rounded-lg text-sm font-medium transition-all duration-200 ${
                                isActive(item.path)
                                    ? 'bg-blue-100 text-blue-800 border border-blue-300 shadow-sm'
                                    : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
                            }`}
                            title={!sidebarExpanded ? item.label : undefined}
                        >
                            <div className="flex-shrink-0">
                                {item.icon}
                            </div>
                            {sidebarExpanded && <span>{item.label}</span>}
                        </button>
                    ))}
                </nav>
            </div>

            {/* Top Navigation Bar - Fixed at top, full width */}
            <nav className="bg-white shadow-md border-b border-gray-300 fixed top-0 left-0 right-0 z-50">
                <div className="flex justify-between items-center h-16">
                    {/* Left side - Menu button and App title - Fixed position */}
                    <div className="flex items-center space-x-4 pl-6">
                        <button
                            onClick={() => setSidebarExpanded(!sidebarExpanded)}
                            className="p-2 rounded-lg text-gray-600 hover:text-gray-900 hover:bg-gray-100 transition-all duration-200"
                        >
                            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                            </svg>
                        </button>
                        <h1 className="text-xl font-semibold text-gray-900">Member Portal</h1>
                    </div>

                    {/* Right side - User actions */}
                    <div className="flex items-center space-x-2 pr-6">
                        {state.isAuthenticated && (
                            <div className="relative">
                                <button
                                    onClick={() => setUserDropdownOpen(!userDropdownOpen)}
                                    className="flex items-center space-x-2 px-3 py-2 rounded-lg text-gray-700 hover:text-gray-900 hover:bg-gray-100 transition-all duration-200"
                                >
                                    <div className="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center">
                                        <svg className="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                                        </svg>
                                    </div>
                                    <span className="hidden sm:inline text-sm font-medium">
                    {userName ? userName : 'User'}
                  </span>
                                    <svg className="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                                    </svg>
                                </button>

                                {userDropdownOpen && (
                                    <div className="absolute right-0 top-full mt-2 w-48 bg-white rounded-lg shadow-xl border border-gray-200 z-10 overflow-hidden">
                                        <div className="px-4 py-3 border-b border-gray-100">
                                            <p className="text-sm font-medium text-gray-900">{userName ? userName : 'User'}</p>
                                        </div>
                                        <div className="border-t border-gray-100">
                                            <button
                                                onClick={handleSignOut}
                                                className="w-full text-left px-4 py-3 text-sm text-red-600 hover:bg-red-50 hover:text-red-800 transition-colors flex items-center space-x-2"
                                            >
                                                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                                                </svg>
                                                <span>Sign Out</span>
                                            </button>
                                        </div>
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                </div>
            </nav>
        </>
    );
};