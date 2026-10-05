import { request } from './client';
import type {
  AdminUser,
  Board,
  ConversationDetail,
  ConversationSummary,
  CreatedPost,
  DirectoryUser,
  FeedPage,
  MessageRecipient,
  PostDetail,
  Profile,
  SendMessageResult,
} from './types';

export async function listBoards(): Promise<Board[]> {
  return (await request<Board[] | null>('/boards', { method: 'GET' })) ?? [];
}

export function getBoard(boardId: string): Promise<Board> {
  return request<Board>(`/boards/${boardId}`, { method: 'GET' });
}

export function createBoard(name: string, description: string): Promise<Board> {
  return request<Board>('/boards', {
    method: 'POST',
    body: { name, description: description || undefined },
  });
}

export function toggleSubscribe(boardId: string): Promise<{ subscribed: boolean; message: string }> {
  return request(`/boards/${boardId}/subscribe`, { method: 'POST' });
}

export function listFeed(opts?: { boardId?: string; cursor?: string | null; limit?: number }): Promise<FeedPage> {
  const q = new URLSearchParams();
  if (opts?.limit) {
    q.set('limit', String(opts.limit));
  }
  if (opts?.cursor) {
    q.set('cursor', opts.cursor);
  }
  const qs = q.toString();
  const path = opts?.boardId ? `/boards/${opts.boardId}/posts` : '/posts';
  return request<FeedPage>(qs ? `${path}?${qs}` : path, { method: 'GET' });
}

export function getPost(postId: string): Promise<PostDetail> {
  return request<PostDetail>(`/posts/${postId}`, { method: 'GET' });
}

export function createPost(
  boardId: string,
  input: { title: string; content: string; post_type?: string; is_pinned?: boolean },
): Promise<CreatedPost> {
  return request<CreatedPost>(`/boards/${boardId}/posts`, { method: 'POST', body: input });
}

export function updatePost(
  postId: string,
  input: { title?: string; content?: string; is_pinned?: boolean },
): Promise<CreatedPost> {
  return request<CreatedPost>(`/posts/${postId}`, { method: 'PATCH', body: input });
}

export function deletePost(postId: string): Promise<void> {
  return request(`/posts/${postId}`, { method: 'DELETE' });
}

export function addComment(postId: string, content: string): Promise<{ id: string }> {
  return request(`/posts/${postId}/comments`, { method: 'POST', body: { content } });
}

export function updateComment(postId: string, commentId: string, content: string): Promise<{ id: string }> {
  return request(`/posts/${postId}/comments/${commentId}`, { method: 'PATCH', body: { content } });
}

export function deleteComment(postId: string, commentId: string): Promise<void> {
  return request(`/posts/${postId}/comments/${commentId}`, { method: 'DELETE' });
}

export function setReaction(postId: string, type: string): Promise<void> {
  return request(`/posts/${postId}/reactions`, { method: 'PUT', body: { type } });
}

export function clearReaction(postId: string): Promise<void> {
  return request(`/posts/${postId}/reactions`, { method: 'DELETE' });
}

export function getProfile(): Promise<Profile> {
  return request<Profile>('/profile/me', { method: 'GET' });
}

export function updateProfile(input: { name?: string; directory_opt_in?: boolean }): Promise<Profile> {
  return request<Profile>('/profile/me', { method: 'PATCH', body: input });
}

export function uploadProfilePhoto(file: Blob): Promise<Profile> {
  const form = new FormData();
  form.append('photo', file, photoFilename(file));
  return request<Profile>('/profile/me/photo', { method: 'POST', form });
}

function photoFilename(file: Blob): string {
  if (file instanceof File && file.name) {
    return file.name;
  }
  const type = file.type;
  if (type === 'image/png') {
    return 'photo.png';
  }
  if (type === 'image/webp') {
    return 'photo.webp';
  }
  return 'photo.jpg';
}

export function changePassword(current_password: string, new_password: string): Promise<void> {
  return request('/auth/change-password', { method: 'POST', body: { current_password, new_password } });
}

export async function listDirectory(): Promise<DirectoryUser[]> {
  return (await request<DirectoryUser[] | null>('/directory', { method: 'GET' })) ?? [];
}

export function inviteResident(email: string, unit_number: string): Promise<{
  email: string;
  registration_token: string;
  passcode: string;
  email_sent: boolean;
  message: string;
}> {
  return request('/auth/register', { method: 'POST', body: { email, unit_number } });
}

export async function listAdminUsers(): Promise<AdminUser[]> {
  return (await request<AdminUser[] | null>('/admin/users', { method: 'GET' })) ?? [];
}

export function offboardUser(userId: string): Promise<void> {
  return request(`/admin/users/${userId}/offboard`, { method: 'POST' });
}

export async function listConversations(): Promise<ConversationSummary[]> {
  return (await request<ConversationSummary[] | null>('/messages/conversations', { method: 'GET' })) ?? [];
}

export function getConversation(
  conversationId: string,
  opts?: { limit?: number; before?: string },
): Promise<ConversationDetail> {
  const q = new URLSearchParams();
  if (opts?.limit) {
    q.set('limit', String(opts.limit));
  }
  if (opts?.before) {
    q.set('before', opts.before);
  }
  const qs = q.toString();
  return request<ConversationDetail>(
    qs ? `/messages/conversations/${conversationId}?${qs}` : `/messages/conversations/${conversationId}`,
    { method: 'GET' },
  );
}

export function sendDirectMessage(input: {
  content: string;
  user_id?: string;
  unit_number?: string;
}): Promise<SendMessageResult> {
  return request<SendMessageResult>('/messages', { method: 'POST', body: input });
}

export async function searchRecipients(q: string): Promise<MessageRecipient[]> {
  const qs = new URLSearchParams({ q });
  return (await request<MessageRecipient[] | null>(`/messages/recipients?${qs}`, { method: 'GET' })) ?? [];
}

