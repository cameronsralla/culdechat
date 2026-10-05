import { Navigate, Outlet, useLocation } from 'react-router';
import { useAuth } from './AuthContext';
import { LoadingScreen } from '@/layout/LoadingScreen';

/** Routes that need a signed-in user. Redirects to /login, remembering where you were. */
export function RequireUser() {
  const { status } = useAuth();
  const loc = useLocation();
  if (status === 'loading') return <LoadingScreen />;
  if (status === 'signed-out') return <Navigate to="/login" replace state={{ from: loc.pathname }} />;
  return <Outlet />;
}

/** Routes that need an admin. Non-admins are sent home. */
export function RequireAdmin() {
  const { user } = useAuth();
  if (!user?.is_admin) return <Navigate to="/" replace />;
  return <Outlet />;
}

/** Login/register: bounce signed-in users to the app. */
export function RequireAnonymous() {
  const { status } = useAuth();
  if (status === 'loading') return <LoadingScreen />;
  if (status === 'signed-in') return <Navigate to="/" replace />;
  return <Outlet />;
}
