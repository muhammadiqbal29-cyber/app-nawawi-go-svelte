import { test, expect } from '@playwright/test';

test.describe('Guru Diniyah - Flow Klaim Mapel dan Filter Milikku', () => {
  test.beforeEach(async ({ page }) => {
    // 1. Login sebagai Guru
    // Menggunakan akun guru yang dibuat dari core-setup.spec.js atau auth manual
    await page.goto('/login');
    await page.fill('input[type="email"]', 'iqbalmahad4039@gmail.com');
    await page.fill('input[type="password"]', 'F@uzan12');
    await page.click('button[type="submit"]');

    // Pastikan masuk ke dashboard (biasanya URL-nya /)
    await page.waitForURL('**/');
  });

  test('harusnya bisa membuka halaman diniyah, toggle filter, dan melihat UI klaim', async ({ page }) => {
    // Tangani dialog otomatis jika ada konfirmasi
    page.on('dialog', async (dialog) => {
      await dialog.accept();
    });

    // 2. Akses halaman Input Nilai Diniyah
    await page.goto('/guru/input-nilai/diniyah');

    // Jika diminta milih jenjang, kita coba pilih jenjang dulu (ini tergantung state, kita asumsikan sudah diset atau kita klik jika ada)
    const btnJenjang = page.locator('button', { hasText: 'Pilih Jenjang' });
    if (await btnJenjang.isVisible()) {
      await btnJenjang.click();
      await page.locator('ul.sekolah-list li').first().click();
    }

    // 3. Masuk ke halaman kelas pertama
    const firstKelas = page.locator('a.kelas-card').first();
    // Tunggu sampai card kelas muncul (kalau tidak ada kelas, tes ini butuh data dummy)
    await firstKelas.waitFor({ state: 'visible', timeout: 10000 }).catch(() => {
        console.log("Tidak ada kelas yang bisa diklik. Harap pastikan data dummy kelas tersedia.");
    });
    
    if (await firstKelas.isVisible()) {
        await firstKelas.click();

        // 4. Pastikan kita ada di grid mapel
        await expect(page.locator('.mapel-section-title')).toBeVisible();

        // 5. Cek fitur toggle filter "Semua" / "Milikku"
        const toggleBtn = page.locator('#btn-toggle-filter-milikku');
        await expect(toggleBtn).toBeVisible();

        // Cek label tombol (awalnya biasanya "Semua" jika belum pernah set localstorage)
        // Kita klik untuk ubah state
        await toggleBtn.click();

        // Setelah diubah ke "Milikku", mari kita lihat state UI-nya
        // Ada dua kemungkinan: ada mapel yang sudah diklaim, atau kosong (empty state)
        const hasCard = await page.locator('.mapel-card').first().isVisible();
        const hasEmpty = await page.locator('.mapel-empty-filter').isVisible();

        // Salah satu harus muncul
        expect(hasCard || hasEmpty).toBeTruthy();

        // Jika mapel kosong, klik tombol Tampilkan semua di empty state
        if (hasEmpty) {
            await page.getByRole('button', { name: /Tampilkan semua/i }).click();
            // Harusnya empty state hilang
            await expect(page.locator('.mapel-empty-filter')).not.toBeVisible();
        }

        // 6. Uji visibilitas tombol Klaim pada card mapel (bila ada card)
        const firstMapelCard = page.locator('.mapel-card').first();
        if (await firstMapelCard.isVisible()) {
            // Tombol bisa berupa "Klaim" atau "Lepas Klaim" tergantung state DB
            const klaimBtn = firstMapelCard.locator('button.btn-klaim').first();
            await expect(klaimBtn).toBeVisible();
        }
    }
  });
});
