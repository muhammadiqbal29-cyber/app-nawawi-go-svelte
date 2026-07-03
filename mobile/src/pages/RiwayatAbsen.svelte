<script lang="ts">
    import { onMount } from 'svelte';
    import BottomNav from '../components/BottomNav.svelte';
    import Icon from '../components/Icon.svelte';
    import { getRiwayatAbsenGuru } from '../services/endpoints';

    let riwayatList = $state<any[]>([]);
    let isLoading = $state(true);
    let errorMessage = $state('');

    onMount(async () => {
        try {
            riwayatList = await getRiwayatAbsenGuru();
        } catch (error: any) {
            errorMessage = error.message || 'Gagal memuat riwayat absen';
        } finally {
            isLoading = false;
        }
    });

    function formatTanggal(isoString: string) {
        if (!isoString) return '';
        const date = new Date(isoString);
        return date.toLocaleDateString('id-ID', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' });
    }
</script>

<div class="mobile-app-wrapper">
    <div class="app-container">
        <div class="scrollable-content">
            <div class="content-padding">

                <!-- Judul -->
                <div class="page-header">
                    <h1 class="page-title">Riwayat Absen</h1>
                    <p class="page-subtitle">Log kehadiran Anda di sekolah</p>
                </div>

                <!-- List -->
                <div class="riwayat-list">
                    {#if isLoading}
                        <div class="loading-state">
                            <p>Memuat data riwayat absen...</p>
                        </div>
                    {:else if errorMessage}
                        <div class="error-state">
                            <p>{errorMessage}</p>
                        </div>
                    {:else if riwayatList && riwayatList.length > 0}
                        {#each riwayatList as item}
                            <div class="riwayat-card">
                                <div class="riwayat-left">
                                    <div class="riwayat-icon {item.status.includes('Terlambat') ? 'icon-warning' : 'icon-success'}">
                                        <Icon name="fingerprint" />
                                    </div>
                                    <div>
                                        <h3 class="riwayat-status">{item.status}</h3>
                                        <p class="riwayat-tanggal">{formatTanggal(item.tanggal)}</p>
                                    </div>
                                </div>
                                <div class="riwayat-right">
                                    <span class="riwayat-waktu {item.status.includes('Terlambat') ? 'text-warning' : 'text-success'}">{item.waktu_masuk || '--:--'}</span>
                                    <span class="riwayat-wib">WIB</span>
                                </div>
                            </div>
                        {/each}
                    {:else}
                        <div class="empty-state">
                            <p>Belum ada riwayat absensi.</p>
                        </div>
                    {/if}
                </div>

            </div>
        </div>

        <BottomNav activeMenu="riwayat" />
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
        overflow: hidden;
    }

    @media (min-width: 768px) {
        .mobile-app-wrapper { padding: 3rem 1.5rem; }
        .app-container {
            max-width: 900px;
            min-height: 0;
            max-height: 90vh;
            border-radius: 1.5rem;
            border: 1px solid #e5e7eb;
            box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1);
        }
    }

    .scrollable-content {
        flex-grow: 1;
        overflow-y: auto;
        -webkit-overflow-scrolling: touch;
        padding-bottom: 7rem;
        padding-top: 2rem;
    }

    .content-padding { padding: 1.5rem; }
    @media (min-width: 768px) { .content-padding { padding: 2rem; } }

    .page-header { margin-bottom: 2rem; padding-left: 0.25rem; }
    .page-title {
        font-size: 1.875rem;
        font-weight: 900;
        color: #1f2937;
        letter-spacing: 0.025em;
        margin: 0;
    }
    .page-subtitle {
        font-size: 0.875rem;
        font-weight: 600;
        color: #6b7280;
        margin-top: 0.25rem;
        margin-bottom: 0;
    }

    .riwayat-list { display: flex; flex-direction: column; gap: 1rem; }

    .riwayat-card {
        background-color: white;
        padding: 1rem;
        border-radius: 1rem;
        border: 1px solid #f3f4f6;
        box-shadow: 0 1px 2px rgba(0,0,0,0.05);
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    .riwayat-left { display: flex; align-items: center; gap: 1rem; }

    .riwayat-icon {
        width: 3rem;
        height: 3rem;
        border-radius: 0.75rem;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 1.25rem;
        flex-shrink: 0;
    }
    .icon-success { background-color: #ecfdf5; color: #10b981; }
    .icon-warning { background-color: #fffbeb; color: #f59e0b; }

    .riwayat-status {
        font-weight: 700;
        color: #1f2937;
        margin: 0;
        font-size: 0.9rem;
    }
    .riwayat-tanggal {
        font-size: 0.7rem;
        color: #6b7280;
        font-weight: 500;
        margin: 0.125rem 0 0;
    }

    .riwayat-right { text-align: right; }
    .riwayat-waktu {
        display: block;
        font-weight: 900;
        font-size: 1rem;
    }
    .text-success { color: #059669; }
    .text-warning { color: #d97706; }
    .riwayat-wib { font-size: 0.6rem; font-weight: 700; color: #9ca3af; }

    .loading-state, .error-state, .empty-state {
        text-align: center;
        padding: 3rem 1rem;
        background-color: white;
        border-radius: 1rem;
        color: #6b7280;
        font-weight: 500;
        font-size: 0.9rem;
    }
    .error-state { color: #dc2626; background-color: #fef2f2; }
</style>
