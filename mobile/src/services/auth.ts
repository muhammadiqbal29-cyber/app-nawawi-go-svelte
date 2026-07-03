/**
 * auth.ts — Modul autentikasi untuk mobile app.
 * Mengelola login, logout, token, dan JWT claim parsing.
 */

import { post } from './api';
import { navigate } from '../store/nav.svelte';

// Tipe response dari backend login
interface LoginResponse {
    token: string;
    peran: string;
}

// Tipe JWT claim (decoded dari token, tanpa verifikasi signature)
export interface JWTClaim {
    id: number;
    email: string;
    peran: string;
    exp: number;
}

/**
 * Login ke backend via `/api/v1/auth/mobile/login`
 * Menyimpan token & peran ke localStorage
 */
export async function login(email: string, password: string): Promise<LoginResponse> {
    const data = await post<LoginResponse>('/api/v1/auth/mobile/login', { email, password });

    if (data.token) {
        localStorage.setItem('token', data.token);
        localStorage.setItem('role', data.peran);
    }

    return data;
}

/**
 * Logout: hapus token & peran, navigasi ke halaman login
 */
export function logout(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    navigate('login');
}

/**
 * Ambil token dari localStorage
 */
export function getToken(): string | null {
    return localStorage.getItem('token');
}

/**
 * Ambil peran (role) dari localStorage
 */
export function getRole(): string | null {
    return localStorage.getItem('role');
}

/**
 * Cek apakah user sudah login dan token belum expired
 */
export function isAuthenticated(): boolean {
    const token = getToken();
    if (!token) return false;

    try {
        const claims = decodeToken(token);
        if (!claims) return false;

        // Cek apakah token sudah expired
        const now = Math.floor(Date.now() / 1000);
        return claims.exp > now;
    } catch {
        return false;
    }
}

/**
 * Decode JWT token payload (tanpa verifikasi signature).
 * Hanya digunakan untuk membaca claims (id, email, peran) di client side.
 */
export function decodeToken(token: string): JWTClaim | null {
    try {
        const parts = token.split('.');
        if (parts.length !== 3) return null;

        const payload = parts[1];
        // Base64url decode
        const base64 = payload.replace(/-/g, '+').replace(/_/g, '/');
        const jsonStr = decodeURIComponent(
            atob(base64)
                .split('')
                .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
                .join('')
        );

        return JSON.parse(jsonStr) as JWTClaim;
    } catch {
        return null;
    }
}

/**
 * Ambil data user saat ini dari JWT token
 */
export function getCurrentUser(): JWTClaim | null {
    const token = getToken();
    if (!token) return null;
    return decodeToken(token);
}
