import { get, post, put } from './api';

// ==========================================
// GURU & AUTHENTICATION
// ==========================================

export async function getCurrentGuru() {
    return get<any>('/api/guru/me/profil');
}

export async function changePassword(payload: { old_password: string; new_password: string }) {
    return put<{ message: string }>('/api/guru/change-password', payload);
}

export async function getActiveTahunAjaran() {
    const list = await get<any[]>('/api/tahun-ajaran');
    return list.find(ta => ta.status === 'Active');
}

export async function clockInGuru(payload: { latitude: number; longitude: number }) {
    return post<any>('/api/guru/me/presensi/masuk', payload);
}

export async function clockOutGuru(payload: { latitude: number; longitude: number }) {
    return post<any>('/api/guru/me/presensi/pulang', payload);
}

export async function getRiwayatAbsenGuru() {
    return get<any[]>('/api/guru/me/presensi/riwayat');
}

export async function getKonfigurasiPresensi() {
    return get<any>('/api/guru/me/presensi/konfigurasi');
}

// ==========================================
// TAHFIZ HALAQAH
// ==========================================

export async function getHalaqahKu() {
    const res = await get<any[]>('/api/guru/me/halaqah');
    return res || [];
}

export async function createHalaqah(payload: { nama_halaqah?: string }) {
    return post<any>('/api/guru/me/halaqah', payload);
}

export async function getAnggotaHalaqah(idKelompok: number) {
    const data = await get<any[]>(`/api/guru/halaqah/${idKelompok}/anggota`);
    return data || [];
}

export async function updateAnggotaHalaqah(idKelompok: number, siswa_ids: number[]) {
    return put<any>(`/api/guru/halaqah/${idKelompok}/anggota`, { siswa_ids });
}

export async function getSiswaAvailableTahfiz() {
    const data = await get<any[]>('/api/guru/sekolah/1/siswa-available-tahfiz');
    return data || [];
}

export async function getRiwayatTahfiz(idSiswa: number, idKelompok: number) {
    const res = await get<any>(`/api/siswa/${idSiswa}/tahfiz/harian`);
    return res.data || [];
}

export async function getMateriTahfizAktif(sekolahId: number) {
    const data = await get<any[]>(`/api/guru/sekolah/${sekolahId}/materi-tahfiz`);
    return data || [];
}

export async function postSetoranTahfiz(payload: any) {
    return post<any>(`/api/siswa/${payload.id_siswa}/tahfiz/harian`, payload);
}

export async function getRiwayatTahsin(idSiswa: number, idKelompok: number) {
    return []; // Optional for now
}

export async function postSetoranTahsin(payload: any) {
    return post<any>(`/api/siswa/${payload.id_siswa}/tahsin/harian`, payload);
}

// ==========================================
// KESEHARIAN
// ==========================================

export async function getKelasKeseharian(sekolahId: number) {
    const res = await get<any[]>(`/api/guru/sekolah/${sekolahId}/kelas`);
    return res || [];
}

export async function getIndikatorKeseharian(sekolahId: number) {
    const data = await get<any>(`/api/guru/sekolah/${sekolahId}/keseharian/indikator`);
    return data?.data || [];
}

export async function getSiswaKelas(kelasId: number, rombel: string = '') {
    const data = await get<any>(`/api/guru/kelas/${kelasId}/siswa?rombel=${rombel}`);
    return Array.isArray(data) ? data : (data?.data || []);
}

export async function submitSesiKeseharian(payload: any) {
    return post<any>('/api/guru/keseharian/sesi', payload);
}
