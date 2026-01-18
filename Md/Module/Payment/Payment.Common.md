# Module Payment.Common

module ini berfungsi untuk:

1. menampilkan list history pembelian
2. menampilkan detail history pembelian
3. melakukan cancel berdasarkan id
4. menyediakan endpoint untuk midtrans callback (webhook)

note:

- untuk menampilkan detail history, jika masih pending, coba cek status pada midtrans (ini fallback jika webhook tidak sampai)
