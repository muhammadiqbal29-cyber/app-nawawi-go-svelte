// 1. Add state for Generate Bulan Modal
  let showGenerateBulanModal = $state(false);
  let genBulanForm = $state({
    teacher_id: 0,
    bulan: parseInt(new Date().getMonth() + 1),
    tahun: new Date().getFullYear(),
    check_in_from: '06:30',
    check_in_to: '07:15',
    check_out_from: '15:30',
    check_out_to: '16:30',
    overwrite_existing: false,
    skip_sundays: true
  });

  function openGenerateBulanModal() {
    genBulanForm = {
      teacher_id: guruList.length > 0 ? guruList[0].id : 0,
      bulan: parseInt(filterMonth),
      tahun: parseInt(filterYear),
      check_in_from: '06:30',
      check_in_to: '07:15',
      check_out_from: '15:30',
      check_out_to: '16:30',
      overwrite_existing: false,
      skip_sundays: true
    };
    showGenerateBulanModal = true;
  }
