package database

import (
	"siakad-nawawi-be/internal/models"
)

func seedBankCatatanTahfizV2() {
	var count int64
	DB.Model(&models.BankCatatanTahfizV2{}).Count(&count)
	if count > 0 {
		return
	}

	seeds := []models.BankCatatanTahfizV2{
		// Kombinasi Mumtaz
		{SekolahID: 1, Kelancaran: "Mumtaz", Tajwid: "Mumtaz", Makharij: "Mumtaz", TeksCatatan: "Masha Allah, baarakallahu fiik. Kelancaran, tajwid, dan makharijul huruf Ananda sangat baik (mumtaz). Pertahankan hafalan ini dan teruslah istiqamah dalam muraja'ah."},
		{SekolahID: 1, Kelancaran: "Mumtaz", Tajwid: "Mumtaz", Makharij: "Jayyid Jiddan", TeksCatatan: "Masha Allah, hafalan Ananda sangat lancar dan tajwidnya mumtaz. Hanya perlu sedikit penyempurnaan pada ketepatan makharijul huruf agar semakin sempurna."},
		{SekolahID: 1, Kelancaran: "Mumtaz", Tajwid: "Jayyid Jiddan", Makharij: "Mumtaz", TeksCatatan: "Alhamdulillah, kelancaran dan makharijul huruf Ananda mumtaz. Tingkatkan ketelitian pada hukum-hukum tajwid agar bacaan semakin tartil."},
		{SekolahID: 1, Kelancaran: "Mumtaz", Tajwid: "Jayyid Jiddan", Makharij: "Jayyid Jiddan", TeksCatatan: "Alhamdulillah, kelancaran hafalan Ananda sangat baik (mumtaz). Pengucapan tajwid dan makharij juga sudah jayyid jiddan. Tetap semangat menjaga hafalan Al-Qur'an."},
		{SekolahID: 1, Kelancaran: "Mumtaz", Tajwid: "Jayyid", Makharij: "Jayyid", TeksCatatan: "Masha Allah, hafalan Ananda sangat lancar. Namun, perhatikan kembali hukum tajwid dan makharij saat membaca cepat (hadr) agar bacaan tetap terjaga."},

		// Kombinasi Jayyid Jiddan
		{SekolahID: 1, Kelancaran: "Jayyid Jiddan", Tajwid: "Mumtaz", Makharij: "Mumtaz", TeksCatatan: "Masha Allah, bacaan Ananda sangat tartil dengan tajwid dan makharij yang mumtaz. Tingkatkan sedikit lagi kelancaran hafalannya dengan rutin muraja'ah."},
		{SekolahID: 1, Kelancaran: "Jayyid Jiddan", Tajwid: "Mumtaz", Makharij: "Jayyid Jiddan", TeksCatatan: "Baarakallahu fiik, pemahaman dan penerapan tajwid Ananda sangat luar biasa (mumtaz). Terus rutinkan muraja'ah agar kelancaran dan makharij semakin mutqin."},
		{SekolahID: 1, Kelancaran: "Jayyid Jiddan", Tajwid: "Jayyid Jiddan", Makharij: "Mumtaz", TeksCatatan: "Alhamdulillah, pengucapan makharijul huruf Ananda mumtaz. Kelancaran dan tajwid juga sudah sangat baik. Tetap semangat mengulang hafalan di rumah."},
		{SekolahID: 1, Kelancaran: "Jayyid Jiddan", Tajwid: "Jayyid Jiddan", Makharij: "Jayyid Jiddan", TeksCatatan: "Alhamdulillah, pencapaian hafalan, tajwid, dan makharij Ananda semuanya sudah sangat baik (jayyid jiddan). Teruslah bersemangat dan jangan lupa muraja'ah."},
		{SekolahID: 1, Kelancaran: "Jayyid Jiddan", Tajwid: "Jayyid", Makharij: "Jayyid", TeksCatatan: "Alhamdulillah, kelancaran hafalan Ananda jayyid jiddan. Ananda perlu lebih berhati-hati dan teliti dengan tajwid serta makharij agar kualitas bacaan selaras dengan kelancarannya."},

		// Kombinasi Jayyid
		{SekolahID: 1, Kelancaran: "Jayyid", Tajwid: "Jayyid Jiddan", Makharij: "Jayyid Jiddan", TeksCatatan: "Baarakallahu fiik, penerapan tajwid dan makharij Ananda sudah sangat baik. Perbanyak lagi frekuensi muraja'ah harian agar kelancaran hafalan juga ikut meningkat."},
		{SekolahID: 1, Kelancaran: "Jayyid", Tajwid: "Jayyid", Makharij: "Jayyid", TeksCatatan: "Alhamdulillah, hafalan Ananda secara umum baik (jayyid). Butuh tambahan porsi muraja'ah serta bimbingan tajwid dan makharij agar bacaan lebih sempurna."},

		// Kombinasi Maqbul
		{SekolahID: 1, Kelancaran: "Maqbul", Tajwid: "Semua Kriteria", Makharij: "Semua Kriteria", TeksCatatan: "Pencapaian hafalan Ananda cukup (maqbul). Ananda perlu menambah semangat dalam mengulang-ulang hafalan dan memperbaiki tahsinnya secara bertahap."},

		// Kombinasi Dha'if
		{SekolahID: 1, Kelancaran: "Dha'if", Tajwid: "Semua Kriteria", Makharij: "Semua Kriteria", TeksCatatan: "Ananda butuh bimbingan lebih intensif untuk melancarkan hafalan. Jangan mudah menyerah, perbanyak waktu mengulang hafalan bersama orang tua di rumah."},
		{SekolahID: 1, Kelancaran: "Semua Kriteria", Tajwid: "Dha'if", Makharij: "Semua Kriteria", TeksCatatan: "Ananda perlu lebih memperhatikan kaidah tajwid saat membaca (panjang pendek dan dengung). Biasakan membaca dengan perlahan dan tartil."},
		{SekolahID: 1, Kelancaran: "Semua Kriteria", Tajwid: "Semua Kriteria", Makharij: "Dha'if", TeksCatatan: "Pengucapan makharijul huruf Ananda masih perlu banyak perbaikan. Sering-seringlah berlatih talaqqi agar tidak tertukar dalam penyebutan huruf."},
	}

	for _, s := range seeds {
		DB.Create(&s)
	}
}

