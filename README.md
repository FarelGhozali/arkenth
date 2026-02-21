# 🕵️‍♂️ Enterprise Web QA Automation CLI

Sebuah _framework Command Line Interface_ (CLI) pengujian otomatis (_Automated QA Testing_) level _Enterprise_ yang ditulis menggunakan **Go**, **Cobra**, dan **Playwright-Go**.

Alat ini dirancang khusus untuk para _Quality Assurance Engineer_, _Bug Bounty Hunters_, dan _Security Researchers_ yang membutuhkan pemindai web yang sangat masif, modular, dan cepat. Berbeda dari _web crawler_ biasa, CLI ini tidak cuma membuka halaman web, namun ia secara aktif "menyerang" (_fuzzing_), menyadap seluruh aliran internet (_network interception_), dan bahkan melakukan perbandingan desain piksel UI!

---

## 🌟 Fitur Utama (Core Features)

### 1. 🌐 Deep Network Interception (Sistem Sadap Jaringan)

Alat ini menyusup ke level terendah dari _browser_ (_headless_) dan menyadap seluruh lalu-lintas jaringan (_Network Traffic_):

- Merekam setiap API _Call_ atau pemuatan aset yang mendapatkan respon **HTTP 4xx (Client Error)** dan **5xx (Server Error)**.
- Mencatat **Payload POST** yang memicu _error_ tersebut, sehingga memudahkan _developer_ untuk melakukan _debug_.
- Menyimpannya dalam bentuk raw `network_anomalies.json` secara mendetail (memuat Method, Endpoint, Status HTTP, dan Pesan _Error_).

### 2. ⚡ Fast-Mode (Optimasi Kecepatan)

Dilengkapi dengan `page.Route()` interception, jika Anda menggunakan bendera `--fast-mode`, CLI akan secara paksa memblokir aset yang memperlambat laju pengujian:

- Pemblokiran _Tracking scripts_ (Google Analytics, Meta Pixel).
- Pemblokiran _Media_ kelas berat (.jpg, .png. .mp4, tipe _font_).
- Meningkatkan kecepatan _scanning_ hingga 3x lipat pada mode non-visual.

### 3. 💥 Advanced Active Fuzzing & Rage Clicks

CLI ini tidak hanya sekadar melihat halaman web. Pada mode _scan_ normal, ia akan masuk ke fase "Penyerangan":

- **Form Injection:** Secara ajaib mendeteksi setiap `<form>`, `<input>`, dan `<textarea>` yang terlihat. CLI akan menyuntikkan _payload_ berbahaya (seperti karakter _Boundary/Max-Integer_, emoji, string yang sangat panjang, XSS simpel `<script>alert</script>`, dan SQL Injection dasar).
- **Rage Clicks:** Setelah form diisi, bot akan mengekstrak semua tombol (`.btn`, `<button>`) di layar dan mengekliknya secara paksa.
- **Crash Detection:** Ia memonitor respon DOM (_Document Object Model_). Jika aplikasi _frontend_ Anda (_React/Vue_) hancur dan memunculkan _Uncaught Exception Error_ karena form yang tidak divalidasi, bot akan langsung mengambil _screenshot_ warna-warni sebagai barang bukti (`crash_*.png`).

### 4. 👁️ Pixel-Perfect Visual Regression

Selain keamanan, CLI ini juga menjaga stabilitas desain Kosmetik (UI/UX) Anda melalui dua fase perintah:

- **Baseline:** Berkeliling web Anda HANYA untuk mengambil _screenshot_ resolusi tinggi (_Full-page_) di kondisi web saat ini sedang sempurna.
- **Compare:** Seminggu kemudian, saat Anda merilis kode CSS baru, bot akan memotret seluruh web Anda lagi dan memotong gambar tersebut lapis-demi-lapis pikselnya.
- Jika ada desain yang melenceng walau hanya 1 piksel warnanya, bot akan membuat **Gambar Rontgen (Visual Diff)** di mana area yang bergeser akan disiram warna Merah Terang!

### 5. 📱 Mobile Emulation & Auth State Injection

- Uji konsistensi web _responsive_ Anda dengan meniru gawai asli. (Misalnya: `--mobile-emulation="iPhone 13"`).
- Jika ada bagian [*Dashboard* admin] yang disembunyikan di balik layar Login, CLI dapat "mencuri" _cookies/storageState_ JSON Anda dan melakukan injeksi otomatis (`--auth-json="auth.json"`) sehingga bot bisa berpatroli seolah-olah ia adalah sang Admin.

---

## ⚙️ Petunjuk Pemasangan (Instalasi)

### Prasyarat:

- **Go 1.18+** terpasang di sistem operasi Anda.
- **Node.js** (Opsional, terkadang dipakai oleh skrip instalasi internal _browser_).

### Langkah-langkah:

1. **Unduh Repositori**:

   ```bash
   git clone <repo-url>
   cd web-qa-automation
   ```

2. **Perbarui Dependencies (Modul Go)**:

   ```bash
   go mod tidy
   ```

3. **Pasang Pustaka Browser Playwright**:
   _Framework_ ini butuh meminjam _browser chrome/playwright_ untuk menjelajah. Jalankan skrip ini sekali saja:

   ```bash
   go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps
   ```

4. **Kompilasi CLI Build (Ubah jadi File '.exe')**:
   ```bash
   go build -o web-qa
   ```

---

## 📖 Buku Panduan Penggunaan (Commands & Flags)

