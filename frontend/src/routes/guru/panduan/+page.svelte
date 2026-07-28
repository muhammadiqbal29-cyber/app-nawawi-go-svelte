<script>
  import Icon from '$lib/components/ui/Icon.svelte';
</script>

<div class="panduan-container">
  <div class="page-header">
    <h1 class="page-title">Panduan Penggunaan</h1>
    <p class="page-subtitle">Informasi seputar cara kerja sistem dan fitur offline</p>
  </div>

  <div class="content-section">
    <div class="card info-card">
      <div class="card-icon blue">
        <Icon name="wifiOff" size={24} />
      </div>
      <div class="card-content">
        <h2>Sistem Anti-Mati Lampu (Offline-First)</h2>
        <p>Aplikasi ini dilengkapi dengan teknologi khusus yang memungkinkan Anda tetap bisa menginput nilai secara normal meskipun koneksi internet Anda tiba-tiba terputus atau sinyal jelek.</p>
        <ul class="feature-list">
          <li><strong>Background Auto-Sync:</strong> Saat Anda login pertama kali saat online, sistem akan secara cerdas mengunduh seluruh data murid di latar belakang agar siap digunakan kapan saja.</li>
          <li><strong>Brankas Lokal:</strong> Semua nilai yang Anda ketik saat offline akan masuk ke "Brankas Lokal" di HP/Laptop Anda. Data ini 100% aman dan tidak akan hilang meskipun halaman tertutup.</li>
          <li><strong>Kirim Otomatis:</strong> Saat koneksi internet kembali stabil, sistem akan otomatis mengirimkan isi brankas ke server pusat tanpa perlu Anda suruh.</li>
        </ul>
      </div>
    </div>

    <div class="card warning-card">
      <div class="card-icon yellow">
        <Icon name="alertTriangle" size={24} />
      </div>
      <div class="card-content">
        <h2>Peringatan: Hindari "Nilai Saling Menimpa"</h2>
        <p>Karena sistem ini bisa digunakan secara offline oleh banyak guru, mohon hindari skenario berikut agar nilai tidak saling menimpa:</p>
        <div class="warning-box">
          <strong>Aturan Emas:</strong> Pastikan hanya ada <strong>SATU GURU</strong> yang bertugas memasukkan nilai untuk satu kelas dan mata pelajaran yang sama.
        </div>
        <p class="text-sm">Jika Guru A dan Guru B sama-sama membuka profil murid yang sama (Aisyah) secara bersamaan tanpa internet, lalu keduanya mengubah nilainya, maka sistem akan menyimpan nilai milik guru yang pertama kali mendapatkan sinyal internet belakangan (Sistem <em>Last-Write-Wins</em>).</p>
      </div>
    </div>

    <div class="card tip-card">
      <div class="card-icon green">
        <Icon name="cloud" size={24} />
      </div>
      <div class="card-content">
        <h2>Memahami Indikator Awan</h2>
        <p>Di pojok kanan atas aplikasi, Anda akan melihat ikon awan. Berikut adalah artinya:</p>
        <div class="indicator-list">
          <div class="indicator-item">
            <span class="cloud-icon success"><Icon name="cloud" size={18} /></span>
            <div>
              <strong>Awan Hijau / Ceklis</strong>
              <p>Semua nilai sudah berhasil terkirim dan aman di server pusat.</p>
            </div>
          </div>
          <div class="indicator-item">
            <span class="cloud-icon warning"><Icon name="cloud" size={18} /></span>
            <div>
              <strong>Awan Kuning</strong>
              <p>Ada nilai di brankas HP Anda yang sedang antre dan belum terkirim ke server (biasanya karena offline/sinyal lemah). Biarkan saja, sistem akan terus mencoba mengirimkannya.</p>
            </div>
          </div>
          <div class="indicator-item">
            <span class="cloud-icon neutral"><Icon name="cloud" size={18} /></span>
            <div>
              <strong>Awan Abu-abu / Putih</strong>
              <p>Belum ada nilai yang diubah di halaman ini, atau Anda sedang tidak berada di halaman input nilai.</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .panduan-container {
    padding: 24px;
    max-width: 800px;
    margin: 0 auto;
    font-family: 'Inter', sans-serif;
  }

  .page-header {
    margin-bottom: 32px;
    text-align: center;
  }

  .page-title {
    font-size: 28px;
    font-weight: 700;
    color: #1e293b;
    margin: 0 0 8px 0;
  }

  .page-subtitle {
    color: #64748b;
    font-size: 15px;
    margin: 0;
  }

  .content-section {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .card {
    background: white;
    border-radius: 16px;
    padding: 24px;
    display: flex;
    gap: 20px;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -1px rgba(0, 0, 0, 0.03);
    border: 1px solid #f1f5f9;
  }

  .card-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .card-icon.blue { background: #eff6ff; color: #3b82f6; }
  .card-icon.yellow { background: #fefce8; color: #eab308; }
  .card-icon.green { background: #f0fdf4; color: #22c55e; }

  .card-content h2 {
    font-size: 18px;
    font-weight: 600;
    color: #0f172a;
    margin: 0 0 12px 0;
  }

  .card-content p {
    color: #475569;
    font-size: 14px;
    line-height: 1.6;
    margin: 0 0 16px 0;
  }

  .card-content p:last-child {
    margin-bottom: 0;
  }

  .feature-list {
    margin: 0;
    padding-left: 20px;
    color: #475569;
    font-size: 14px;
    line-height: 1.6;
  }

  .feature-list li {
    margin-bottom: 8px;
  }

  .feature-list li:last-child {
    margin-bottom: 0;
  }

  .warning-box {
    background: #fef2f2;
    border-left: 4px solid #ef4444;
    padding: 12px 16px;
    color: #991b1b;
    font-size: 14px;
    border-radius: 0 8px 8px 0;
    margin-bottom: 16px;
  }

  .text-sm {
    font-size: 13px !important;
    color: #64748b !important;
  }

  .indicator-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .indicator-item {
    display: flex;
    gap: 16px;
    align-items: flex-start;
  }

  .cloud-icon {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .cloud-icon.success { background: #22c55e; color: white; }
  .cloud-icon.warning { background: #eab308; color: white; }
  .cloud-icon.neutral { background: #cbd5e1; color: white; }

  .indicator-item strong {
    display: block;
    color: #1e293b;
    font-size: 14px;
    margin-bottom: 4px;
  }

  .indicator-item p {
    margin: 0;
    font-size: 13px;
    color: #64748b;
  }

  @media (max-width: 640px) {
    .card {
      flex-direction: column;
      padding: 20px;
      gap: 16px;
    }
    .panduan-container {
      padding: 16px;
    }
  }
</style>
