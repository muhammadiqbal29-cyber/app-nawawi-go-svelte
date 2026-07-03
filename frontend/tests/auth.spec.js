import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  test('harus berhasil login dengan admin', async ({ page }) => {
    // Navigasi ke halaman login
    await page.goto('/login');

    // Mengisi form login
    await page.fill('input[type="email"]', 'admin@siakad.com');
    await page.fill('input[type="password"]', 'password');

    // Menekan tombol login (biasanya tombol submit di form)
    await page.click('button[type="submit"]');

    // Menunggu navigasi selesai, biasanya diarahkan ke dashboard
    await page.waitForURL('/');

    // Verifikasi bahwa user berada di halaman dashboard (mengecek elemen khusus dashboard)
    // const dashboardHeading = page.getByRole('heading', { name: /dashboard|beranda/i });
    // Kita cek URL saja jika tidak ada heading spesifik
    expect(page.url()).toBe('http://localhost:5173/');
  });

  test('harus gagal login dengan password salah', async ({ page }) => {
    await page.goto('/login');

    await page.fill('input[type="email"]', 'admin@siakad.com');
    await page.fill('input[type="password"]', 'passwordsalah123');
    await page.click('button[type="submit"]');

    // Verifikasi muncul pesan error (biasanya muncul alert atau text merah)
    // Svelte template Siakad ini mungkin memunculkan text "Email atau password salah"
    const errorMessage = page.getByText(/salah|tidak valid/i);
    await expect(errorMessage).toBeVisible({ timeout: 5000 });
  });
});
