import { createContext, useContext, useState, useEffect, useCallback, ReactNode, useMemo } from 'react';
import { jwtDecode } from 'jwt-decode';

interface User {
  id: string;
  email: string;
  name: string;
  role: string;
}

interface JWTPayload {
  sub: string;
  exp: number;
  iat: number;
  role?: string;
  email?: string;
  name?: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (token: string, user: User) => void;
  logout: () => void;
  checkAuth: () => boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Storage keys
const TOKEN_KEY = 'auth_token';
const USER_KEY = 'user';

// Token refresh buffer (refresh 5 minutes before expiry)
const REFRESH_BUFFER_MS = 5 * 60 * 1000;

// Check if token is expired
function isTokenExpired(token: string): boolean {
  try {
    const decoded = jwtDecode<JWTPayload>(token);
    const expiresAt = decoded.exp * 1000; // Convert to milliseconds
    return Date.now() >= expiresAt;
  } catch {
    return true;
  }
}

// Get time until token expires
function getTimeUntilExpiry(token: string): number {
  try {
    const decoded = jwtDecode<JWTPayload>(token);
    const expiresAt = decoded.exp * 1000;
    return expiresAt - Date.now();
  } catch {
    return 0;
  }
}

// Broadcast channel for cross-tab communication
const authChannel = typeof BroadcastChannel !== 'undefined' 
  ? new BroadcastChannel('auth_channel') 
  : null;

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Initialize auth state from storage
  useEffect(() => {
    const initAuth = () => {
      const storedToken = localStorage.getItem(TOKEN_KEY);
      const storedUser = localStorage.getItem(USER_KEY);

      if (storedToken && storedUser) {
        // Check if token is still valid
        if (!isTokenExpired(storedToken)) {
          setToken(storedToken);
          setUser(JSON.parse(storedUser));
        } else {
          // Clear expired token
          localStorage.removeItem(TOKEN_KEY);
          localStorage.removeItem(USER_KEY);
        }
      }
      setIsLoading(false);
    };

    initAuth();
  }, []);

  // Set up token expiration check
  useEffect(() => {
    if (!token) return;

    const timeUntilExpiry = getTimeUntilExpiry(token);
    
    // If token is about to expire, log out
    if (timeUntilExpiry <= 0) {
      logout();
      return;
    }

    // Set up timer to check before expiry
    const checkTime = Math.max(timeUntilExpiry - REFRESH_BUFFER_MS, 1000);
    const timer = setTimeout(() => {
      // Token is about to expire
      if (token && isTokenExpired(token)) {
        logout();
      }
    }, checkTime);

    return () => clearTimeout(timer);
  }, [token]);

  // Listen for auth changes in other tabs
  useEffect(() => {
    if (!authChannel) return;

    const handleMessage = (event: MessageEvent) => {
      const { type, payload } = event.data;
      
      switch (type) {
        case 'LOGIN':
          setToken(payload.token);
          setUser(payload.user);
          break;
        case 'LOGOUT':
          setToken(null);
          setUser(null);
          break;
      }
    };

    authChannel.addEventListener('message', handleMessage);
    return () => authChannel.removeEventListener('message', handleMessage);
  }, []);

  // Listen for storage changes (fallback for browsers without BroadcastChannel)
  useEffect(() => {
    const handleStorageChange = (event: StorageEvent) => {
      if (event.key === TOKEN_KEY) {
        if (event.newValue) {
          const storedUser = localStorage.getItem(USER_KEY);
          if (storedUser) {
            setToken(event.newValue);
            setUser(JSON.parse(storedUser));
          }
        } else {
          setToken(null);
          setUser(null);
        }
      }
    };

    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, []);

  const login = useCallback((newToken: string, newUser: User) => {
    setToken(newToken);
    setUser(newUser);
    localStorage.setItem(TOKEN_KEY, newToken);
    localStorage.setItem(USER_KEY, JSON.stringify(newUser));

    // Broadcast to other tabs
    authChannel?.postMessage({ type: 'LOGIN', payload: { token: newToken, user: newUser } });
  }, []);

  const logout = useCallback(() => {
    setToken(null);
    setUser(null);
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);

    // Broadcast to other tabs
    authChannel?.postMessage({ type: 'LOGOUT' });
  }, []);

  const checkAuth = useCallback((): boolean => {
    if (!token) return false;
    
    if (isTokenExpired(token)) {
      logout();
      return false;
    }
    
    return true;
  }, [token, logout]);

  const value = useMemo(() => ({
    user,
    token,
    isAuthenticated: !!token && !isTokenExpired(token),
    isLoading,
    login,
    logout,
    checkAuth,
  }), [user, token, isLoading, login, logout, checkAuth]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
