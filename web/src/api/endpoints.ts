import { api } from './client';
import type { DirectoryEntry, InvitePeek, InviteResult, PublicSettings, Session, User } from './types';

export const authApi = {
  login: (email: string, password: string) => api.post<Session>('/auth/login', { email, password }, { auth: false }),
  refresh: (refresh_token: string) => api.post<Session>('/auth/refresh', { refresh_token }, { auth: false }),
  logout: (refresh_token: string) => api.post<void>('/auth/logout', { refresh_token }, { auth: false }),
  logoutAll: () => api.post<void>('/auth/logout-all'),
  peekInvite: (token: string) => api.get<InvitePeek>(`/auth/invite?token=${encodeURIComponent(token)}`, { auth: false }),
  completeInvite: (input: { token: string; passcode: string; password: string; display_name: string }) =>
    api.post<Session>('/auth/complete-invite', input, { auth: false }),
  changePassword: (current_password: string, new_password: string) =>
    api.post<Session>('/auth/change-password', { current_password, new_password }),
};

export const usersApi = {
  me: () => api.get<User>('/me'),
  updateMe: (input: { display_name: string; directory_opt_in: boolean }) => api.patch<User>('/me', input),
  directory: () => api.get<DirectoryEntry[]>('/directory'),
};

export const adminApi = {
  users: () => api.get<User[]>('/admin/users'),
  invite: (input: { email: string; unit_number: string; display_name?: string; is_admin?: boolean }) =>
    api.post<InviteResult>('/admin/users/invite', input),
  reinvite: (id: string) => api.post<InviteResult>(`/admin/users/${id}/reinvite`),
  setAdmin: (id: string, is_admin: boolean) => api.put<User>(`/admin/users/${id}/admin`, { is_admin }),
  setStatus: (id: string, active: boolean) => api.put<User>(`/admin/users/${id}/status`, { active }),
};

export const settingsApi = {
  public: () => api.get<PublicSettings>('/settings'),
};
