export type Author = {
  id: string;
  unit_number: string;
  name?: string | null;
};

export type BoardRef = {
  id: string;
  name: string;
};

export type FeedPost = {
  id: string;
  title: string;
  snippet: string;
  author: Author;
  board: BoardRef;
  comment_count: number;
  reaction_count: number;
  post_type: string;
  is_pinned: boolean;
  my_reaction?: string;
  created_at: string;
};

export type FeedPage = {
  posts: FeedPost[];
  next_page_cursor?: string | null;
};

export type Board = {
  id: string;
  name: string;
  description?: string | null;
  creator_id: string;
  subscriber_count: number;
  is_subscribed: boolean;
};

export type ReactionCount = {
  type: string;
  count: number;
};

export type Comment = {
  id: string;
  content: string;
  author: Author;
  created_at?: string;
};

export type PostDetail = {
  id: string;
  title: string;
  content: string;
  author: Author;
  board: BoardRef;
  post_type: string;
  is_pinned: boolean;
  comments_disabled: boolean;
  comments: Comment[];
  reactions: ReactionCount[];
  my_reaction?: string;
  created_at: string;
};

export type CreatedPost = {
  id: string;
  title: string;
  content: string;
  author_id: string;
  board_id: string;
  post_type: string;
  is_pinned: boolean;
};

export type Profile = {
  id: string;
  email: string;
  name: string;
  unit_number: string;
  profile_picture_url?: string | null;
  directory_opt_in: boolean;
  is_admin: boolean;
};

export type DirectoryUser = {
  id: string;
  name: string;
  unit_number: string;
  profile_picture_url?: string | null;
};

export type AdminUser = {
  id: string;
  email: string;
  name: string;
  unit_number: string;
  status: string;
  is_admin: boolean;
};

export type MessagePeer = {
  id: string;
  unit_number: string;
  name?: string | null;
  profile_picture_url?: string | null;
  directory_opt_in: boolean;
};

export type DirectMessage = {
  id: string;
  conversation_id: string;
  sender_id: string;
  content: string;
  created_at: string;
};

export type ConversationSummary = {
  id: string;
  peer: MessagePeer;
  last_message?: DirectMessage | null;
  updated_at: string;
};

export type ConversationDetail = {
  id: string;
  peer: MessagePeer;
  messages: DirectMessage[];
};

export type MessageRecipient = {
  kind: 'user' | 'unit';
  id?: string | null;
  name?: string | null;
  unit_number: string;
  profile_picture_url?: string | null;
};

export type SendMessageResult = {
  conversation_id: string;
  message: DirectMessage;
  created: boolean;
};

export const reactionTypes = ['like', 'love', 'laugh', 'wow', 'sad', 'angry'] as const;
export type ReactionType = (typeof reactionTypes)[number];

export const reactionEmoji: Record<ReactionType, string> = {
  like: '👍',
  love: '❤️',
  laugh: '😄',
  wow: '😮',
  sad: '😢',
  angry: '😡',
};

export function authorLabel(author: Author): string {
  const name = author.name?.trim();
  if (name) {
    return name;
  }
  return `Unit ${author.unit_number}`;
}

export function peerLabel(peer: Pick<MessagePeer, 'name' | 'unit_number'>): string {
  const name = peer.name?.trim();
  if (name) {
    return name;
  }
  return `Unit ${peer.unit_number}`;
}

export function postBadge(post: { post_type: string; is_pinned: boolean }): 'bulletin' | 'pinned' | null {
  if (post.post_type === 'bulletin') {
    return 'bulletin';
  }
  if (post.is_pinned) {
    return 'pinned';
  }
  return null;
}
