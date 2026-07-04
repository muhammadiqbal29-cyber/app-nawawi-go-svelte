<script lang="ts">
    import { onMount } from 'svelte';
    import { navigate, routerState } from '../store/nav.svelte';
    import Icon from '../components/Icon.svelte';
    import { getSiswaAvailableTahfiz, getAnggotaHalaqah, createHalaqah, updateAnggotaHalaqah } from '../services/endpoints';
    import { toastStore } from '../store/toast.svelte';
    import { modalStore } from '../store/modal.svelte';

    let id_halaqah = $derived(routerState.navParams?.id_halaqah);

    let availableStudents = $state<any[]>([]);
    let currentMembers = $state<any[]>([]);
    
    // Set of selected student IDs
    let selectedStudentIds = $state<Set<number>>(new Set());
    
    let isLoading = $state(true);
    let isSaving = $state(false);

    // Grouping for UI
    let studentsByClass = $state<Record<string, any[]>>({});
    let expandedClasses = $state<Record<string, boolean>>({});

    onMount(async () => {
        try {
            isLoading = true;
            
            // Fetch available students
            const availableRes = await getSiswaAvailableTahfiz();
            availableStudents = availableRes || [];

            // If editing existing halaqah
            if (id_halaqah) {
                const memberRes = await getAnggotaHalaqah(id_halaqah);
                currentMembers = memberRes || [];
                
                // Add current members to selected set
                for (const m of currentMembers) {
                    selectedStudentIds.add(m.siswa?.id || m.siswa_id);
                }
            }

            // Group all students (available + current members) by Class
            const grouped: Record<string, any[]> = {};
            
            // Add current members first so they appear
            for (const m of currentMembers) {
                const s = m.siswa;
                if (!s) continue;
                // Since getAnggotaHalaqah might not return NamaKelas easily, we can just put them in 'Anggota Saat Ini'
                const rombel = s.rombel && s.rombel !== '__kosong__' ? s.rombel : '';
                const className = s.nama_kelas ? `${s.nama_kelas} ${rombel}`.trim() : 'Anggota Tersimpan';
                if (!grouped[className]) grouped[className] = [];
                grouped[className].push({ ...s, is_current: true });
            }

            // Add available students
            for (const s of availableStudents) {
                const className = `${s.nama_kelas} ${s.rombel === '__kosong__' ? '' : s.rombel}`.trim();
                if (!grouped[className]) grouped[className] = [];
                grouped[className].push({ ...s, is_current: false });
            }

            studentsByClass = grouped;
            
            // Expand all by default
            for (const key in grouped) {
                expandedClasses[key] = false;
            }
            // If editing, expand the 'Anggota Tersimpan' or first class
            const keys = Object.keys(grouped);
            if (keys.length > 0) expandedClasses[keys[0]] = true;

        } catch (e) {
            console.error(e);
            toastStore.show("Gagal memuat data siswa", "error");
        } finally {
            isLoading = false;
        }
    });

    function toggleStudent(id: number) {
        if (selectedStudentIds.has(id)) {
            selectedStudentIds.delete(id);
        } else {
            selectedStudentIds.add(id);
        }
        selectedStudentIds = new Set(selectedStudentIds); // trigger reactivity
    }

    function toggleClass(className: string) {
        expandedClasses[className] = !expandedClasses[className];
    }

    async function handleSave() {
        if (selectedStudentIds.size === 0) {
            const confirmed = await modalStore.confirm("Konfirmasi", "Anda tidak memilih satupun santri. Lanjutkan?");
            if (!confirmed) return;
        }

        isSaving = true;
        try {
            let targetHalaqahId = id_halaqah;

            // If creating new
            if (!targetHalaqahId) {
                const res = await createHalaqah({});
                targetHalaqahId = res.id;
            }

            // Update members
            await updateAnggotaHalaqah(targetHalaqahId, Array.from(selectedStudentIds));
            
            toastStore.show("Berhasil menyimpan anggota halaqah!", "success");
            navigate('daftar_santri_tahfiz');
        } catch (e) {
            console.error(e);
            toastStore.show("Gagal menyimpan halaqah.", "error");
        } finally {
            isSaving = false;
        }
    }
</script>

