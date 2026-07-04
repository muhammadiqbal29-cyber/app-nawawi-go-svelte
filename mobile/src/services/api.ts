/**
 * api.ts — Modul HTTP utility terpusat untuk komunikasi dengan Backend Golang.
 * Semua request API di mobile app harus melewati fungsi-fungsi ini.
 */

const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export class ApiError extends Error {
    status: number;
    constructor(message: string, status: number) {
        super(message);
        this.status = status;
        this.name = 'ApiError';
    }
}

function getAuthHeaders(): Record<string, string> {
    const token = localStorage.getItem('token');
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        'X-Requested-With': 'XMLHttpRequest',
    };
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    return headers;
}

async function handleResponse<T>(response: Response): Promise<T> {
    if (response.status === 401) {
        // Token expired atau invalid — hapus dan redirect ke login
        localStorage.removeItem('token');
        localStorage.removeItem('role');
        // Import navigate dynamically to avoid circular dependency
        const { navigate } = await import('../store/nav.svelte');
        navigate('login');
        throw new ApiError('Sesi telah berakhir, silakan login kembali.', 401);
    }

    if (response.status === 204) {
        return undefined as unknown as T;
    }

    const text = await response.text();

    if (!response.ok) {
        throw new ApiError(text || `Error ${response.status}`, response.status);
    }

    // Coba parse sebagai JSON
    try {
        return JSON.parse(text) as T;
    } catch {
        // Jika bukan JSON, kembalikan text sebagai any
        return text as unknown as T;
    }
}

/**
 * GET request
 */
export async function get<T = any>(path: string, params?: Record<string, string | number>): Promise<T> {
    let url = `${BASE_URL}${path}`;

    if (params) {
        const searchParams = new URLSearchParams();
        for (const [key, value] of Object.entries(params)) {
            if (value !== undefined && value !== null && value !== 0 && value !== '') {
                searchParams.append(key, String(value));
            }
        }
        const qs = searchParams.toString();
        if (qs) url += `?${qs}`;
    }

    const response = await fetch(url, {
        method: 'GET',
        headers: getAuthHeaders(),
    });

    return handleResponse<T>(response);
}

/**
 * POST request
 */
export async function post<T = any>(path: string, body?: any): Promise<T> {
    const response = await fetch(`${BASE_URL}${path}`, {
        method: 'POST',
        headers: getAuthHeaders(),
        body: body !== undefined ? JSON.stringify(body) : undefined,
    });

    return handleResponse<T>(response);
}

/**
 * PUT request
 */
export async function put<T = any>(path: string, body?: any): Promise<T> {
    const response = await fetch(`${BASE_URL}${path}`, {
        method: 'PUT',
        headers: getAuthHeaders(),
        body: body !== undefined ? JSON.stringify(body) : undefined,
    });

    return handleResponse<T>(response);
}

/**
 * DELETE request
 */
export async function del<T = any>(path: string): Promise<T> {
    const response = await fetch(`${BASE_URL}${path}`, {
        method: 'DELETE',
        headers: getAuthHeaders(),
    });

    return handleResponse<T>(response);
}
