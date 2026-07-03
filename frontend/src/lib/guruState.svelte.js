class GuruState {
  activeSekolahId = $state(null);
  activeSekolahName = $state('');
  sekolahList = $state([]);
  aksesMap = $state({});
  showJenjangModal = $state(false);
}

export const guruState = new GuruState();