<div class="mobile-app-wrapper">
    <div class="app-container">
        <!-- Header -->
        <div class="page-header">
            <button class="back-btn" onclick={() => navigate('daftar_santri_tahfiz')}>
                <Icon name="arrow-left" /> Kembali
            </button>
            <h1 class="page-title">{id_halaqah ? 'Kelola Anggota' : 'Buat Halaqah'}</h1>
            <p class="page-subtitle">{selectedStudentIds.size} Santri Dipilih</p>
        </div>

        <div class="scrollable-content">
            <div class="content-padding">
                
                <div class="info-box mb-4">
                    <Icon name="circle-info" class="text-blue" />
                    <p>Centang siswa yang ingin dimasukkan ke halaqah Anda. Siswa yang sudah diklaim guru lain tidak akan muncul di sini.</p>
                </div>

                {#if isLoading}
                    <div class="loading-state">Memuat data siswa...</div>
                {:else if Object.keys(studentsByClass).length === 0}
                    <div class="empty-state">Tidak ada siswa yang tersedia.</div>
                {:else}
                    {#each Object.entries(studentsByClass) as [className, students]}
                        <div class="class-group">
                            <button class="class-header" onclick={() => toggleClass(className)}>
                                <div class="class-title">
                                    <Icon name="chalkboard-user" /> {className} 
                                    <span class="badge">{students.length}</span>
                                </div>
                                <Icon name={expandedClasses[className] ? 'chevron-up' : 'chevron-down'} />
                            </button>
                            
                            {#if expandedClasses[className]}
                                <div class="class-students">
                                    {#each students as s}
                                        <label class="student-row">
                                            <div class="checkbox-wrapper">
                                                <input type="checkbox" 
                                                    checked={selectedStudentIds.has(s.id)}
                                                    onchange={() => toggleStudent(s.id)} />
                                            </div>
                                            <div class="student-info">
                                                <span class="name">{s.nama_lengkap}</span>
                                                <span class="nis">
                                                    NIS: {s.nis}
                                                    {#if s.nama_kelas}
                                                        | {s.nama_kelas} {s.rombel && s.rombel !== '__kosong__' ? s.rombel : ''}
                                                    {/if}
                                                </span>
                                            </div>
                                        </label>
                                    {/each}
                                </div>
                            {/if}
                        </div>
                    {/each}
                {/if}

            </div>
        </div>

        <div class="bottom-action">
            <button class="btn-save" onclick={handleSave} disabled={isSaving || isLoading}>
                {#if isSaving}
                    Memproses...
                {:else}
                    <Icon name="floppy-disk" /> Simpan Halaqah
                {/if}
            </button>
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
    .back-btn { position: absolute; top: 1.5rem; left: 1.5rem; color: rgba(255,255,255,0.8); background: transparent; border: none; cursor: pointer; font-size: 0.875rem; font-weight: 600; display: flex; align-items: center; gap: 0.5rem; }
    .back-btn:hover { color: white; }
    .page-title { font-size: 1.5rem; font-weight: 900; margin: 2rem 0 0.25rem; }
    .page-subtitle { font-size: 0.875rem; font-weight: 600; color: rgba(255,255,255,0.9); margin: 0; }

    .scrollable-content { flex-grow: 1; overflow-y: auto; padding-bottom: 5rem; }
    .content-padding { padding: 1.5rem; }
    .mb-4 { margin-bottom: 1rem; }

    .info-box {
        background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 0.75rem; padding: 1rem;
        display: flex; gap: 0.75rem; align-items: flex-start;
    }
    .info-box p { margin: 0; font-size: 0.875rem; color: #1e3a8a; line-height: 1.4; }
    .text-blue { color: #3b82f6; font-size: 1.25rem; }

    .loading-state, .empty-state { text-align: center; color: #6b7280; font-weight: 500; margin-top: 2rem; }

    .class-group { margin-bottom: 1rem; background: white; border-radius: 1rem; border: 1px solid #e5e7eb; overflow: hidden; }
    .class-header {
        width: 100%; display: flex; justify-content: space-between; align-items: center; padding: 1rem;
        background: #f8fafc; border: none; cursor: pointer; font-size: 1rem; font-weight: 700; color: #334155;
        border-bottom: 1px solid #e5e7eb;
    }
    .class-title { display: flex; align-items: center; gap: 0.5rem; }
    .badge { background: #e2e8f0; color: #475569; font-size: 0.75rem; padding: 0.1rem 0.5rem; border-radius: 9999px; }
    
    .class-students { display: flex; flex-direction: column; }
    .student-row {
        display: flex; align-items: center; gap: 1rem; padding: 1rem; border-bottom: 1px solid #f1f5f9; cursor: pointer; transition: background 0.2s;
    }
    .student-row:last-child { border-bottom: none; }
    .student-row:hover { background: #f8fafc; }
    
    .checkbox-wrapper input[type="checkbox"] { width: 1.25rem; height: 1.25rem; accent-color: #0891b2; cursor: pointer; }
    
    .student-info { display: flex; flex-direction: column; }
    .student-info .name { font-weight: 600; color: #1f2937; font-size: 0.95rem; }
    .student-info .nis { font-size: 0.75rem; color: #6b7280; }

    .bottom-action {
        position: absolute; bottom: 0; left: 0; width: 100%; background: white; padding: 1rem 1.5rem;
        border-top: 1px solid #e5e7eb; box-shadow: 0 -4px 6px -1px rgba(0,0,0,0.05);
    }
    .btn-save {
        width: 100%; background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%); color: white; border: none; padding: 1rem; border-radius: 0.75rem;
        font-size: 1rem; font-weight: 700; cursor: pointer; display: flex; justify-content: center; align-items: center; gap: 0.5rem;
        box-shadow: 0 4px 6px -1px rgba(8, 145, 178, 0.3); transition: transform 0.2s;
    }
    .btn-save:active { transform: scale(0.98); }
    .btn-save:disabled { opacity: 0.7; cursor: not-allowed; }
</style>
