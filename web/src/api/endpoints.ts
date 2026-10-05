import { api } from './client';
import type { Conversation, DirectoryEntry, InvitePeek, InviteResult, PublicSettings, Session, Unit, UnitPage, User } from './types';

export type UserListQuery = {
  q?: string;
  sort?: string;
  dir?: 'asc' | 'desc';
  page?: number;
  pageSize?: number;
  name?: string;
  email?: string;
  unit?: string;
  status?: string;
  admin?: string;
  directory?: string;
};

export type UserPage = {
  items: User[];
  total: number;
  page: number;
  page_size: number;
};

export type DirectoryPage = {
  items: DirectoryEntry[];
  total: number;
  page: number;
  page_size: number;
};

export type DirectoryQuery = {
  q?: string;
  sort?: string;
  dir?: 'asc' | 'desc';
  page?: number;
  pageSize?: number;
  name?: string;
  email?: string;
  unit?: string;
};

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
  directory: (query: DirectoryQuery = {}) => {
    const params = new URLSearchParams();
    const set = (key: string, value: string | number | undefined) => {
      if (value !== undefined && value !== '') params.set(key, String(value));
    };
    set('q', query.q);
    set('sort', query.sort);
    set('dir', query.dir);
    set('page', query.page);
    set('page_size', query.pageSize);
    set('name', query.name);
    set('email', query.email);
    set('unit', query.unit);
    const qs = params.toString();
    return api.get<DirectoryPage>(`/directory${qs ? `?${qs}` : ''}`);
  },
};

export const adminApi = {
  users: (query: UserListQuery = {}) => {
    const params = new URLSearchParams();
    const set = (key: string, value: string | number | undefined) => {
      if (value !== undefined && value !== '') params.set(key, String(value));
    };
    set('q', query.q);
    set('sort', query.sort);
    set('dir', query.dir);
    set('page', query.page);
    set('page_size', query.pageSize);
    set('name', query.name);
    set('email', query.email);
    set('unit', query.unit);
    set('status', query.status);
    set('admin', query.admin);
    set('directory', query.directory);
    const qs = params.toString();
    return api.get<UserPage>(`/admin/users${qs ? `?${qs}` : ''}`);
  },
  units: (query: { q?: string; dir?: 'asc' | 'desc'; page?: number; pageSize?: number; vacant?: boolean } = {}) => {
    const params = new URLSearchParams();
    const set = (key: string, value: string | number | boolean | undefined) => {
      if (value !== undefined && value !== '') params.set(key, String(value));
    };
    set('q', query.q);
    set('dir', query.dir);
    set('page', query.page);
    set('page_size', query.pageSize);
    if (query.vacant) set('vacant', true);
    const qs = params.toString();
    return api.get<UnitPage>(`/admin/units${qs ? `?${qs}` : ''}`);
  },
  createUnit: (number: string) => api.post<Unit>('/admin/units', { number }),
  deleteUnit: (id: string) => api.delete<void>(`/admin/units/${id}`),
  invite: (input: { email: string; unit_id: string; display_name?: string; is_admin?: boolean }) =>
    api.post<InviteResult>('/admin/users/invite', input),
  reinvite: (id: string) => api.post<InviteResult>(`/admin/users/${id}/reinvite`),
  resetPassword: (id: string) =>
    api.post<{ user: User; temporary_password: string }>(`/admin/users/${id}/reset-password`),
  setAdmin: (id: string, is_admin: boolean) => api.put<User>(`/admin/users/${id}/admin`, { is_admin }),
  setStatus: (id: string, active: boolean) => api.put<User>(`/admin/users/${id}/status`, { active }),
};

export const messagesApi = {
  list: () => api.get<Conversation[]>('/messages/conversations'),
  get: (id: string) => api.get<Conversation>(`/messages/conversations/${id}`),
  send: (input: { content: string; user_id?: string; unit_number?: string; conversation_id?: string }) =>
    api.post<Conversation>('/messages', input),
  accept: (id: string) => api.post<Conversation>(`/messages/conversations/${id}/accept`),
  decline: (id: string) => api.post<Conversation>(`/messages/conversations/${id}/decline`),
};

export const settingsApi = {
  public: () => api.get<PublicSettings>('/settings'),
};
