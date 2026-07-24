class GuruState {
  activeSekolahId = $state(null);
  activeSekolahName = $state('');
  sekolahList = $state([]);
  aksesMap = $state({});
  showJenjangModal = $state(false);

  // State untuk Tahun Ajaran & Semester
  activeTahunId = $state(null);
  activeSemesterId = $state(null);
  activeTahunName = $state('');
  tahunList = $state([]);
  showTahunModal = $state(false);
}

export const guruState = new GuruState();
