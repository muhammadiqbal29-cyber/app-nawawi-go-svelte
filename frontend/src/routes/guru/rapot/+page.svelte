<script>
  import Icon from '$lib/components/ui/Icon.svelte';

  const reportCards = [
    {
      title: 'Rapot Ujian Tahfiz',
      description: 'Lihat rapot hasil ujian tahfiz siswa dan unduh cetakannya.',
      href: '/guru/rapot/ujian-tahfiz',
      icon: 'award'
    },
    {
      title: "Rapot Ujian Tasmi'",
      description: "Lihat rapot hasil ujian tasmi' siswa dan tinjau perkembangan bacaannya.",
      href: '/guru/rapot/tasmi',
      icon: 'bookOpen'
    },
    {
      title: 'Rapot Ujian Kelulusan',
      description: 'Lihat rapot ujian kelulusan siswa serta status kelulusan dan rekomendasi.',
      href: '/guru/rapot/kelulusan',
      icon: 'checkSquare'
    },
    {
      title: 'Rapot Diniyah',
      description: 'Lihat laporan hasil pembelajaran mata pelajaran diniyyah siswa.',
      href: '/guru/rapot/diniyah',
      icon: 'book'
    },
    {
      title: 'Rapot Keseharian',
      description: 'Pantau catatan kedisiplinan dan nilai keseharian siswa (anomali).',
      href: '/guru/rapot/keseharian',
      icon: 'clipboardCheck'
    }
  ];
</script>

<div class="rapot-shell">
  <div class="section-header">
    <div>
      <p class="eyebrow">Rapot</p>
      <h1>Jenis Rapot</h1>
      <p class="subtitle">Pilih jenis rapot yang ingin Anda lihat untuk siswa.</p>
    </div>
  </div>

  <div class="report-grid">
    {#each reportCards as card (card.href)}
      <a class="report-card" href={card.href}>
        <div class="card-icon">
          <Icon name={card.icon} size={28} />
        </div>
        <div class="card-content">
          <h2>{card.title}</h2>
          <p>{card.description}</p>
        </div>
      </a>
    {/each}
  </div>
</div>

<style>
  .rapot-shell {
    max-width: 1100px;
  }

  .section-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1.5rem;
    margin-bottom: 2rem;
  }

  .eyebrow {
    margin: 0 0 0.5rem;
    color: #16a34a;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    font-size: 0.75rem;
  }

  .section-header h1 {
    margin: 0;
    font-size: 2rem;
    color: #0f172a;
  }

  .subtitle {
    margin: 0.5rem 0 0;
    color: #475569;
    line-height: 1.6;
    max-width: 680px;
  }

  .report-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1.5rem;
  }

  .report-card {
    display: block;
    background: #ffffff;
    border: 1px solid #e2e8f0;
    border-radius: 1rem;
    padding: 1.5rem;
    text-decoration: none;
    transition:
      transform 0.2s ease,
      box-shadow 0.2s ease,
      border-color 0.2s ease;
    color: inherit;
    min-height: 180px;
  }

  .report-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 18px 40px rgba(15, 23, 42, 0.08);
    border-color: #6366f1;
  }

  .card-icon {
    width: 48px;
    height: 48px;
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(99, 102, 241, 0.08);
    color: #4f46e5;
    margin-bottom: 1.25rem;
  }

  .card-content h2 {
    margin: 0;
    font-size: 1.15rem;
    color: #111827;
  }

  .card-content p {
    margin: 0.75rem 0 0;
    color: #64748b;
    line-height: 1.7;
    font-size: 0.95rem;
  }

  @media (max-width: 640px) {
    .section-header {
      flex-direction: column;
    }
  }
</style>