Aplikasi ini menggunakan teknologi Command-Line dari `spf13/cobra`. File binari (_executable_) yang telah di-_build_ menyajikan tiga perintah sakti:

### 1. `scan` - The Functional Security Audit

Ini adalah mode palu godam. Bot akan merayap ke dalam, merekam anomali jaringan, melakukan _Fuzzing_ paksa, merekam video (jika diminta), dan menangkap letak halaman yang rusak.

**Contoh Cepat (Fast Mode):**

```bash
./web-qa scan --target="https://example.com" --depth=2 --fast-mode
```

**Contoh Audit Dashboard (Menggunakan Login State):**

```bash
./web-qa scan --target="https://example.com/admin" --depth=3 --auth-json="./session.json"
```

### 2. `baseline` - The Pristine Snapshot

Perintah ini akan menyuruh bot berjalan-jalan dengan Sangat Sopan (Tanpa meretas/klik paksa). Ia akan menunggu sampai animasi halaman web selesai lalu diam-diam merekam _screenshot layout_ yang sempurna (_Baseline_).

```bash
./web-qa baseline --target="https://example.com" --depth=2
```

_Barang bukti foto murni akan disimpan ke folder: `proofs/baseline/`_

### 3. `compare` - The Visual Regression Engine

Kapanpun tim _developer_ Anda mencurigai perubahan kode `CSS/Frontend` merusak tampilan, jalankan kode ini! Ia akan mencocokkan kondisi website hari ini dengan foto _Baseline_ yang Anda simpan di mode sebelumnya.

```bash
./web-qa compare --target="https://example.com" --depth=2
```

_Hasil perhitungan pergeseran visualnya dicetak ke `visual_regression_report.md` beserta foto kemerah-merahannya di `proofs/diff/`_.

### 🚩 Kumpulan Flags Global Tersedia:

Anda bisa menyambungkan atribut ini di bagian belakang perintah apapun di atas:

- `--target` : **(WAJIB)** URL utama yang akan ditembus. (Contoh: `https://web-saya.com`).
- `--depth` : **(OPSIONAL)** Seberapa dalam bot harus mengklik link internal yang baru ditemukan. _Default: 1_. _Catatan: Skala eksponensial. Depth 3 ke atas butuh waktu berjam-jam untuk web besar._
- `--fast-mode`: **(OPSIONAL)** Mode miskin data (Mematikan pemuatan gambar beban berat agar pindai Fuzzing berjalan super ngebut).
- `--mobile-emulation`: **(OPSIONAL)** Meniru fisik gawai (User-Agent, Dimensi Layar). _Contoh: `--mobile-emulation="Pixel 5"`_
- `--record-video`: **(OPSIONAL)** WebM _Screen-record_. Seluruh sesi perjalanan peretas-an yang dilakukan bot akan dibungkus sebagai barang bukti video di folder `proofs/`.
- `--auth-json`: **(OPSIONAL)** Alamat path menuju file rekaman sesi cookies (_Playwright Storage State_).

---

## � Apa yang Dihasilkan Oleh Alat Ini? (Output / Report)

Setiap kali bot selesai mengemban misinya, Anda akan meraup sekumpulan emas Data Audit:

1. **`qa_audit_report.md`** : Buku Suci Laporan Utama. Terbagi atas ringkasan jumlah laman yang ditelusuri (![Crawl Map]), Log Kesalahan Jaringan Rinci (Daftar API yang ambruk dan mati), dan Bug-Bug Kritis. Khusus bug fungsionalitas, akan ada rincian tabel _Steps-To-Reproduce_ (Aksi tombol mana yang membuat aplikasi Anda rusak).
2. **`network_anomalies.json`** : Basis data mentah berformat JSON bagi Anda yang ingin membuat skrip pipa otomatis (_CI/CD_) untuk sekadar membaca rekap kesalahan HTTP/POST.
3. **`visual_regression_report.md`** : Laporan kembaran. Ini HANYA dicetak bila perintah `compare` dipanggil. Memamerkan daftar persentase (%) kerusakan layout per halaman lengkap dengan tabel foto rontgen visual-nya.
4. **Folder `./proofs/`** : Laci barang-bukti.
   - `proofs/crash_X.png`: Saksi bisu letak tombol pemicu kepanikan Java Script.
   - `proofs/view_X.png`: Sketsa muka web di setiap halaman.
   - `proofs/baseline/` & `proofs/diff/`: Repositori operasi tata-letak (_Visual testing_).

---

## 🛡️ Disclaimer (Peringatan Penting Hukum)

Aplikasi _Command-line_ ini dirakit dengan teknik manipulasi DOM tingkat tinggi dan algoritma _Form-Fuzzing_ agresif.

1. **Jangan PERNAH menjalankan instruksi `scan` di aplikasi Produksi/Server Utama tanpa sepengetahuan Atasan/Pemilik IT.** Active Fuzzing akan mengirim ratusan entri aneh (Emoji, karakter pembobol SQL) ke Form Kontak Terbuka, Database, maupun Registrasi Web. Ini bisa memicu alarm keamanan AWS/Cloudflare atau merusak (_corrupting_) Database asli.
2. Selalu arahkan atribut `--target` ke lingkungan _Sandbox_, _Localhost_, atau _Staging Server_ demi keselamatan Anda.
3. Modul Visual (`baseline` & `compare`) bersifat mode _Read-Only_ dan **100% aman** untuk dipergunakan pada situs yang tayang (Live/Produksi).

---

**Dibuat untuk merevolusi Dunia Quality Assurance secara Otomatis.**
