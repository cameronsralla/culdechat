import { createBrowserRouter, Navigate } from 'react-router';
import { RequireAdmin, RequireAnonymous, RequireUser } from '@/auth/guards';
import { AppShell } from '@/layout/AppShell';
import { LoginPage } from '@/features/auth/LoginPage';
import { RegisterPage } from '@/features/auth/RegisterPage';
import { HomePage } from '@/features/home/HomePage';
import { DirectoryPage } from '@/features/directory/DirectoryPage';
import { ChatPage } from '@/features/messages/MessagesPage';
import { YouPage } from '@/features/you/YouPage';
import { AdminPage } from '@/features/admin/AdminPage';
import { UnitsPage } from '@/features/admin/UnitsPage';

/**
 * Route tree. Feature pages register here; guards wrap by role.
 * Add a feature: create features/<name>/<Name>Page.tsx, add a route, add a nav item if top-level.
 */
export const router = createBrowserRouter([
  {
    element: <RequireAnonymous />,
    children: [
      { path: '/login', element: <LoginPage /> },
      { path: '/register', element: <RegisterPage /> },
    ],
  },
  {
    element: <RequireUser />,
    children: [
      {
        element: <AppShell />,
        children: [
          { path: '/', element: <HomePage /> },
          { path: '/directory', element: <DirectoryPage /> },
          { path: '/chat', element: <ChatPage /> },
          { path: '/messages', element: <Navigate to="/chat" replace /> },
          { path: '/you', element: <YouPage /> },
          {
            element: <RequireAdmin />,
            children: [
              { path: '/admin', element: <AdminPage /> },
              { path: '/units', element: <UnitsPage /> },
            ],
          },
        ],
      },
    ],
  },
  { path: '*', element: <Navigate to="/" replace /> },
]);
