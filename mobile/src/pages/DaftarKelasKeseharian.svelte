<script lang="ts">
    import { onMount } from 'svelte';
    import { navigate } from '../store/nav.svelte';
    import Icon from '../components/Icon.svelte';
    import { getKelasKeseharian } from '../services/endpoints';
    import { guruState } from '../store/guru.svelte';

    let kelasList = $state<any[]>([]);
    let isLoading = $state(true);

    onMount(async () => {
        try {
            const res = await getKelasKeseharian(guruState.activeSekolahId || 1);
            kelasList = res || [];
        } catch (e) {
            console.error("Gagal memuat kelas", e);
        } finally {
            isLoading = false;
        }
    });

    function goToInput(kelas: any) {
        navigate('input_keseharian', { 
            kelas_id: kelas.id,
            rombel: kelas.rombel, 
            nama_kelas: `${kelas.tingkat} - ${kelas.nama_kelas}${kelas.rombel && kelas.rombel !== '__kosong__' ? ' Rombel ' + kelas.rombel : ''}`
        });
    }
</script>

<div class="mobile-app-wrapper">
    <div class="app-container">
        
        <!-- HEADER -->
        <div class="header">
            <button class="back-btn" onclick={() => navigate('akademik')}>
                <Icon name="arrow-left" />
            </button>
            <h1 class="header-title">Pilih Kelas</h1>
        </div>

        <div class="content-container">
            {#if isLoading}
                <div class="loading-state">
                    <div class="spinner"></div>
                    <p>Memuat kelas...</p>
                </div>
            {:else if kelasList.length === 0}
                <div class="empty-state">
                    <div class="empty-icon">
                        <Icon name="folder-open" />
                    </div>
                    <p>Belum ada kelas yang ditugaskan kepada Anda.</p>
                </div>
            {:else}
                <div class="kelas-list">
                    {#each kelasList as kelas}
                        <button class="kelas-card" onclick={() => goToInput(kelas)}>
                            <div class="kelas-icon">
                                <Icon name="users" />
                            </div>
                            <div class="kelas-info">
                                <h3>{kelas.tingkat} - {kelas.nama_kelas}</h3>
                                {#if kelas.rombel && kelas.rombel !== '__kosong__'}
                                    <p class="rombel-badge">Rombel {kelas.rombel}</p>
                                {/if}
                                <p class="siswa-count">{kelas.jumlah_siswa} Siswa</p>
                            </div>
                            <div class="chevron">
                                <Icon name="chevron-right" />
                            </div>
                        </button>
                    {/each}
                </div>
            {/if}
        </div>
    </div>
</div>

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
            overflow: hidden;
        }
    }

    .header {
        background-color: #ffffff;
        padding: 1rem;
        display: flex;
        align-items: center;
        gap: 1rem;
        border-bottom: 1px solid #f3f4f6;
        position: sticky;
        top: 0;
        z-index: 50;
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
    @keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }

    .empty-icon {
        font-size: 3rem;
        color: #d1d5db;
        margin-bottom: 1rem;
    }

    .kelas-list {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
    }

    .kelas-card {
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
    }

    .kelas-card:active {
        transform: scale(0.98);
        background: #f9fafb;
    }

    .kelas-icon {
        width: 3rem;
        height: 3rem;
        background: #ecfdf5;
        color: #10b981;
        border-radius: 0.75rem;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 1.25rem;
        flex-shrink: 0;
    }

    .kelas-info {
        flex-grow: 1;
    }

    .kelas-info h3 {
        margin: 0;
        font-size: 1rem;
        font-weight: 700;
        color: #1f2937;
    }

    .rombel-badge {
        display: inline-block;
        background: #f3f4f6;
        color: #4b5563;
        font-size: 0.75rem;
        font-weight: 600;
        padding: 0.125rem 0.5rem;
        border-radius: 999px;
        margin: 0.25rem 0 0 0;
    }

    .siswa-count {
        margin: 0.25rem 0 0 0;
        font-size: 0.875rem;
        color: #6b7280;
    }

    .chevron {
        color: #9ca3af;
    }
</style>
