<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { guruState } from '$lib/guruState.svelte.js';

  $effect(() => {
    if (guruState.activeSekolahId && guruState.aksesMap[guruState.activeSekolahId]) {
      const fiturGuru = guruState.aksesMap[guruState.activeSekolahId] || [];
      
      if (fiturGuru.includes('input_tahfiz')) {
        goto('/guru/input-nilai/tahfiz', { replaceState: true });
      } else if (fiturGuru.includes('input_tahfiz_v2')) {
        goto('/guru/input-nilai/tahfiz-v2', { replaceState: true });
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

<div style="padding: 2rem; text-align: center; color: #64748b;">
  Mengalihkan halaman...
</div>
