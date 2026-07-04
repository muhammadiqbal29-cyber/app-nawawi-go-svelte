<script lang="ts">
    import { navigate } from '../store/nav.svelte';
    import Icon from '../components/Icon.svelte';
    import { changePassword } from '../services/endpoints';
    import { toastStore } from '../store/toast.svelte';

    let oldPassword = $state('');
    let newPassword = $state('');
    let confirmPassword = $state('');
    let showOld = $state(false);
    let showNew = $state(false);
    let showConfirm = $state(false);
    let isLoading = $state(false);

    // Password strength indicator
    let strength = $derived(() => {
        if (newPassword.length === 0) return 0;
        if (newPassword.length < 6) return 1;
        if (newPassword.length >= 6 && newPassword.length < 8) return 2;
        if (newPassword.length >= 8 && /[A-Z]/.test(newPassword) && /[0-9]/.test(newPassword)) return 3;
        if (newPassword.length >= 10 && /[A-Z]/.test(newPassword) && /[0-9]/.test(newPassword) && /[^A-Za-z0-9]/.test(newPassword)) return 4;
        return 2;
    });

    const strengthColors = ['#e5e7eb', '#ef4444', '#f97316', '#10b981', '#059669'];

    function getStrengthColor(index: number) {
        const s = strength();
        return index < s ? strengthColors[s] : '#e5e7eb';
    }

    async function simpanPassword() {
        if (!oldPassword || !newPassword || !confirmPassword) {
            toastStore.show('Harap lengkapi semua kolom!', 'warning');
            return;
        }
        if (newPassword !== confirmPassword) {
            toastStore.show('Password baru dan konfirmasi tidak cocok!', 'error');
            return;
        }

        isLoading = true;
        try {
            const res = await changePassword({ old_password: oldPassword, new_password: newPassword });
            toastStore.show('Password berhasil diperbarui!', 'success');
            navigate('profil');
        } catch (e: any) {
            toastStore.show(e.message || 'Gagal mengubah password', 'error');
        } finally {
            isLoading = false;
        }
    }
</script>

