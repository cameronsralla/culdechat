// API shapes. Keep in sync with server/internal/features/*.

export type User = {
  id: string;
  email: string;
  unit_number: string;
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

export type DirectoryEntry = Pick<User, 'id' | 'unit_number' | 'display_name' | 'email'>;

export type InvitePeek = { email: string; unit_number: string; display_name: string };

export type InviteResult = {
  user: User;
  passcode: string;
  expires_at: string;
  invite_url?: string;
};

export type PublicSettings = { community_name?: string; capability_tier?: 'CORE' | 'STANDARD' | 'PLUS' };

export type ApiErrorBody = { error: { code: string; message: string } };
