USE siakad;

-- Disable foreign key checks temporarily
SET FOREIGN_KEY_CHECKS = 0;

-- Clear existing dummy data if any
DELETE FROM riwayat_kelas;
DELETE FROM siswas;
DELETE FROM kelas;
DELETE FROM sekolahs;
DELETE FROM semesters;
DELETE FROM tahun_ajarans;

-- Create Tahun Ajaran
INSERT INTO tahun_ajarans (id, tahun_awal, tahun_akhir, nama_tahun, status) VALUES (1, 2026, 2027, '2026/2027', 'Active');
INSERT INTO semesters (tahun_ajaran_id, tipe_semester, status_aktif) VALUES (1, 'Ganjil', 1), (1, 'Genap', 0);

-- Create Sekolah
INSERT INTO sekolahs (id, nama_sekolah, npsn, jenjang, status, alamat) VALUES 
(1, 'TK Nawawi', 'NPSN001', 'TK', 'Swasta', 'Jl. Nawawi 1'),
(2, 'SD Nawawi', 'NPSN002', 'SD', 'Swasta', 'Jl. Nawawi 2'),
(3, 'SMP Nawawi', 'NPSN003', 'SMP', 'Swasta', 'Jl. Nawawi 3'),
(4, 'SMA Nawawi', 'NPSN004', 'SMA', 'Swasta', 'Jl. Nawawi 4');

-- Create Kelas
-- TK (Level 1,2)
INSERT INTO kelas (id, sekolah_id, tingkat, tingkat_level, nama_kelas) VALUES
(1, 1, 'TK', 1, 'A'), (2, 1, 'TK', 2, 'B');

-- SD (Level 1..6)
INSERT INTO kelas (id, sekolah_id, tingkat, tingkat_level, nama_kelas) VALUES
(3, 2, 'SD', 1, '1'), (4, 2, 'SD', 2, '2'), (5, 2, 'SD', 3, '3'),
(6, 2, 'SD', 4, '4'), (7, 2, 'SD', 5, '5'), (8, 2, 'SD', 6, '6');

-- SMP (Level 1..3)
INSERT INTO kelas (id, sekolah_id, tingkat, tingkat_level, nama_kelas) VALUES
(9, 3, 'SMP', 1, '7'), (10, 3, 'SMP', 2, '8'), (11, 3, 'SMP', 3, '9');

-- SMA (Level 1..3)
INSERT INTO kelas (id, sekolah_id, tingkat, tingkat_level, nama_kelas) VALUES
(12, 4, 'SMA', 1, '10'), (13, 4, 'SMA', 2, '11'), (14, 4, 'SMA', 3, '12');

-- Create Siswas (1 per kelas)
-- We will just insert 10 students across different classes
INSERT INTO siswas (id, nisn, nis, nama_lengkap, status, jenis_kelamin, sekolah_id) VALUES
(1, '0000000001', 'NIS001', 'Andi (TK A)', 'Aktif', 'L', 1),
(2, '0000000002', 'NIS002', 'Budi (TK B)', 'Aktif', 'L', 1),
(3, '0000000003', 'NIS003', 'Caca (SD 1)', 'Aktif', 'P', 2),
(4, '0000000004', 'NIS004', 'Deni (SD 6)', 'Aktif', 'L', 2),
(5, '0000000005', 'NIS005', 'Eka (SMP 7)', 'Aktif', 'P', 3),
(6, '0000000006', 'NIS006', 'Fani (SMP 9)', 'Aktif', 'P', 3),
(7, '0000000007', 'NIS007', 'Gilang (SMA 10)', 'Aktif', 'L', 4),
(8, '0000000008', 'NIS008', 'Hana (SMA 12)', 'Aktif', 'P', 4),
(9, '0000000009', 'NIS009', 'Iwan (SD 1)', 'Aktif', 'L', 2),
(10, '0000000010', 'NIS010', 'Joko (SD 6)', 'Aktif', 'L', 2);

-- Assign to Riwayat Kelas
INSERT INTO riwayat_kelas (siswa_id, sekolah_id, kelas_id, rombel, tahun_ajaran_id, status_akhir) VALUES
(1, 1, 1, 'A', 1, 'Aktif'),
(2, 1, 2, 'A', 1, 'Aktif'),
(3, 2, 3, 'A', 1, 'Aktif'),
(4, 2, 8, 'A', 1, 'Aktif'),
(5, 3, 9, 'A', 1, 'Aktif'),
(6, 3, 11, 'A', 1, 'Aktif'),
(7, 4, 12, 'A', 1, 'Aktif'),
(8, 4, 14, 'A', 1, 'Aktif'),
(9, 2, 3, 'B', 1, 'Aktif'),
(10, 2, 8, 'B', 1, 'Aktif');

SET FOREIGN_KEY_CHECKS = 1;