func seedBankCatatanDiniyah() {
	var count int64
	DB.Model(&models.BankCatatanDiniyah{}).Count(&count)
	if count > 0 {
		return
	}

	seeds := []models.BankCatatanDiniyah{
		{SekolahID: 1, Predikat: "Mumtaz", TeksCatatan: "Alhamdulillah nilai Diniyah Ananda rata-rata sangat baik (Mumtaz). Prestasi Ananda sangat menonjol di mapel [MAPEL_TERTINGGI], namun perlu sedikit muraja'ah lagi pada mapel [MAPEL_TERENDAH]."},
		{SekolahID: 1, Predikat: "Jayyid Jiddan", TeksCatatan: "Masha Allah, secara umum pemahaman Ananda terhadap materi-materi Diniyah sudah sangat baik (Jayyid Jiddan). Kemampuan di mapel [MAPEL_TERTINGGI] sangat memuaskan, perbanyak lagi pengulangan di mapel [MAPEL_TERENDAH]."},
		{SekolahID: 1, Predikat: "Jayyid", TeksCatatan: "Capaian Diniyah Ananda sudah baik (Jayyid). Ananda unggul dalam mapel [MAPEL_TERTINGGI], akan tetapi perlu meningkatkan semangat belajarnya untuk mapel [MAPEL_TERENDAH]."},
		{SekolahID: 1, Predikat: "Maqbul", TeksCatatan: "Hasil belajar Diniyah Ananda cukup baik (Maqbul). Ananda cukup menguasai mapel [MAPEL_TERTINGGI], namun harus jauh lebih giat lagi belajar pada mapel [MAPEL_TERENDAH]."},
		{SekolahID: 1, Predikat: "Dha'if", TeksCatatan: "Ananda perlu lebih giat dan sungguh-sungguh lagi dalam mempelajari ilmu Diniyah. Saat ini mapel [MAPEL_TERTINGGI] menjadi yang terbaik bagi Ananda, sedangkan mapel [MAPEL_TERENDAH] membutuhkan perhatian lebih ekstra."},
	}

	for _, s := range seeds {
		DB.Create(&s)
	}
}

func seedBankCatatanKeseharian() {
	var count int64
	DB.Model(&models.BankCatatanKeseharian{}).Count(&count)
	if count > 0 {
		return
	}

	seeds := []models.BankCatatanKeseharian{
		{SekolahID: 1, Predikat: "Sangat Baik", TeksCatatan: "Alhamdulillah, Ananda memiliki kedisiplinan dan akhlak yang sangat baik selama di pondok. Terus pertahankan prestasi dan kelakuan baiknya!"},
		{SekolahID: 1, Predikat: "Baik", TeksCatatan: "Ananda menunjukkan sikap yang baik dan kooperatif, meskipun sesekali masih perlu diingatkan. Tingkatkan terus kedisiplinannya!"},
		{SekolahID: 1, Predikat: "Cukup", TeksCatatan: "Ananda cukup baik, namun perlu lebih memperhatikan aturan pondok. Mohon kerja sama orang tua untuk memotivasi Ananda agar lebih tertib."},
		{SekolahID: 1, Predikat: "Kurang", TeksCatatan: "Ananda memerlukan perhatian dan bimbingan yang lebih intensif dari orang tua terkait kepatuhan terhadap tata tertib pondok pesantren."},
	}
	
	for _, s := range seeds {
		DB.Create(&s)
	}
}
