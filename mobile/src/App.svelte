<script lang="ts">
    import { Route, router } from 'tinro';
    import { routerState } from "./store/nav.svelte";
    import Login from "./pages/Login.svelte";
    import MenuUtama from "./pages/MenuUtama.svelte";
    import Profil from "./pages/Profil.svelte";
    import GantiPassword from "./pages/GantiPassword.svelte";
    import DaftarSantriTahfiz from "./pages/DaftarSantriTahfiz.svelte";
    import InputTahfiz from "./pages/InputTahfiz.svelte";
    import InputTahsin from "./pages/InputTahsin.svelte";
    import Tahfiz from "./pages/Tahfiz.svelte";
    import KelolaHalaqah from "./pages/KelolaHalaqah.svelte";
    import Akademik from "./pages/Akademik.svelte";
    import Izin from "./pages/Izin.svelte";
    import RiwayatAbsen from "./pages/RiwayatAbsen.svelte";
    import MenuPerkembangan from "./pages/MenuPerkembangan.svelte";
    import DaftarKelasKeseharian from "./pages/DaftarKelasKeseharian.svelte";
    import InputKeseharian from "./pages/InputKeseharian.svelte";
    import DeviceOrientationOverlay from "./components/DeviceOrientationOverlay.svelte";
    import ToastContainer from "./components/ToastContainer.svelte";
    import ConfirmModal from "./components/ConfirmModal.svelte";
    import { onMount } from 'svelte';
    import { isAuthenticated } from './services/auth';

    // Setup tinro router for hash-based mobile friendly routing
    router.mode.hash();

    onMount(() => {
        // Initial splash/auth check
        setTimeout(() => {
            if (isAuthenticated()) {
                if (window.location.hash === '' || window.location.hash === '#/') {
                    router.goto('/menu_utama');
                }
            } else {
                router.goto('/login');
            }
        }, 1500);
    });
</script>

<ToastContainer />
<ConfirmModal />
<DeviceOrientationOverlay />


<main>
    <Route path="/">
        <div class="splash-screen">
            <img src="/logo.png" alt="Logo" class="splash-logo" onerror={(e) => e.currentTarget.style.display = 'none'} />
            <h1 class="splash-title">SIAKAD Nawawi</h1>
            <p class="splash-subtitle">Memuat aplikasi...</p>
        </div>
    </Route>

    <Route path="/login"><Login /></Route>
    <Route path="/menu_utama"><MenuUtama /></Route>
    <Route path="/profil"><Profil /></Route>
    <Route path="/ganti_password"><GantiPassword /></Route>
    <Route path="/input_tahfiz"><InputTahfiz /></Route>
    <Route path="/input_tahsin"><InputTahsin /></Route>
    <Route path="/tahfiz"><Tahfiz /></Route>
    <Route path="/daftar_santri_tahfiz"><DaftarSantriTahfiz /></Route>
    <Route path="/kelola_halaqah"><KelolaHalaqah /></Route>
    <Route path="/akademik"><Akademik /></Route>
    <Route path="/izin"><Izin /></Route>
    <Route path="/riwayat"><RiwayatAbsen /></Route>
    <Route path="/menu_perkembangan"><MenuPerkembangan /></Route>
    <Route path="/daftar_kelas_keseharian"><DaftarKelasKeseharian /></Route>
    <Route path="/input_keseharian"><InputKeseharian /></Route>
    
    <!-- Fallback route -->
    <Route fallback>
        <div class="flex items-center justify-center h-screen bg-gray-100 flex-col">
            <h1 class="text-2xl font-bold mb-2">404</h1>
            <p class="mb-4">Halaman tidak ditemukan</p>
            <button class="bg-blue-500 text-white px-4 py-2 rounded" onclick={() => router.goto('/menu_utama')}>Kembali ke Beranda</button>
        </div>
    </Route>
</main>

<style>
    :global(body) {
        margin: 0;
        font-family: 'Inter', sans-serif;
        background-color: #f3f4f6;
    }
    
    main {
        width: 100vw;
        min-height: 100vh;
        overflow-x: hidden;
    }

    .splash-screen {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 100vh;
        background: linear-gradient(135deg, #06b6d4 0%, #3b82f6 100%);
        color: white;
    }
    .splash-logo { width: 100px; height: 100px; margin-bottom: 2rem; border-radius: 20px; box-shadow: 0 10px 25px rgba(0,0,0,0.2); }
    .splash-title { font-size: 2rem; font-weight: 800; margin: 0 0 0.5rem 0; text-shadow: 0 2px 4px rgba(0,0,0,0.1); }
    .splash-subtitle { font-size: 1rem; opacity: 0.8; margin: 0; }
</style>
