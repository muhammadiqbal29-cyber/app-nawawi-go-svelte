import mysql.connector

try:
    conn = mysql.connector.connect(
        host="localhost",
        user="siakad",
        password="siakad123",
        database="siakad"
    )
    cursor = conn.cursor()
    cursor.execute("SELECT nama_materi, juz FROM materi_tahfiz WHERE sekolah_id=4 LIMIT 10")
    for row in cursor.fetchall():
        print(row)
except Exception as e:
    print(e)
