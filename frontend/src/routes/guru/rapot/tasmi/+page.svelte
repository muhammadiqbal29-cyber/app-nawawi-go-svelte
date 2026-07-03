<script>
  import Icon from '$lib/components/ui/Icon.svelte';
</script>

<div class="rapot-detail">
  <div class="detail-header">
    <div class="card-icon">
      <Icon name="bookOpen" size={32} />
    </div>
    <div>
      <h1>Rapot Ujian Tasmi'</h1>
      <p>Lihat rapot hasil ujian tasmi' siswa dan evaluasi bacaan.</p>
    </div>
  </div>

  <div class="detail-content">
    <p>Halaman khusus rapot untuk ujian tasmi' akan ditampilkan di sini.</p>
  </div>
</div>

<style>
  .rapot-detail {
    max-width: 900px;
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 1rem;
    padding: 2rem;
  }

  .detail-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .card-icon {
    width: 56px;
    height: 56px;
    border-radius: 18px;
    display: flex;
    justify-content: center;
    align-items: center;
    background: rgba(59, 130, 246, 0.12);
    color: #1d4ed8;
  }

  .detail-header h1 {
    margin: 0;
    font-size: 1.75rem;
    color: #0f172a;
  }

  .detail-header p {
    margin: 0.5rem 0 0;
    color: #475569;
  }

  .detail-content p {
    color: #334155;
    line-height: 1.75;
  }
</style>
