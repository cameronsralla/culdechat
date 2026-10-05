// API shapes. Keep in sync with server/internal/features/*.

export type User = {
  id: string;
  email: string;
  unit_number: string;
  unit_id: string;
  is_primary: boolean;
  display_name: string;
  is_admin: boolean;
  status: 'invited' | 'active' | 'inactive';
  directory_opt_in: boolean;
  created_at: string;
};

export type Session = {
  access_token: string;
  expires_in: number;
  refresh_token: string;
  user: User;
};

export type DirectoryEntry = {
  kind: 'person' | 'unit';
  id?: string;
  unit_number: string;
  display_name?: string;
  email?: string;
  self?: boolean;
};

export type MessagePeer = {
  id?: string;
  unit_number: string;
  display_name?: string;
  listed: boolean;
};

export type DirectMessage = {
  id: string;
  mine: boolean;
  body: string;
  created_at: string;
};

export type Conversation = {
  id: string;
  status: 'pending' | 'open' | 'declined';
  incoming: boolean;
  requested_by_me: boolean;
  peer: MessagePeer;
  last_message?: string;
  updated_at: string;
  messages?: DirectMessage[];
};

export type UnitResident = {
  id: string;
  email: string;
  display_name: string;
  status: User['status'];
  is_primary: boolean;
};

export type Unit = {
  id: string;
  number: string;
  created_at: string;
  residents: UnitResident[];
};

export type UnitPage = {
  items: Unit[];
  total: number;
  page: number;
  page_size: number;
};

export type InvitePeek = { email: string; unit_number: string; display_name: string };

export type InviteResult = {
  user: User;
  passcode: string;
  expires_at: string;
  invite_url?: string;
};

export type PublicSettings = { community_name?: string; capability_tier?: 'CORE' | 'STANDARD' | 'PLUS' };

export type ApiErrorBody = { error: { code: string; message: string } };
