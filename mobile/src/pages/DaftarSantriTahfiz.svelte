<script lang="ts">
    import { onMount } from 'svelte';
    import { navigate } from '../store/nav.svelte';
    import Icon from '../components/Icon.svelte';
    import { getHalaqahKu, getAnggotaHalaqah } from '../services/endpoints';

    let halaqahList = $state<any[]>([]);
    let selectedHalaqah = $state<any>(null);
    let santriList = $state<any[]>([]);
    let isLoading = $state(true);

    onMount(async () => {
        try {
            const res = await getHalaqahKu();
            halaqahList = res || [];
            if (halaqahList.length > 0) {
                await selectHalaqah(halaqahList[0]);
            }
        } catch (e) {
            console.error("Gagal memuat halaqah", e);
        } finally {
            isLoading = false;
        }
    });

    async function selectHalaqah(halaqah: any) {
        selectedHalaqah = halaqah;
        santriList = [];
        isLoading = true;
        try {
            const res = await getAnggotaHalaqah(halaqah.id);
            santriList = res || [];
        } catch (e) {
            console.error("Gagal memuat anggota", e);
        } finally {
            isLoading = false;
        }
    }

    function gotoInputSiswa(anggota: any) {
        navigate('input_tahfiz', {
            id_siswa: anggota.siswa?.id || anggota.id_siswa,
            nama_siswa: anggota.siswa?.nama_lengkap,
            id_kelompok: selectedHalaqah.id
        });
    }

    function gotoKelolaHalaqah() {
        navigate('kelola_halaqah', { 
            id_halaqah: selectedHalaqah ? selectedHalaqah.id : null,
            nama_halaqah: selectedHalaqah ? selectedHalaqah.nama_halaqah : ''
        });
    }
</script>

