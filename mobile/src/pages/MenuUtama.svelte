<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import BottomNav from '../components/BottomNav.svelte';
    import Icon from '../components/Icon.svelte';
    import { getCurrentGuru, clockInGuru, clockOutGuru, getKonfigurasiPresensi, getActiveTahunAjaran } from '../services/endpoints';
    import { navigate } from '../store/nav.svelte';
    import { toastStore } from '../store/toast.svelte';
    import { modalStore } from '../store/modal.svelte';

    let timeString = $state('');
    let timer: number;
    let guruData = $state<any>(null);
    let activeTa = $state<any>(null);
    let isLoadingAbsen = $state(false);


    // Koordinat Sekolah (Akan di-fetch dari API)
    let schoolLat = $state(0);
    let schoolLng = $state(0);
    let radiusMeter = $state(50);

    // Koordinat User
    let latitude = $state(0);
    let longitude = $state(0);
    let isLocationReady = $state(false);
    


    function getUserLocation() {
        if (navigator.geolocation) {
            navigator.geolocation.getCurrentPosition(
                (position) => {
                    latitude = position.coords.latitude;
                    longitude = position.coords.longitude;
                    isLocationReady = true;
                },
                (error) => {
                    console.error("Error getting location:", error);
                    isLocationReady = true;
                },
                { enableHighAccuracy: true }
            );
        } else {
            isLocationReady = true;
        }
    }

    // Helper for Haversine distance
    function calculateDistance(lat1: number, lon1: number, lat2: number, lon2: number) {
        const R = 6371e3; // Earth radius in meters
        const p1 = lat1 * Math.PI/180;
        const p2 = lat2 * Math.PI/180;
        const dp = (lat2-lat1) * Math.PI/180;
        const dl = (lon2-lon1) * Math.PI/180;

        const a = Math.sin(dp/2) * Math.sin(dp/2) +
                  Math.cos(p1) * Math.cos(p2) *
                  Math.sin(dl/2) * Math.sin(dl/2);
        const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));

        return R * c;
    }

    let currentDistance = $derived(
        calculateDistance(latitude, longitude, schoolLat, schoolLng)
    );

    function updateClock() {
        const now = new Date();
        timeString = now.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }).replace('.', ':');
    }

    async function fetchData() {
        try {
            guruData = await getCurrentGuru();
            activeTa = await getActiveTahunAjaran();
            
            // Ambil konfigurasi presensi dari API
            const configRes = await getKonfigurasiPresensi();
            if (configRes && configRes.latitude && configRes.longitude) {
                schoolLat = configRes.latitude;
                schoolLng = configRes.longitude;
                radiusMeter = configRes.radius_meter || 50;
                if (!isLocationReady) {
                    latitude = configRes.latitude;
                    longitude = configRes.longitude;
                }
            }
        } catch (error) {
            console.error("Failed to fetch data:", error);
        }
    }

    async function handleClockIn() {
        if (schoolLat !== 0 && currentDistance > radiusMeter) {
            toastStore.show(`Gagal: Anda berada di luar radius sekolah (${currentDistance.toFixed(0)}m > ${radiusMeter}m)`, 'error');
            return;
        }

        const confirmed = await modalStore.confirm('Konfirmasi Absen', 'Apakah Anda yakin ingin melakukan Absen Masuk sekarang?');
        if (!confirmed) return;
        
        isLoadingAbsen = true;
        try {
            await clockInGuru({ latitude, longitude });
            toastStore.show('Absen masuk berhasil direkam.', 'success');
        } catch (e: any) {
            toastStore.show(e.message || 'Gagal merekam absen.', 'error');
        } finally {
            isLoadingAbsen = false;
        }
    }

    async function handleClockOut() {
        if (schoolLat !== 0 && currentDistance > radiusMeter) {
            toastStore.show(`Gagal: Anda berada di luar radius sekolah (${currentDistance.toFixed(0)}m > ${radiusMeter}m)`, 'error');
            return;
        }

        const confirmed = await modalStore.confirm('Konfirmasi Absen', 'Apakah Anda yakin ingin melakukan Absen Pulang sekarang?');
        if (!confirmed) return;
        
        isLoadingAbsen = true;
        try {
            await clockOutGuru({ latitude, longitude });
            toastStore.show('Absen pulang berhasil direkam.', 'success');
        } catch (e: any) {
            toastStore.show(e.message || 'Gagal merekam absen.', 'error');
        } finally {
            isLoadingAbsen = false;
        }
    }

    onMount(() => {
        updateClock();
        fetchData();
        getUserLocation();
        timer = setInterval(updateClock, 1000);
    });

    onDestroy(() => {
        clearInterval(timer);
    });

