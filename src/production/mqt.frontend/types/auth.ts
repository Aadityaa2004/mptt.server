export type UserRole = "user" | "admin";

export interface LoginRequest {
  username: string;
  password: string;
  turnstile_token?: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  role?: UserRole;
  turnstile_token?: string;
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

export interface RegisterResponse {
  id?: string; // Only present after verification; not set during pending signup
  username: string;
  email: string;
  role: UserRole;
  requires_verification: boolean;
  message: string;
}