<div class="mobile-app-wrapper">
    <div class="app-container">

        <!-- Header Rose -->
        <div class="page-header">
            <button class="back-btn" onclick={() => navigate('profil')}>
                <Icon name="arrow-left" /> Kembali
            </button>
            <div class="header-icon-wrap">
                <div class="header-icon">
                    <Icon name="lock" />
                </div>
            </div>
            <h1 class="header-title">Ganti Password</h1>
            <p class="header-subtitle">Pastikan password baru Anda kuat dan mudah diingat</p>
        </div>

        <!-- Form -->
        <div class="scrollable-content">
            <div class="form-container">
                <form onsubmit={(e) => { e.preventDefault(); simpanPassword(); }}>

                    <!-- Password Lama -->
                    <div class="input-wrap">
                        <input
                            type={showOld ? 'text' : 'password'}
                            class="input-field"
                            placeholder=" "
                            bind:value={oldPassword}
                        />
                        <label class="floating-label" for="old-password">Password Lama</label>
                        <button type="button" class="eye-btn" onclick={() => showOld = !showOld}>
                            <Icon name={showOld ? 'eye' : 'eye-slash'} class="eye-icon" />
                        </button>
                    </div>

                    <!-- Password Baru -->
                    <div class="input-wrap">
                        <input
                            type={showNew ? 'text' : 'password'}
                            class="input-field"
                            placeholder=" "
                            bind:value={newPassword}
                        />
                        <label class="floating-label" for="new-password">Password Baru</label>
                        <button type="button" class="eye-btn" onclick={() => showNew = !showNew}>
                            <Icon name={showNew ? 'eye' : 'eye-slash'} class="eye-icon" />
                        </button>
                    </div>

                    <!-- Strength Bars -->
                    <div class="strength-bars">
                        {#each [0,1,2,3] as i}
                            <div class="strength-bar" style="background-color: {getStrengthColor(i)};"></div>
                        {/each}
                    </div>

                    <!-- Konfirmasi Password -->
                    <div class="input-wrap">
                        <input
                            type={showConfirm ? 'text' : 'password'}
                            class="input-field"
                            placeholder=" "
                            bind:value={confirmPassword}
                        />
                        <label class="floating-label" for="confirm-password">Ulangi Password Baru</label>
                        <button type="button" class="eye-btn" onclick={() => showConfirm = !showConfirm}>
                            <Icon name={showConfirm ? 'eye' : 'eye-slash'} class="eye-icon" />
                        </button>
                    </div>

                    <div class="submit-wrap">
                        <button type="submit" class="btn-submit" disabled={isLoading}>
                            <Icon name="floppy-disk" /> {isLoading ? 'Menyimpan...' : 'Simpan Password'}
                        </button>
                    </div>

                </form>
            </div>
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

    /* Header */
    .page-header {
        background-color: #f43f5e;
        color: white;
        padding: 1.5rem;
        text-align: center;
        position: relative;
        z-index: 10;
        border-bottom: 4px solid #e11d48;
    }
    @media (min-width: 768px) { .page-header { padding: 2rem; } }

    .back-btn {
        position: absolute;
        top: 1.5rem;
        left: 1.5rem;
        color: rgba(255,255,255,0.8);
        background: transparent;
        border: none;
        cursor: pointer;
        font-size: 0.875rem;
        font-weight: 600;
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }
    .back-btn:hover { color: white; }

    .header-icon-wrap { display: flex; justify-content: center; margin-bottom: 1rem; margin-top: 1.5rem; }
    @media (min-width: 768px) { .header-icon-wrap { margin-top: 0; } }

    .header-icon {
        width: 4rem;
        height: 4rem;
        background: rgba(255,255,255,0.2);
        border-radius: 9999px;
        display: flex;
        align-items: center;
        justify-content: center;
        border: 2px solid rgba(255,255,255,0.5);
        font-size: 1.875rem;
    }

    .header-title { font-size: 1.5rem; font-weight: 900; letter-spacing: 0.025em; margin: 0; }
    .header-subtitle { font-size: 0.875rem; font-weight: 500; color: rgba(255,200,210,0.9); margin-top: 0.5rem; margin-bottom: 0; }

    /* Form */
    .scrollable-content {
        flex-grow: 1;
        overflow-y: auto;
        -webkit-overflow-scrolling: touch;
        padding-bottom: 2rem;
    }

    .form-container { padding: 1.5rem; }
    @media (min-width: 768px) { .form-container { padding: 2rem; } }

    .input-wrap {
        position: relative;
        background: white;
        border-radius: 1rem;
        border: 1px solid #e5e7eb;
        box-shadow: 0 1px 2px rgba(0,0,0,0.05);
        margin-bottom: 1.5rem;
        transition: border-color 0.2s, box-shadow 0.2s;
    }
    .input-wrap:focus-within {
        border-color: #10b981;
        box-shadow: 0 0 0 3px rgba(16,185,129,0.1);
    }

    .input-field {
        width: 100%;
        background: transparent;
        padding: 1rem;
        padding-right: 3rem;
        color: #1f2937;
        font-weight: 500;
        outline: none;
        border-radius: 1rem;
        border: none;
        font-family: inherit;
        font-size: 1rem;
        box-sizing: border-box;
    }

    .floating-label {
        position: absolute;
        left: 1rem;
        top: 50%;
        transform: translateY(-50%);
        color: #9ca3af;
        font-weight: 500;
        transition: all 0.2s;
        pointer-events: none;
        background: transparent;
        font-size: 1rem;
    }

    .input-field:focus ~ .floating-label,
    .input-field:not(:placeholder-shown) ~ .floating-label {
        top: 0;
        transform: translateY(-50%) scale(0.85);
        background-color: white;
        padding: 0 4px;
        color: #f43f5e;
    }

    .eye-btn {
        position: absolute;
        right: 1rem;
        top: 50%;
        transform: translateY(-50%);
        color: #9ca3af;
        background: transparent;
        border: none;
        cursor: pointer;
        transition: color 0.2s;
    }
    .eye-btn:hover { color: #10b981; }
    :global(.eye-icon) { width: 1rem; height: 1rem; }

    .strength-bars {
        display: flex;
        gap: 0.375rem;
        padding: 0 0.5rem;
        margin-top: -0.75rem;
        margin-bottom: 1.5rem;
    }
    .strength-bar {
        height: 0.375rem;
        flex: 1;
        border-radius: 9999px;
        transition: background-color 0.3s;
    }

    .submit-wrap { padding-top: 1rem; }
    .btn-submit {
        width: 100%;
        background-color: #f43f5e;
        color: white;
        font-weight: 700;
        font-size: 1.125rem;
        padding: 1rem;
        border-radius: 1rem;
        border: none;
        cursor: pointer;
        display: flex;
        justify-content: center;
        align-items: center;
        gap: 0.5rem;
        box-shadow: 0 10px 15px -3px rgba(244,63,94,0.3);
        transition: background-color 0.2s, transform 0.1s;
        font-family: inherit;
    }
    .btn-submit:hover { background-color: #e11d48; }
    .btn-submit:active { transform: scale(0.98); }
</style>
