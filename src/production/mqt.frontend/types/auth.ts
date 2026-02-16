export type UserRole = "user" | "admin";

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  role?: UserRole;
}

export interface AuthResponse {
  access_token: string;
  token_id: string;
  expires_at?: number;
  user_id: string;
  username: string;
  email: string;
  role: UserRole;
}

export interface RefreshTokenResponse {
  access_token: string;
  token_id: string;
  expires_at: number;
}

export interface User {
  user_id: string;
  username: string;
  email: string;
  role: UserRole;
}

