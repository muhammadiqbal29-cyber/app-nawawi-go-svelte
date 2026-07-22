<script lang="ts">
    import { onMount } from 'svelte';
    import { navigate, routerState } from '../store/nav.svelte';
    import Icon from '../components/Icon.svelte';
    import Toast from '../components/Toast.svelte';
    import { getIndikatorKeseharian, getSiswaKelas, submitSesiKeseharian } from '../services/endpoints';
    import { guruState } from '../store/guru.svelte';

    let showToast = $state(false);
    let toastMessage = $state('');
    let toastType = $state<'success' | 'error' | 'info'>('success');
    
    let isLoading = $state(false);
    let isFetching = $state(true);
    let isSubmitting = $state(false);

    let kelasId = $derived(routerState.navParams?.kelas_id);
    let rombel = $derived(routerState.navParams?.rombel);
    let namaKelas = $derived(routerState.navParams?.nama_kelas || 'Kelas');

    let indikatorList = $state<any[]>([]);
    let siswaList = $state<any[]>([]);
    let tanggal = $state(new Date().toISOString().split('T')[0]);

    // State untuk Modal
    let isModalOpen = $state(false);
    let selectedSiswa = $state<any>(null);
    let checklist = $state<{ [indikator_id: number]: boolean }>({});
    let catatan = $state('');

    onMount(async () => {
        try {
            indikatorList = await getIndikatorKeseharian(guruState.activeSekolahId || 1);
            if (kelasId) {
                siswaList = await getSiswaKelas(kelasId, rombel);
            }
        } catch (e: any) {
            console.error("Gagal load data keseharian", e);
            toastMessage = "Gagal memuat data. Periksa koneksi Anda.";
            toastType = "error";
            showToast = true;
        } finally {
            isFetching = false;
        }
    });

    function openModal(siswa: any) {
        selectedSiswa = siswa;
        catatan = '';
        
        // Reset checklist
        const initCheck: any = {};
        for (const ind of indikatorList) {
            initCheck[ind.id] = false;
        }
        checklist = initCheck;
        
        isModalOpen = true;
    }

    function closeModal() {
        isModalOpen = false;
        selectedSiswa = null;
    }

    function toggleIndikator(id: number) {
        checklist[id] = !checklist[id];
    }

    async function handleSimpanPelanggaran() {
        if (!selectedSiswa) return;

        // Ambil indikator yang dicentang
        const selectedIndikatorIds = Object.keys(checklist).filter(id => checklist[Number(id)]);

        if (selectedIndikatorIds.length === 0) {
            toastMessage = 'Pilih minimal satu pelanggaran / indikator!';
            toastType = 'error';
            showToast = true;
            return;
        }

        isSubmitting = true;
        try {
            // Karena backend mengharapkan 1 indikator per payload (untuk banyak siswa),
            // kita lakukan perulangan untuk setiap indikator yang dipilih (untuk 1 siswa ini).
            const promises = selectedIndikatorIds.map(indId => {
                const payload = {
                    sekolah_id: guruState.activeSekolahId || 1,
                    indikator_id: Number(indId),
                    tanggal: tanggal,
                    catatan: catatan,
                    siswa_list: [
                        {
                            siswa_id: selectedSiswa.id,
                            status: true // true = terjadi anomali / pelanggaran
                        }
                    ]
                };
                return submitSesiKeseharian(payload);
            });

            await Promise.all(promises);

            toastMessage = `Pelanggaran untuk ${selectedSiswa.nama_lengkap} berhasil disimpan!`;
            toastType = 'success';
            showToast = true;
            
            closeModal();
        } catch (e: any) {
            toastMessage = e.message || 'Gagal menyimpan data';
            toastType = 'error';
            showToast = true;
        } finally {
            isSubmitting = false;
        }
    }
</script>