<div class="mobile-app-wrapper">
    <div class="app-container">
        <!-- Header -->
        <div class="page-header">
            <button class="back-btn" onclick={() => navigate('akademik')}>
                <Icon name="arrow-left" /> Kembali
            </button>
            <h1 class="page-title">Pilih Santri</h1>
            <p class="page-subtitle">Tap santri untuk input nilai tahfiz</p>
        </div>

        <div class="scrollable-content">
            <div class="content-padding">

                {#if isLoading && halaqahList.length === 0}
                    <div class="loading-state">Memuat data halaqah...</div>

                {:else if halaqahList.length === 0}
                    <div class="empty-halaqah">
                        <div class="empty-icon-wrap">
                            <Icon name="users-slash" />
                        </div>
                        <h2>Belum Ada Halaqah</h2>
                        <p>Anda belum mengelola halaqah tahfiz.</p>
                        <button class="btn-primary mt-4" onclick={() => navigate('kelola_halaqah')}>
                            <Icon name="plus" /> Buat Halaqah Baru
                        </button>
                    </div>
                {:else}

                    {#if halaqahList.length > 1}
                        <div class="form-group mb-4">
                            <select class="form-select" onchange={(e) => selectHalaqah(halaqahList[e.currentTarget.selectedIndex])}>
                                {#each halaqahList as halaqah}
                                    <option value={halaqah.id}>{halaqah.nama_halaqah}</option>
                                {/each}
                            </select>
                        </div>
                    {/if}

                    {#if santriList.length === 0}
                        <div class="empty-state">Belum ada santri di halaqah ini.</div>
                    {:else}
                        <div class="section-header">
                            <h2 class="section-title">Daftar Santri</h2>
                            <button class="btn-outline-sm" onclick={gotoKelolaHalaqah}>
                                <Icon name="users-gear" /> Kelola Anggota
                            </button>
                        </div>
                        
                        <div class="santri-list">
                            {#each santriList as anggota}
                                <button class="santri-pick-card" onclick={() => gotoInputSiswa(anggota)}>
                                    <div class="avatar">{anggota.siswa?.nama_lengkap?.substring(0,2).toUpperCase() || 'SN'}</div>
                                    <div class="santri-info">
                                        <h3 class="santri-name">{anggota.siswa?.nama_lengkap || 'Siswa Tanpa Nama'}</h3>
                                        <p class="santri-nis">
                                            NIS: {anggota.siswa?.nis || '-'}
                                            {#if anggota.siswa?.nama_kelas}
                                                | {anggota.siswa.nama_kelas} {anggota.siswa.rombel && anggota.siswa.rombel !== '__kosong__' ? anggota.siswa.rombel : ''}
                                            {/if}
                                        </p>
                                    </div>
                                    <div class="icon-go">
                                        <Icon name="pen-to-square" />
                                    </div>
                                </button>
                            {/each}
                        </div>
                    {/if}
                {/if}

            </div>
        </div>
    </div>
</div>

<style>
    .mobile-app-wrapper {
        display: flex; align-items: center; justify-content: center;
        min-height: 100vh; width: 100vw; background-color: #f3f4f6; font-family: 'Inter', sans-serif;
    }
    .app-container {
        background-color: #f9fafb; width: 100%; min-height: 100vh;
        position: relative; display: flex; flex-direction: column; overflow: hidden;
    }
    @media (min-width: 768px) {
        .mobile-app-wrapper { padding: 3rem 1.5rem; }
        .app-container { max-width: 900px; min-height: 0; max-height: 90vh; border-radius: 1.5rem; border: 1px solid #e5e7eb; box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1); }
    }

    /* Header */
    .page-header { background-color: #06b6d4; color: white; padding: 1.5rem; position: relative; z-index: 10; border-bottom: 4px solid #0891b2; text-align: center; }
    @media (min-width: 768px) { .page-header { padding: 2rem; } }
    .back-btn { position: absolute; top: 1.5rem; left: 1.5rem; color: rgba(255,255,255,0.8); background: transparent; border: none; cursor: pointer; font-size: 0.875rem; font-weight: 600; display: flex; align-items: center; gap: 0.5rem; }
    .back-btn:hover { color: white; }
    .page-title { font-size: 1.5rem; font-weight: 900; margin: 2rem 0 0.25rem; }
    @media (min-width: 768px) { .page-title { margin-top: 0; } }
    .page-subtitle { font-size: 0.875rem; font-weight: 600; color: rgba(255,255,255,0.9); margin: 0; }

    /* Content */
    .scrollable-content { flex-grow: 1; overflow-y: auto; -webkit-overflow-scrolling: touch; padding-bottom: 2rem; }
    .content-padding { padding: 1.5rem; }
    .mb-4 { margin-bottom: 1rem; }
    .mt-4 { margin-top: 1rem; }

    .loading-state, .empty-state { text-align: center; color: #6b7280; font-weight: 500; margin-top: 2rem; }

    /* Empty Halaqah */
    .empty-halaqah { text-align: center; padding: 3rem 1rem; color: #4b5563; }
    .empty-icon-wrap { font-size: 3rem; color: #9ca3af; margin-bottom: 1rem; }
    .empty-halaqah h2 { font-size: 1.25rem; font-weight: 700; color: #1f2937; margin: 0 0 0.5rem; }
    .empty-halaqah p { font-size: 0.875rem; margin-bottom: 0; }

    .btn-primary {
        background: linear-gradient(135deg, #0ea5e9 0%, #2563eb 100%); color: white; border: none;
        padding: 0.75rem 1.5rem; border-radius: 0.75rem; font-size: 1rem; font-weight: 600; cursor: pointer;
        display: inline-flex; align-items: center; gap: 0.5rem;
        box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.3); transition: transform 0.2s;
    }
    .btn-primary:active { transform: scale(0.95); }

    .form-group { display: flex; flex-direction: column; }
    .form-select {
        width: 100%; padding: 0.875rem 1rem; border-radius: 0.75rem; border: 1px solid #d1d5db;
        background-color: white; font-size: 0.875rem; font-family: inherit; color: #1f2937; font-weight: 600; outline: none;
    }
    .form-select:focus { border-color: #06b6d4; box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1); }

    /* Section Header */
    .section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
    .section-title { font-size: 1.125rem; font-weight: 700; color: #1f2937; margin: 0; }
    .btn-outline-sm {
        background: white; border: 1px solid #d1d5db; color: #4b5563; padding: 0.4rem 0.75rem;
        border-radius: 0.5rem; font-size: 0.75rem; font-weight: 600; cursor: pointer;
        display: inline-flex; align-items: center; gap: 0.3rem; transition: background 0.2s;
    }
    .btn-outline-sm:active { background: #f3f4f6; }

    /* Santri Pick Card (mode pilih siswa) */
    .santri-list { display: flex; flex-direction: column; gap: 0.75rem; }
    .santri-pick-card {
        background: white; border: 1px solid #e5e7eb; border-radius: 1rem; padding: 1rem;
        display: flex; align-items: center; gap: 1rem; width: 100%; text-align: left;
        cursor: pointer; transition: all 0.2s; box-shadow: 0 1px 3px rgba(0,0,0,0.05);
    }
    .santri-pick-card:hover { border-color: #06b6d4; background: #f8fafc; }
    .santri-pick-card:active { transform: scale(0.98); }
    .avatar { width: 3rem; height: 3rem; background: #cffafe; color: #0891b2; font-weight: 800; border-radius: 9999px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
    .santri-info { flex-grow: 1; display: flex; flex-direction: column; gap: 0.25rem; }
    .santri-name { font-weight: 700; color: #1f2937; margin: 0; font-size: 1rem; }
    .santri-nis { font-size: 0.75rem; color: #6b7280; margin: 0; }
    .icon-go { color: #06b6d4; display: flex; align-items: center; justify-content: center; width: 2.5rem; height: 2.5rem; background: #f3f4f6; border-radius: 9999px; }
    .santri-pick-card:hover .icon-go { background: #06b6d4; color: white; }
</style>
