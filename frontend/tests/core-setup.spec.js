import { test, expect } from '@playwright/test';

test.describe('E2E Core Setup Flow (Happy Path)', () => {
  // Use a longer timeout for E2E flow
  test.setTimeout(60000);

  test('admin dapat melakukan setup master data dari awal', async ({ page }) => {
    // Tangani semua alert dan konfirmasi secara otomatis
    page.on('dialog', async (dialog) => {
      await dialog.accept();
    });

    // ---------------------------------------------------------
    // 1. LOGIN ADMIN
    // ---------------------------------------------------------
    await test.step('Login sebagai Admin', async () => {
      await page.goto('/login');
      await page.fill('input[type="email"]', 'admin@siakad.com');
      await page.fill('input[type="password"]', 'password');
      await page.click('button[type="submit"]');
      await page.waitForURL('/');
      await expect(page).toHaveURL('/');
    });

    // ---------------------------------------------------------
    // 2. MASTER INSTITUSI (SEKOLAH) & TAHUN AJARAN
    // ---------------------------------------------------------
    await test.step('Tambah Master Institusi (Sekolah) & TA', async () => {
      // Buka menu master
      await page.goto('/institusi');

      // 2A. Tambah Sekolah
      await page.getByRole('button', { name: /Tambah Unit Sekolah/i }).click();

      const modal = page.locator('.modal-content');
      await expect(modal).toBeVisible();

      await modal.locator('input[placeholder*="NPSN"]').fill('12345678');
      await modal.locator('input[placeholder*="Cth: SMP IT Nawawi"]').fill('SD ISLAM IMAM NAWAWI');
      await modal.locator('select').nth(0).selectOption('SD');
      await modal.locator('select').nth(1).selectOption('Swasta');
      await modal.locator('textarea[placeholder*="Jl. Pendidikan"]').fill('Jl. Pendidikan No. 1');

      // Simpan Sekolah
      await modal.getByRole('button', { name: /simpan/i }).click();
      await expect(modal).not.toBeVisible();
      await expect(page.getByText('SD ISLAM IMAM NAWAWI')).toBeVisible();

      // 2B. Tambah Tahun Ajaran & Aktifkan Semester
      await page.getByRole('button', { name: /^Tahun Ajaran$/i }).click();
      await page.getByRole('button', { name: /Tambah Tahun Ajaran/i }).click();

      const modalTA = page.locator('.modal-content');
      await expect(modalTA).toBeVisible();

      await modalTA.locator('input[type="number"]').fill('2024');
      await modalTA.getByRole('button', { name: /Simpan/i }).click();
      await expect(modalTA).not.toBeVisible();

      // Tunggu data muncul
      await expect(page.getByText('2024/2025').first()).toBeVisible();

      const btnActivate = page.locator('button', { hasText: 'Pilih Semester Ini' }).first();
      await btnActivate.click();

      await expect(page.getByText('Terpilih').first()).toBeVisible();
    });

    // ---------------------------------------------------------
    // 3. MASTER KELAS
    // ---------------------------------------------------------
    await test.step('Tambah Master Kelas', async () => {
      await page.goto('/akademik');

      // Pilih sekolah dengan mengklik card
      await page.getByText('SD ISLAM IMAM NAWAWI').click();

      // Pindah ke tab Master Kelas
      await page.getByRole('button', { name: /Master Kelas/i }).click();

      // Klik tombol tambah kelas
      await page.getByRole('button', { name: /Tambah Kelas/i }).click();

      const modal = page.locator('.modal-content');
      await expect(modal).toBeVisible();

      await modal.locator('input[placeholder*="Cth: 10, 11, VII"]').fill('1');
      await modal.locator('input[type="number"]').fill('1');
      await modal.locator('input[placeholder*="Cth: X IPA 1"]').fill('1A');

      await modal.getByRole('button', { name: /simpan/i }).click();
      await expect(modal).not.toBeVisible();
      await expect(page.getByText('1A')).toBeVisible();
    });

    // ---------------------------------------------------------
    // 4. MASTER GURU
    // ---------------------------------------------------------
    await test.step('Tambah Master Guru', async () => {
      await page.goto('/admin/guru/tambah');

      await page.locator('#nama').fill('Budi Santoso');
      await page.locator('#email').fill('budi@siakad.com');
      await page.locator('#password').fill('password');
      await page.locator('#nik').fill('198001012000011001');

      // Ceklis sekolah untuk memberi hak akses ke SD ISLAM IMAM NAWAWI
      // (checkbox paling atas pada group sekolah tsb)
      await page.getByText('SD ISLAM IMAM NAWAWI').click();

      await page.getByRole('button', { name: /Simpan Data Guru/i }).click();

      // Pastikan kembali ke daftar guru dan data muncul
      await expect(page).toHaveURL(/.*\/admin\/guru/);
      await expect(page.getByText('Budi Santoso')).toBeVisible();
    });

    // ---------------------------------------------------------
    // 5. MASTER SISWA
    // ---------------------------------------------------------
    await test.step('Tambah Master Siswa', async () => {
      await page.goto('/akademik');

      // Pilih sekolah
      await page.getByText('SD ISLAM IMAM NAWAWI').click();

      // Secara default tab aktif adalah "Data Siswa", tapi untuk memastikan:
      await page.getByRole('button', { name: /Data Siswa/i }).click();

      await page.getByRole('button', { name: /Tambah Siswa/i }).click();

      const modal = page.locator('.modal');
      await expect(modal).toBeVisible();

      await modal.locator('input[type="text"]').first().fill('0011223344'); // NISN input
      await modal.getByRole('button', { name: /Cek Data/i }).click();

      // Setelah cek data, input nama lengkap muncul
      await modal.locator('input[type="text"]').nth(1).fill('Ahmad Dhani');
      await modal.locator('select').filter({ hasText: 'Laki-laki' }).selectOption('L');

      await modal.getByRole('button', { name: /Simpan Data Siswa/i }).click();

      await expect(modal).not.toBeVisible();
      await expect(page.getByText('Ahmad Dhani')).toBeVisible();
    });

    // Selesai! Semua fitur inti (happy path) terverifikasi.
  });
});
