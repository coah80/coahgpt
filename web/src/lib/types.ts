export interface Message {
  readonly id: string;
  readonly role: 'user' | 'assistant';
  readonly content: string;
  readonly timestamp: number;
}

export interface Conversation {
  readonly id: string;
  readonly messages: ReadonlyArray<Message>;
  readonly preview: string;
  readonly createdAt: number;
  readonly updatedAt: number;
}

export interface StreamToken {
  readonly token: string;
  readonly done: boolean;
  readonly session_id?: string;
}