</script>

<svelte:head>

</svelte:head>

<div class="mobile-app-wrapper">
    <div class="app-container">
        <div class="scrollable-content">
            <div class="content-padding">
                
                <!-- TA Badge Centered -->
                {#if activeTa}
                    <div class="ta-badge-container">
                        <div class="ta-badge">
                            <Icon name="calendar-days" />
                            Tahun Ajaran {activeTa.nama_tahun}
                        </div>
                    </div>
                {/if}

                <!-- Live Clock & Date -->
                <div class="header-section">
                    <div>
                        <h1 class="live-clock">{timeString}</h1>
                        <p class="user-name">{guruData ? guruData.nama_lengkap : 'Memuat data...'}</p>
                    </div>
                    <div class="weather-icon">
                        <Icon name="cloud-sun" />
                    </div>
                </div>

                <!-- Teacher Attendance Widget -->
                <div class="widget-card" style="margin-bottom: 2rem;">
                    <div class="widget-header">
                        <h3 class="widget-title">Kehadiran Anda</h3>
                        <span class="badge-red">Belum Absen</span>
                    </div>
                    
                    <div class="btn-grid" style="display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.75rem;">
                        <button class="btn-absen btn-masuk" onclick={handleClockIn} disabled={isLoadingAbsen}>
                            <Icon name="fingerprint" class="btn-icon" />
                            <span class="btn-text">{isLoadingAbsen ? 'Memproses...' : 'Absen Masuk'}</span>
                        </button>
                        
                        <button class="btn-absen btn-masuk bg-orange" style="background-color: #f97316; box-shadow: 0 10px 15px -3px rgba(249, 115, 22, 0.3);" onclick={handleClockOut} disabled={isLoadingAbsen}>
                            <Icon name="house-person-leave" class="btn-icon" />
                            <span class="btn-text">{isLoadingAbsen ? 'Memproses...' : 'Absen Pulang'}</span>
                        </button>
                    </div>
                    {#if schoolLat !== 0 && isLocationReady}
                        <p class="location-text" style="font-size: 0.9rem; padding: 0.5rem; background: #f3f4f6; border-radius: 0.5rem; margin-top: 1.5rem; color: #4b5563;">
                            <Icon name="map-pin" /> Jarak dari titik absen: 
                            <strong style="color: {currentDistance <= radiusMeter ? '#10b981' : '#ef4444'}">
                                {currentDistance.toFixed(0)} meter
                            </strong>
                            <br/>
                            <span style="font-size: 0.75rem; color: #6b7280;">Radius maksimal: {radiusMeter} meter</span>
                        </p>
                    {:else if schoolLat === 0}
                        <p class="location-text" style="color: #ef4444; font-size: 0.85rem; margin-top: 1.5rem;"><Icon name="map-pin" /> Titik absensi sekolah belum dikonfigurasi admin.</p>
                    {:else}
                        <p class="location-text" style="color: #f59e0b; font-size: 0.85rem; margin-top: 1.5rem;"><Icon name="spinner" /> Mendapatkan lokasi Anda...</p>
                    {/if}
                </div>

                <!-- widgets removed -->

            </div>
        </div>

        <BottomNav activeMenu="home" />
    </div>
</div>

<style>
    .ta-badge-container {
        display: flex;
        justify-content: center;
        margin-bottom: 1.5rem;
        animation: slideDown 0.5s ease-out;
    }

    .ta-badge {
        display: inline-flex;
        align-items: center;
        gap: 0.5rem;
        background: linear-gradient(135deg, #0ea5e9 0%, #2563eb 100%);
        color: white;
        padding: 0.5rem 1.25rem;
        border-radius: 9999px;
        font-size: 0.875rem;
        font-weight: 700;
        letter-spacing: 0.025em;
        box-shadow: 0 10px 15px -3px rgba(37, 99, 235, 0.3), 0 4px 6px -2px rgba(37, 99, 235, 0.15);
        border: 2px solid rgba(255,255,255,0.2);
    }

    @keyframes slideDown {
        from { transform: translateY(-10px); opacity: 0; }
        to { transform: translateY(0); opacity: 1; }
    }

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
        .mobile-app-wrapper {
            padding: 3rem 1.5rem;
        }
        .app-container {
            max-width: 900px;
            min-height: 0;
            max-height: 90vh;
            border-radius: 1.5rem;
            border: 1px solid #e5e7eb;
            box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
        }
    }

    .scrollable-content {
        flex-grow: 1;
        overflow-y: auto;
        -webkit-overflow-scrolling: touch;
        padding-bottom: 7rem;
        padding-top: 2rem;
    }

    .content-padding {
        padding: 1.5rem;
    }

    /* Header */
    .header-section {
        display: flex;
        justify-content: space-between;
        align-items: flex-end;
        margin-bottom: 1.5rem;
        padding-left: 0.25rem;
    }

    .live-clock {
        font-size: 1.875rem;
        font-weight: 900;
        color: #1f2937;
        margin: 0;
    }

    .user-name {
        font-size: 0.875rem;
        font-weight: 600;
        color: #6b7280;
        margin-top: 0.25rem;
        margin-bottom: 0;
    }

    .weather-icon {
        width: 3rem;
        height: 3rem;
        background-color: white;
        border-radius: 9999px;
        box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
        border: 1px solid #f3f4f6;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #10b981;
        font-size: 1.25rem;
    }

    .widget-card {
        background-color: white;
        border-radius: 2rem;
        padding: 1.25rem;
        box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
        border: 1px solid #f3f4f6;
    }

    .widget-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 1rem;
    }

    .widget-title {
        font-weight: 700;
        color: #1f2937;
        margin: 0;
        font-size: 1rem;
    }

    .btn-menu {
        border-radius: 1rem;
        padding: 1rem;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 0.5rem;
        border: none;
        cursor: pointer;
        transition: all 0.2s;
        box-shadow: 0 10px 15px -3px rgba(16, 185, 129, 0.3);
    }
    .btn-menu:hover {
        background-color: #059669 !important;
    }
    .btn-menu:active {
        transform: scale(0.95);
    }

    :global(.btn-icon) {
        font-size: 1.875rem;
        margin-bottom: 0.25rem;
    }

    .btn-text {
        font-size: 0.75rem;
        font-weight: 700;
    }

    /* Tabs & Charts CSS */
    .tabs-widget-card {
        background-color: white;
        border-radius: 1.5rem;
        padding: 1.25rem;
        box-shadow: 0 4px 6px -1px rgba(0,0,0,0.05), 0 2px 4px -1px rgba(0,0,0,0.03);
        margin-top: 1.5rem;
        border: 1px solid #f3f4f6;
    }

    .tabs-header {
        display: flex;
        gap: 0.5rem;
        border-bottom: 2px solid #f3f4f6;
        margin-bottom: 1.25rem;
    }

    .tab-btn {
        flex: 1;
        background: transparent;
        border: none;
        padding: 0.75rem 0.25rem;
        font-size: 0.85rem;
        font-weight: 700;
        color: #6b7280;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 0.4rem;
        border-bottom: 3px solid transparent;
        transition: all 0.2s;
    }

    .tab-btn.active {
        color: #10b981;
        border-bottom-color: #10b981;
    }

    .chart-card {
        display: flex;
        flex-direction: column;
    }

    .chart-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 1rem;
    }

    .chart-title {
        font-size: 0.95rem;
        font-weight: 800;
        color: #1f2937;
        margin: 0;
    }

    .chart-select {
        font-size: 0.75rem;
        font-weight: 600;
        color: #4b5563;
        border: 1px solid #e5e7eb;
        background: #f9fafb;
        padding: 0.25rem 0.5rem;
        border-radius: 0.5rem;
        outline: none;
    }

    .chart-container {
        height: 200px;
        display: flex;
        align-items: flex-end;
        gap: 0.75rem;
        padding: 1.5rem 0.5rem 0.5rem;
        border-bottom: 2px solid #f3f4f6;
        position: relative;
        margin-bottom: 1rem;
        background: none;
        border-radius: 0;
        width: 100%;
        overflow: visible;
        box-sizing: border-box;
    }

    .bar-group {
        display: flex;
        flex-direction: column;
        align-items: center;
        flex: 1;
        height: 100%;
        justify-content: flex-end;
        gap: 0.5rem;
    }

    .bar-track {
        width: 100%;
        max-width: 32px;
        height: 100%;
        background-color: #f3f4f6;
        border-radius: 6px 6px 0 0;
        position: relative;
        display: flex;
        align-items: flex-end;
    }

    .bar-fill {
        width: 100%;
        border-radius: 4px 4px 0 0;
        transition: height 0.5s ease;
        position: relative;
    }

    .bar-value {
        position: absolute;
        top: -1.25rem;
        left: 50%;
        transform: translateX(-50%);
        font-size: 0.7rem;
        font-weight: 700;
        color: #1f2937;
    }

    .bar-label {
        font-size: 0.65rem;
        font-weight: 700;
        color: #9ca3af;
        text-align: center;
        white-space: nowrap;
    }

    /* Grouped bars for Tahfiz */
    .grouped-bars {
        flex-direction: row;
        align-items: flex-end;
        justify-content: center;
        gap: 0.15rem;
        position: relative;
        padding-bottom: 1.25rem;
    }

    .grouped-bars .bar-label {
        position: absolute;
        bottom: 0;
        left: 50%;
        transform: translateX(-50%);
    }

    .grouped-track {
        max-width: 16px;
        height: calc(100% - 1.25rem);
    }

    .cyan { background-color: #06b6d4; }
    .orange { background-color: #f97316; }

    .chart-legend {
        display: flex;
        justify-content: center;
        flex-wrap: wrap;
        gap: 0.75rem;
        font-size: 0.7rem;
        font-weight: 600;
        color: #6b7280;
    }

    .legend-item {
        display: flex;
        align-items: center;
        gap: 0.25rem;
    }

    .legend-color {
        width: 0.625rem;
        height: 0.625rem;
        border-radius: 0.25rem;
        display: inline-block;
    }

    /* Restored Styles */
    .badge-red {
        background-color: #fef2f2;
        color: #dc2626;
        font-size: 0.625rem;
        font-weight: 700;
        padding: 0.25rem 0.5rem;
        border-radius: 0.375rem;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .btn-absen {
        border-radius: 1rem;
        padding: 1rem;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 0.5rem;
        border: none;
        cursor: pointer;
        transition: all 0.2s;
    }
    .btn-absen:active:not(:disabled) { transform: scale(0.95); }
    .btn-masuk { background-color: #10b981; color: white; box-shadow: 0 10px 15px -3px rgba(16, 185, 129, 0.3); }
    .btn-masuk:hover { background-color: #059669; }

    .location-text {
        text-align: center;
        font-size: 0.75rem;
        color: #9ca3af;
        font-weight: 500;
        margin-top: 1rem;
        margin-bottom: 0;
    }

    .map-widget-card {
        background-color: white;
        border-radius: 2rem;
        padding: 1.25rem;
        box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
        border: 1px solid #f3f4f6;
    }
    .map-widget-header {
        display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.75rem;
    }
    .map-widget-header .widget-title {
        display: flex; align-items: center; gap: 0.4rem; font-size: 1rem; margin: 0; font-weight: 700; color: #1f2937;
    }
    .map-coord-badge {
        font-size: 0.65rem; font-weight: 600; color: #6b7280; background: #f3f4f6; padding: 0.2rem 0.6rem; border-radius: 999px;
    }
    .map-container {
        height: 220px; width: 100%; border-radius: 1.25rem; border: 1px solid #e5e7eb; overflow: hidden; z-index: 1;
    }
    @media (min-width: 768px) {
        .map-container { height: 280px; }
        .map-widget-card { padding: 1.5rem; }
    }
</style>
