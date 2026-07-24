<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { guruState } from '$lib/guruState.svelte.js';

  $effect(() => {
    if (guruState.activeSekolahId && guruState.aksesMap[guruState.activeSekolahId]) {
      const fiturGuru = guruState.aksesMap[guruState.activeSekolahId] || [];

      if (
        fiturGuru.includes('paket_alquran') ||
        fiturGuru.includes('input_tahfiz') ||
        fiturGuru.includes('input_tahfiz_v2') ||
        fiturGuru.includes('input_tasmi') ||
        fiturGuru.includes('input_tasmi_v2')
      ) {
        let isV2 = false;
        const sekolah = guruState.sekolahList.find(s => s.id == guruState.activeSekolahId);
        if (sekolah && sekolah.fitur) {
          try {
            const sf = JSON.parse(sekolah.fitur);
            isV2 = sf.includes('input_tahfiz_v2') || sf.includes('input_tasmi_v2');
          } catch {}
        }
        
        if (isV2) {
          goto('/guru/input-nilai/tahfiz-v2', { replaceState: true });
        } else {
          goto('/guru/input-nilai/tahfiz', { replaceState: true });
        }
      } else if (fiturGuru.includes('input_diniyah')) {
        goto('/guru/input-nilai/diniyah', { replaceState: true });
      } else if (fiturGuru.includes('rapor_keseharian')) {
        goto('/guru/input-nilai/keseharian', { replaceState: true });
      } else {
        goto('/guru', { replaceState: true });
      }
    }
  });
</script>

<div style="padding: 2rem; text-align: center; color: #64748b;">Mengalihkan halaman...</div>