<div class="mobile-app-wrapper">
    <div class="app-container">
        
        <div class="header">
            <button class="back-btn" onclick={() => navigate('daftar_kelas_keseharian')} disabled={isFetching || isSubmitting}>
                <Icon name="arrow-left" />
            </button>
            <h1 class="header-title">Input Keseharian</h1>
        </div>

        <div class="content-container" class:blurred={isModalOpen}>
            {#if isFetching}
                <div class="loading-state">
                    <div class="spinner"></div>
                    <p>Memuat data...</p>
                </div>
            {:else}
                <div class="info-card">
                    <div class="info-row" style="margin-bottom: 0.5rem;">
                        <span class="info-label">Kelas</span>
                        <span class="info-value">{namaKelas}</span>
                    </div>
                    <div class="form-group">
                        <label for="tanggal" style="font-size: 0.8rem; color: #6b7280; font-weight: 600;">Tanggal</label>
                        <input type="date" id="tanggal" class="form-input" bind:value={tanggal} />
                    </div>
                </div>

                <div class="siswa-section">
                    <div class="section-header">
                        <h3>Daftar Siswa</h3>
                        <span class="siswa-count">{siswaList.length} Anak</span>
                    </div>
                    <p class="instruction-text">
                        <Icon name="circle-info" /> Pilih siswa untuk mencatat pelanggaran / anomali harian.
                    </p>

                    {#if siswaList.length === 0}
                        <div class="empty-state">Belum ada siswa di kelas ini.</div>
                    {:else}
                        <div class="siswa-list">
                            {#each siswaList as siswa}
                                <button class="siswa-card" onclick={() => openModal(siswa)}>
                                    <div class="siswa-avatar">
                                        <Icon name="user" />
                                    </div>
                                    <div class="siswa-info">
                                        <p class="siswa-name">{siswa.nama_lengkap}</p>
                                        <p class="siswa-nisn">NISN: {siswa.nisn || '-'}</p>
                                    </div>
                                    <div class="chevron">
                                        <Icon name="chevron-right" />
                                    </div>
                                </button>
                            {/each}
                        </div>
                    {/if}
                </div>
            {/if}
        </div>

        <!-- MODAL BOTTOM SHEET -->
        {#if isModalOpen}
            <div class="modal-backdrop" onclick={closeModal}></div>
            <div class="bottom-sheet">
                <div class="sheet-header">
                    <h3>Catat Pelanggaran</h3>
                    <button class="btn-close" onclick={closeModal}><Icon name="xmark" /></button>
                </div>
                
                <div class="sheet-content">
                    {#if selectedSiswa}
                        <div class="target-siswa">
                            <div class="siswa-avatar-small"><Icon name="user" /></div>
                            <div class="target-info">
                                <strong>{selectedSiswa.nama_lengkap}</strong>
                                <span>{namaKelas}</span>
                            </div>
                        </div>

                        <div class="form-group" style="margin-top: 1rem;">
                            <label style="font-weight: 700; margin-bottom: 0.5rem; display: block;">Pilih Indikator Pelanggaran</label>
                            
                            {#if indikatorList.length === 0}
                                <div class="empty-state" style="padding: 1rem;">Belum ada indikator yang dikonfigurasi.</div>
                            {:else}
                                <div class="indikator-list">
                                    {#each indikatorList as ind}
                                        <button class="indikator-item {checklist[ind.id] ? 'selected' : ''}" onclick={() => toggleIndikator(ind.id)}>
                                            <div class="ind-info">
                                                <span class="ind-nama">{ind.nama}</span>
                                                <span class="ind-minus">Poin -{ind.nilai_pengurang || 1}</span>
                                            </div>
                                            <div class="custom-checkbox {checklist[ind.id] ? 'checked' : ''}">
                                                {#if checklist[ind.id]}
                                                    <Icon name="check" />
                                                {/if}
                                            </div>
                                        </button>
                                    {/each}
                                </div>
                            {/if}
                        </div>

                        <div class="form-group" style="margin-top: 1rem;">
                            <label for="catatan" style="font-weight: 700; display: block; margin-bottom: 0.5rem;">Catatan (Opsional)</label>
                            <textarea id="catatan" class="form-input" bind:value={catatan} rows="2" placeholder="Tulis catatan jika perlu..."></textarea>
                        </div>
                    {/if}
                </div>

                <div class="sheet-footer">
                    <button class="btn-submit" onclick={handleSimpanPelanggaran} disabled={isSubmitting}>
                        {#if isSubmitting}
                            <span class="spinner-small"></span> Memproses...
                        {:else}
                            <Icon name="save" /> Simpan Pelanggaran
                        {/if}
                    </button>
                </div>
            </div>
        {/if}
    </div>
</div>

{#if showToast}
    <Toast message={toastMessage} type={toastType} onClose={() => showToast = false} />
{/if}

<style>
    .mobile-app-wrapper {
        display: flex;
        align-items: center;
        justify-content: center;
        min-height: 100vh;
        width: 100vw;
        background-color: #f3f4f6;
        font-family: 'Inter', sans-serif;
    }

    .app-container {
        background-color: #f9fafb;
        width: 100%;
        min-height: 100vh;
        position: relative;
        display: flex;
        flex-direction: column;
        overflow: hidden;
    }

    @media (min-width: 768px) {
        .mobile-app-wrapper { padding: 3rem 1.5rem; }
        .app-container {
            max-width: 480px;
            min-height: 800px;
            max-height: 90vh;
            border-radius: 1.5rem;
            border: 1px solid #e5e7eb;
            box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1);
        }
    }

    .header {
        background-color: #ffffff;
        padding: 1rem;
        display: flex;
        align-items: center;
        gap: 1rem;
        border-bottom: 1px solid #f3f4f6;
        position: relative;
        z-index: 10;
    }

    .back-btn {
        background: transparent;
        border: none;
        padding: 0.5rem;
        cursor: pointer;
        color: #4b5563;
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: 0.5rem;
    }

    .header-title {
        font-size: 1.25rem;
        font-weight: 700;
        color: #111827;
        margin: 0;
    }

    .content-container {
        padding: 1rem;
        flex-grow: 1;
        overflow-y: auto;
        transition: filter 0.3s ease;
    }

    .content-container.blurred {
        filter: blur(2px);
    }

    .loading-state, .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 4rem 1rem;
        text-align: center;
        color: #6b7280;
    }

    .spinner {
        border: 3px solid #f3f4f6;
        border-top: 3px solid #10b981;
        border-radius: 50%;
        width: 2rem;
        height: 2rem;
        animation: spin 1s linear infinite;
        margin-bottom: 1rem;
    }

    .spinner-small {
        border: 2px solid rgba(255,255,255,0.3);
        border-top: 2px solid #ffffff;
        border-radius: 50%;
        width: 1rem;
        height: 1rem;
        animation: spin 1s linear infinite;
    }
    
    @keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }

    .info-card {
        background: white;
        border-radius: 1rem;
        padding: 1rem;
        border: 1px solid #e5e7eb;
        margin-bottom: 1.25rem;
    }

    .info-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .info-label {
        color: #6b7280;
        font-size: 0.875rem;
        font-weight: 600;
    }

    .info-value {
        color: #1f2937;
        font-weight: 700;
        font-size: 0.95rem;
    }

    .form-input {
        width: 100%;
        padding: 0.75rem 1rem;
        border: 1px solid #d1d5db;
        border-radius: 0.75rem;
        font-size: 1rem;
        font-family: inherit;
        background: #f9fafb;
        color: #1f2937;
        box-sizing: border-box;
    }
    
    .form-input:focus {
        outline: none;
        border-color: #10b981;
        box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
    }

    .siswa-section {
        margin-bottom: 1rem;
    }

    .section-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 0.5rem;
        padding: 0 0.25rem;
    }

    .section-header h3 {
        margin: 0;
        font-size: 1.125rem;
        font-weight: 800;
        color: #1f2937;
    }

    .siswa-count {
        background: #e5e7eb;
        color: #4b5563;
        font-size: 0.75rem;
        font-weight: 700;
        padding: 0.25rem 0.75rem;
        border-radius: 999px;
    }

    .instruction-text {
        font-size: 0.8rem;
        color: #6b7280;
        margin: 0 0.25rem 1rem 0.25rem;
        line-height: 1.4;
    }

    .siswa-list {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
    }

    .siswa-card {
        background: white;
        border: 1px solid #e5e7eb;
        border-radius: 1rem;
        padding: 1rem;
        display: flex;
        align-items: center;
        gap: 1rem;
        text-align: left;
        cursor: pointer;
        transition: all 0.2s;
        width: 100%;
    }

    .siswa-card:active {
        transform: scale(0.98);
        background: #f9fafb;
    }

    .siswa-avatar {
        width: 2.5rem;
        height: 2.5rem;
        background: #f3f4f6;
        color: #9ca3af;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 1rem;
        flex-shrink: 0;
    }

    .siswa-info {
        flex-grow: 1;
    }

    .siswa-name {
        margin: 0;
        font-weight: 700;
        font-size: 1rem;
        color: #1f2937;
    }

    .siswa-nisn {
        margin: 0.25rem 0 0 0;
        font-size: 0.75rem;
        color: #6b7280;
    }

    .chevron {
        color: #9ca3af;
    }

    /* Modal Bottom Sheet */
    .modal-backdrop {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0,0,0,0.4);
        z-index: 40;
    }

    .bottom-sheet {
        position: absolute;
        bottom: 0;
        left: 0;
        right: 0;
        background: white;
        border-top-left-radius: 1.5rem;
        border-top-right-radius: 1.5rem;
        z-index: 50;
        display: flex;
        flex-direction: column;
        max-height: 85vh;
        box-shadow: 0 -10px 25px rgba(0,0,0,0.1);
        animation: slideUp 0.3s ease-out;
    }

    @keyframes slideUp {
        from { transform: translateY(100%); }
        to { transform: translateY(0); }
    }

    .sheet-header {
        padding: 1.25rem 1.5rem;
        display: flex;
        justify-content: space-between;
        align-items: center;
        border-bottom: 1px solid #f3f4f6;
    }

    .sheet-header h3 {
        margin: 0;
        font-size: 1.125rem;
        font-weight: 800;
        color: #111827;
    }

    .btn-close {
        background: #f3f4f6;
        border: none;
        width: 2rem;
        height: 2rem;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #4b5563;
        cursor: pointer;
    }

    .sheet-content {
        padding: 1.5rem;
        overflow-y: auto;
    }

    .target-siswa {
        display: flex;
        align-items: center;
        gap: 1rem;
        background: #f9fafb;
        padding: 1rem;
        border-radius: 1rem;
        border: 1px solid #e5e7eb;
    }

    .siswa-avatar-small {
        width: 2.5rem;
        height: 2.5rem;
        background: #e0e7ff;
        color: #4f46e5;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    .target-info {
        display: flex;
        flex-direction: column;
    }

    .target-info strong {
        color: #111827;
        font-size: 1rem;
    }
    
    .target-info span {
        color: #6b7280;
        font-size: 0.8rem;
    }

    .indikator-list {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    .indikator-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        background: white;
        border: 1px solid #e5e7eb;
        padding: 1rem;
        border-radius: 0.75rem;
        cursor: pointer;
        transition: all 0.2s;
        text-align: left;
    }

    .indikator-item.selected {
        background: #fef2f2;
        border-color: #f87171;
    }

    .ind-info {
        display: flex;
        flex-direction: column;
    }

    .ind-nama {
        font-weight: 600;
        color: #374151;
        font-size: 0.95rem;
    }
    
    .indikator-item.selected .ind-nama {
        color: #b91c1c;
    }

    .ind-minus {
        font-size: 0.75rem;
        color: #ef4444;
        font-weight: 700;
        margin-top: 0.25rem;
    }

    .custom-checkbox {
        width: 1.5rem;
        height: 1.5rem;
        border: 2px solid #d1d5db;
        border-radius: 6px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: white;
        color: white;
        transition: all 0.2s;
        font-size: 0.8rem;
    }

    .custom-checkbox.checked {
        background: #ef4444;
        border-color: #ef4444;
    }

    .sheet-footer {
        padding: 1.25rem 1.5rem;
        border-top: 1px solid #f3f4f6;
        background: white;
    }

    .btn-submit {
        width: 100%;
        background: #ef4444;
        color: white;
        border: none;
        padding: 1rem;
        border-radius: 1rem;
        font-size: 1rem;
        font-weight: 700;
        font-family: inherit;
        display: flex;
        justify-content: center;
        align-items: center;
        gap: 0.5rem;
        cursor: pointer;
        transition: all 0.2s;
        box-shadow: 0 4px 6px -1px rgba(239, 68, 68, 0.2);
    }

    .btn-submit:active {
        transform: scale(0.98);
    }

    .btn-submit:disabled {
        background: #9ca3af;
        cursor: not-allowed;
        box-shadow: none;
        transform: none;
    }
</style>
