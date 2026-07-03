<script lang="ts">
    import { onMount } from 'svelte';
    import { navigate } from '../store/nav.svelte';
    import Icon from '../components/Icon.svelte';

    let email = '';
    let password = '';
    let errorMessage = '';
    let isLoading = false;

    // Arahkan ke dashboard jika sudah login
    onMount(() => {
        const token = localStorage.getItem('token');
        if (token) {
            navigate('menu_utama');
        }
    });

    async function handleLogin() {
        isLoading = true;
        errorMessage = '';

        try {
            // Di environment mobile, kita perlu URL API absolut
            const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
            const response = await fetch(`${apiUrl}/api/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ email, password })
            });

            const text = await response.text();
            let data;
            try {
                data = JSON.parse(text);
            } catch (e) {
                errorMessage = "Gagal membaca respons dari server.";
                return;
            }

            if (!response.ok) {
                errorMessage = text || 'Login failed. Please check your credentials.';
                return;
            }

            if (data.token) {
                localStorage.setItem('token', data.token);
                localStorage.setItem('role', data.role);
                // Arahkan ke rute utama/dashboard setelah berhasil
                navigate('menu_utama'); 
            } else {
                errorMessage = 'Token tidak diterima.';
            }

        } catch (error) {
            errorMessage = 'Tidak dapat terhubung ke server.';
            console.error(error);
        } finally {
            isLoading = false;
        }
    }
</script>

<div class="app-container">
    <!-- LEFT/TOP PANEL: Brand Section -->
    <div class="brand-panel">
        <div class="welcome-text">Selamat Datang di</div>

        <div class="logo-container">
            <div class="logo-circle">
                <img src="/logo.webp" alt="Logo Yayasan Imam Nawawi" class="logo-image" />
            </div>
        </div>

        <div class="brand-desc">
            Sistem Informasi Yayasan Terpadu. Portal untuk manajemen akademik, keuangan, presensi, dan e-rapor.
        </div>

        <div class="brand-footer">
            <span class="footer-text">Powered by <br class="mobile-break" /><strong>Yayasan Imam Nawawi Palangka Raya</strong></span>
        </div>

        <!-- Decorative SVG Cloud - Posisi diatur via CSS background -->
        <div class="cloud-divider"></div>
    </div>

    <!-- RIGHT/BOTTOM PANEL: Form Section -->
    <div class="form-panel">
        <h2 class="form-title">Masuk ke akun Anda</h2>

        {#if errorMessage}
            <div class="error-alert">{errorMessage}</div>
        {/if}

        <form onsubmit={(e) => { e.preventDefault(); handleLogin(); }}>
            <div class="form-group">
                <label class="form-label" for="email">Alamat Email</label>
                <div class="input-wrapper">
                    <input type="email" id="email" class="form-input" bind:value={email} placeholder="Masukkan email Anda" required disabled={isLoading}>
                    {#if email}
                        <Icon name="check" class="check-icon" />
                    {/if}
                </div>
            </div>

            <div class="form-group">
                <label class="form-label" for="password">Kata Sandi</label>
                <div class="input-wrapper">
                    <input type="password" id="password" class="form-input" bind:value={password} placeholder="Masukkan kata sandi" required disabled={isLoading}>
                </div>
            </div>

            <div class="terms-group">
                <input type="checkbox" id="rememberMe" class="custom-checkbox">
                <label for="rememberMe" class="terms-text">
                    Ingat saya di perangkat ini
                </label>
            </div>

            <div class="button-group">
                <button type="submit" class="btn btn-primary" disabled={isLoading}>
                    {#if isLoading}
                        <Icon name="spinner" class="spinner-icon" /> Memproses...
                    {:else}
                        Sign In
                    {/if}
                </button>
            </div>
        </form>
    </div>
</div>

<style>
    /* --- Main Container --- */
    :global(body) {
        background-color: #e0e5ec;
        display: flex;
        justify-content: center;
        align-items: center;
        min-height: 100vh;
        padding: 0;
        margin: 0;
        font-family: 'Poppins', sans-serif;
        box-sizing: border-box;
    }

    * {
        box-sizing: border-box;
    }

    .app-container {
        background: #ffffff;
        border-radius: 20px;
        box-shadow: 0 15px 35px rgba(0, 0, 0, 0.1);
        overflow: hidden;
        display: flex;
        width: 100%;
        max-width: 900px;
        position: relative;
    }

    /* --- Left Panel (Brand Section) --- */
    .brand-panel {
        background: white;
        color: #333;
        padding: 40px;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: space-between;
        position: relative;
        z-index: 4; /* Diubah agar awan berada di atas form-panel */
        width: 45%;
    }

    .welcome-text {
        font-size: 1.2rem;
        font-weight: 400;
        margin-top: 20px;
        color: #666;
    }

    .logo-container {
        display: flex;
        flex-direction: column;
        align-items: center;
        margin: 40px 0;
    }

    .logo-circle {
        display: flex;
        justify-content: center;
        align-items: center;
        margin-bottom: 25px;
    }

    .logo-image {
        width: 240px;
        height: 240px;
        object-fit: contain;
    }


    .brand-desc {
        font-size: 0.85rem;
        text-align: center;
        line-height: 1.6;
        color: #64748b;
        max-width: 280px;
        margin-bottom: 40px;
    }

    .brand-footer {
        font-size: 0.6rem;
        text-transform: uppercase;
        letter-spacing: 2px;
        color: #94a3b8;
        display: flex;
        gap: 10px;
        align-items: center;
    }

    .footer-text {
        text-align: center;
        line-height: 1.6;
    }

    .mobile-break {
        display: none;
    }

    .brand-footer strong {
        color: #22c55e;
        font-weight: 600;
    }

    .cloud-divider {
        position: absolute;
        z-index: 2;
        pointer-events: none;
    }

    /* --- Right Panel (Form Section) --- */
    .form-panel {
        background: #22c55e;
        color: white;
        padding: 50px;
        width: 55%;
        display: flex;
        flex-direction: column;
        justify-content: center;
        position: relative;
        z-index: 3;
    }

    .form-title {
        color: white;
        font-size: 1.5rem;
        font-weight: 600;
        margin-bottom: 30px;
        text-align: left;
    }

    .error-alert {
        background-color: rgba(255, 255, 255, 0.2);
        color: #ffe4e6;
        padding: 10px 15px;
        border-radius: 8px;
        font-size: 0.85rem;
        margin-bottom: 20px;
        border-left: 4px solid #fecdd3;
        backdrop-filter: blur(5px);
    }

    .form-group {
        margin-bottom: 25px;
        position: relative;
    }

    .form-label {
        display: block;
        font-size: 0.75rem;
        font-weight: 500;
        color: rgba(255, 255, 255, 0.9);
        margin-bottom: 5px;
        letter-spacing: 0.5px;
    }

    .input-wrapper {
        position: relative;
    }

    .form-input {
        width: 100%;
        border: none;
        border-radius: 8px;
        padding: 12px 15px;
        padding-right: 40px;
        font-size: 0.9rem;
        color: #333;
        background: white;
        outline: none;
        box-shadow: 0 4px 10px rgba(0, 0, 0, 0.05);
        transition: box-shadow 0.3s;
        box-sizing: border-box;
    }

    .form-input:focus {
        box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
    }

    .form-input::placeholder {
        color: #94a3b8;
    }

    .form-input:disabled {
        opacity: 0.7;
        background: #f8fafc;
        cursor: not-allowed;
    }

    :global(.check-icon) {
        position: absolute;
        right: 15px;
        top: 50%;
        transform: translateY(-50%);
        width: 1rem;
        height: 1rem;
        color: #22c55e;
    }
    :global(.spinner-icon) {
        width: 1rem;
        height: 1rem;
        margin-right: 5px;
        animation: spin 0.8s linear infinite;
        transform-origin: center;
        display: inline-block;
        vertical-align: middle;
    }
    @keyframes spin {
        from { transform: rotate(0deg); }
        to { transform: rotate(360deg); }
    }

    /* --- Checkbox --- */
    .terms-group {
        display: flex;
        align-items: center;
        margin-top: 10px;
        margin-bottom: 30px;
    }

    .custom-checkbox {
        appearance: none;
        -webkit-appearance: none;
        width: 16px;
        height: 16px;
        background: rgba(255, 255, 255, 0.1);
        border: 1px solid rgba(255, 255, 255, 0.5);
        border-radius: 4px;
        margin-right: 10px;
        cursor: pointer;
        position: relative;
        display: flex;
        justify-content: center;
        align-items: center;
        transition: all 0.2s;
    }

    .custom-checkbox:checked {
        background: white;
        border-color: white;
    }

    .custom-checkbox::after {
        content: '✓';
        color: #22c55e;
        font-size: 11px;
        font-weight: 900;
        display: none;
    }

    .custom-checkbox:checked::after {
        display: block;
    }

    .terms-text {
        font-size: 0.8rem;
        color: rgba(255, 255, 255, 0.8);
    }

    /* --- Buttons --- */
    .button-group {
        display: flex;
        gap: 15px;
    }

    .btn {
        padding: 12px 35px;
        border-radius: 30px;
        font-size: 0.9rem;
        font-weight: 600;
        cursor: pointer;
        border: none;
        transition: all 0.3s;
        outline: none;
        width: 100%;
    }

    .btn:disabled {
        opacity: 0.7;
        cursor: not-allowed;
        transform: none !important;
        box-shadow: none !important;
    }

    .btn-primary {
        background: white;
        color: #22c55e;
        box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
    }

    .btn-primary:hover:not(:disabled) {
        box-shadow: 0 6px 20px rgba(0, 0, 0, 0.15);
        transform: translateY(-2px);
    }

    /* --- Desktop Layout Specifics --- */
    @media (min-width: 769px) {
        .app-container {
            flex-direction: row;
            width: 100vw;
            height: 100vh;
            max-width: none;
            border-radius: 0;
            box-shadow: none;
        }

        .brand-panel {
            width: 45%;
            justify-content: center;
        }

        .form-panel {
            width: 55%;
            padding: 50px 10%;
        }

        .logo-container {
            margin: 60px 0;
        }

        .cloud-divider {
            top: 0;
            bottom: 0;
            right: -80px;
            width: 150px;
            height: 100%;
            background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 500' preserveAspectRatio='none'%3E%3Cpath fill='%23ffffff' opacity='0.2' d='M50,0 C20,50 40,100 10,150 C-20,200 30,250 0,300 C-30,350 20,400 0,450 C-20,500 50,500 50,500 L0,500 L0,0 Z'/%3E%3Cpath fill='%23ffffff' opacity='0.5' d='M60,0 C40,60 70,120 30,180 C-10,240 50,300 10,360 C-30,420 40,480 60,500 L0,500 L0,0 Z'/%3E%3Cpath fill='%23ffffff' d='M80,0 C60,80 90,160 40,240 C-10,320 80,400 30,480 C0,500 80,500 80,500 L0,500 L0,0 Z'/%3E%3C/svg%3E");
            background-size: 100% 100%;
            background-repeat: no-repeat;
        }
    }

    /* --- Mobile & Tablet Layout Specifics --- */
    @media (max-width: 1024px) {
        .mobile-break {
            display: block;
        }
    }

    /* --- Mobile Layout Specifics --- */
    @media (max-width: 768px) {
        .app-container {
            flex-direction: column;
            max-width: none;
            width: 100vw;
            height: 100vh;
            border-radius: 0;
            box-shadow: none;
        }

        .brand-panel {
            width: 100%;
            padding: 20px 20px 60px 20px;
            border-radius: 0;
            flex: 0 0 auto;
        }

        .logo-image {
            width: 140px;
            height: 140px;
        }

        .logo-container {
            margin: 15px 0;
        }

        .welcome-text {
            font-size: 1rem;
            margin-top: 5px;
        }

        .brand-name {
            font-size: 1.8rem;
        }

        .brand-desc {
            display: none;
        }

        .brand-footer {
            position: absolute;
            bottom: 15px;
            left: 50%;
            transform: translateX(-50%);
            width: 100%;
            justify-content: center;
        }

        .cloud-divider {
            top: auto;
            bottom: -60px;
            left: 0;
            right: 0;
            width: 100%;
            height: 65px;
            background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 500 100' preserveAspectRatio='none'%3E%3Cpath fill='%23ffffff' opacity='0.2' d='M0,50 C50,20 100,40 150,10 C200,-20 250,30 300,0 C350,-30 400,20 450,0 C500,-20 500,50 500,50 L500,0 L0,0 Z'/%3E%3Cpath fill='%23ffffff' opacity='0.5' d='M0,60 C60,40 120,70 180,30 C240,-10 300,50 360,10 C420,-30 480,40 500,60 L500,0 L0,0 Z'/%3E%3Cpath fill='%23ffffff' d='M0,80 C80,60 160,90 240,40 C320,-10 400,80 480,30 C500,0 500,80 500,80 L500,0 L0,0 Z'/%3E%3C/svg%3E");
            background-size: 100% 100%;
            background-repeat: no-repeat;
        }

        .form-panel {
            width: 100%;
            padding: 60px 25px 20px 25px;
            border-radius: 0;
            flex: 1;
            justify-content: flex-start;
        }

        .form-title {
            text-align: center;
            font-size: 1.2rem;
            margin-bottom: 20px;
        }
        
        .form-group {
            margin-bottom: 15px;
        }

        .button-group {
            justify-content: center;
        }
    }
</style>
