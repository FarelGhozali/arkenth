# 🕵️‍♂️ Enterprise Web QA Automation CLI

Sebuah _framework_ pengujian otomatis (_Automated QA Testing_) level _Enterprise_ yang ditulis menggunakan **Go**, **Cobra**, dan **Playwright-Go**, dilengkapi dengan **Web UI Dashboard** berbasis **Svelte + Vite** yang tersemat langsung ke dalam satu _executable binary_.

Alat ini dirancang khusus untuk para _Quality Assurance Engineer_, _Bug Bounty Hunters_, dan _Security Researchers_ yang membutuhkan pemindai web yang sangat masif, modular, dan cepat. Anda dapat menggunakannya melalui **CLI (Command Line)** maupun **Web Dashboard** yang modern dan intuitif — tanpa perlu mengetik _command_ sama sekali! Berbeda dari _web crawler_ biasa, alat ini tidak cuma membuka halaman web, namun ia secara aktif "menyerang" (_fuzzing_), menyadap seluruh aliran internet (_network interception_), dan bahkan melakukan perbandingan desain piksel UI!

---

## 🌟 Fitur Utama (Core Features)

### 1. 🌐 Deep Network Interception & Fingerprinting

Alat ini menyusup ke level terendah dari _browser_ (_headless_) dan menyadap seluruh lalu-lintas jaringan (_Network Traffic_):

- Merekam setiap API _Call_ atau pemuatan aset yang mendapatkan respon **HTTP 4xx (Client Error)** dan **5xx (Server Error)**.
- **Server/WAF Fingerprinting:** Saat mesin target ambruk (Error), bot membaca _HTTP Response Header_ secara mendalam untuk melacak teknologi pertahanan yang mereka gunakan (Contoh: menemukan pelacak `Cloudflare`, `AWS CloudFront`, atau mendeteksi _framework_ `Express` / `Spring Boot`).
- Mencatat **Payload POST** yang memicu _error_ tersebut, sehingga memudahkan _developer_ untuk melakukan _debug_.
- Menyimpannya dalam bentuk raw `network_anomalies.json` secara mendetail (memuat Intelijen Target, Method, Endpoint, Status HTTP, dan Pesan _Error_).

### 2. ⚡ Fast-Mode (Optimasi Kecepatan)

Dilengkapi dengan `page.Route()` interception, jika Anda menggunakan bendera `--fast-mode`, CLI akan secara paksa memblokir aset yang memperlambat laju pengujian:

- Pemblokiran _Tracking scripts_ (Google Analytics, Meta Pixel).
- Pemblokiran _Media_ kelas berat (.jpg, .png. .mp4, tipe _font_).
- Meningkatkan kecepatan _scanning_ hingga 3x lipat pada mode non-visual.

### 3. 💥 Advanced Security & Smart API Discovery

CLI ini tidak hanya sekadar melihat halaman web. Pada mode `scan` normal, ia akan masuk ke fase "Penetrasi":

- **Smart API Discovery:** Bot akan secara diam-diam mengunduh semua file `.js` _Frontend_ di website target Anda. Menggunakan Regex, ia akan mengekstrak direktori rute **Hidden API/Internal** (seperti `/api/v1/users`) yang tidak pernah dipublikasikan, lalu menyusup ke dalamnya.
- **Form Injection:** Secara ajaib mendeteksi setiap `<form>`, `<input>`, dan `<textarea>` yang terlihat. CLI akan menyuntikkan ratusan _payload_ berbahaya (seperti karakter _Boundary_, emoji, XSS `<script>alert</script>`, Deep SQLi `' OR 1=1 --`, NoSQLi, hingga _Command Injection_).
- **Rage Clicks:** Setelah form diisi, bot akan mengekstrak semua tombol (`.btn`, `<button>`) di layar dan mengekliknya secara paksa.
- **JWT Token Tampering:** _(Membutuhkan flag `--auth-json`)_. Ia memburu _Cookie Authorization (JWT)_ dari sesi login Anda. Jika ditemukan, bot akan melucuti kriptografi keamanannya secara sepihak (CVE-2015-9256) untuk menguji celah **Broken Access Control**.
- **Crash Detection:** Ia memonitor respon DOM dan HTTP. Jika aplikasi _frontend/backend_ hancur (Uncaught Exception Error / HTTP 500) bot akan langsung menangkap layar dan memasukkannya ke rekap **Critical Bugs**.

### 4. 👁️ Pixel-Perfect Visual Regression

Selain keamanan, CLI ini juga menjaga stabilitas desain Kosmetik (UI/UX) Anda melalui dua fase perintah:

- **Baseline:** Berkeliling web Anda HANYA untuk mengambil _screenshot_ resolusi tinggi (_Full-page_) di kondisi web saat ini sedang sempurna.
- **Compare:** Seminggu kemudian, saat Anda merilis kode CSS baru, bot akan memotret seluruh web Anda lagi dan memotong gambar tersebut lapis-demi-lapis pikselnya.
- Jika ada desain yang melenceng walau hanya 1 piksel warnanya, bot akan membuat **Gambar Rontgen (Visual Diff)** di mana area yang bergeser akan disiram warna Merah Terang!

### 5. 📱 Mobile Emulation & Auth State Injection

- Uji konsistensi web _responsive_ Anda dengan meniru gawai asli. (Misalnya: `--mobile-emulation="iPhone 13"`).
- Jika ada bagian [*Dashboard* admin] yang disembunyikan di balik layar Login, CLI dapat "mencuri" _cookies/storageState_ JSON Anda dan melakukan injeksi otomatis (`--auth-json="auth.json"`) sehingga bot bisa berpatroli seolah-olah ia adalah sang Admin.

### 6. ♿ Accessibility (WCAG a11y) Engine with Dynamic UI

Menginjeksi pustaka standar industri `axe-core` ke dalam otak Playwright untuk memindai seluruh halaman web dari cacat struktural (warna terlalu silau bagi penyandang disabilitas, _Screen Reader/ARIA_ tags yang tidak berfungsi, dsb).
**Peningkatan Level Lanjut:** Bot ini secara mandiri akan membuka seluruh _Dropdown Menu, Pop-up Modal, dan Dialog/Accordion_ yang disembunyikan CSS _sebelum_ memindai halamannya, sehingga cacat _UI tersembunyi_ tidak bisa lolos dari pemindaian.

### 7. 🚀 Concurrent & Stateful Load Testing

Dirancang ulang menggunakan _Goroutines_ bahasa Go tingkat rendah (tanpa GUI Browser) untuk menembakkan ribuan permintaan HTTP secara serentak ke mesin _Server/API_.
**Dukungan Stateful POST/Bypass Cache:** Bot load-tester ini mampu menembakkan lalu lintas HTTP `POST/PUT` seraya merubah-rubah _Payload JSON_ secara dinamis di setiap tembakan. Contoh: `{"email": "user_{{RANDOM}}@test.com"}` guna memaksa lalu lintas menembus blokade _Cache CDN Cloudflare_ dan langsung meledakkan _Database Server_ Anda sesungguhnya.

### 8. 🖥️ Web UI Dashboard (Antarmuka Web Internal)

Sebagai sebuah *tool QA Automation*, kami memahami bahwa pengujian dan pemantauan seringkali lebih efisien bila dieksekusi melalui antarmuka grafis (GUI). Oleh karena itu, *executable binary* hasil kompilasi *tool* ini telah dilengkapi dengan **Web Dashboard** modern bawaan.

Fitur ini memungkinkan Anda (maupun tim QA non-teknis internal Anda) untuk mengoperasikan seluruh instruksi canggih (Scan, Regresi Visual, Audit, dsb) secara interaktif melalui *browser*, tanpa perlu menghafal _flag_ panjang di CLI.

**Kelebihan Menggunakan Web Dashboard:**
- **Satu Aplikasi Utuh (Single Executive):** Anda TIDAK PERLU menginstal Node.js, Web Server Apache, Nginx, atau dependensi web rumit lainnya. Kami membungkus seluruh wujud *website frontend* (Svelte + Tailwind) langsung ke dalam satu buah file *executable* milik Golang kita. 
- **User-Friendly:** Tinggal isi "URL Target", geser *slider* ke dalaman scan, hidup-matikan opsi video melalui tombol, lalu klik "Mulai". Sangat praktis.
- **Log Visual Terpadu:** Alih-alih membaca teks hitam-putih bergerak cepat di Terminal, Dashboard menyajikan laporan status *real-time* yang enak dipandang (mendukung Tema Gelap/Night Mode).
- **Asinkron & Background Task:** Setelah Anda menekan tombol "Mulai Scan" di web, Anda bisa meninggalkan halaman itu. CLI akan mengeksekusi operasi berat tersebut di *background*.

### 9. 🪟 Desktop Application (Native Wails)

Bagi Anda yang lebih suka aplikasi *standalone* layaknya program profesional, kami juga menyediakan jembatan ke format **Aplikasi Desktop Desktop** menggunakan *framework* Wails. 
Fitur Desktop ini memberikan pengalaman UI/UX yang jauh lebih kuat dibandingkan versi *Web Dashboard* berkat kemampuannya untuk meng-intervensi dan merender seluruh lalu-lintas **Terminal Go Log** langsung ke lapisan antarmuka grafis secara *Real-Time*.
- **Tanpa Localhost:** Aplikasi berjalan mandiri tanpa harus memaksa *user* membuka *browser* atau bingung mengatur *port* jaringan lokal.
- **Transparansi Pekerjaan (Live Logs):** Pergerakan kursor dan log bot penyerang dikirim sedetik itu juga (*live stream*) ke layar *dashboard* Desktop Anda!

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

4. **Build Frontend (Web UI Dashboard)**:
   ```bash
   cd frontend
   npm install
   npm run build
   cd ..
   ```

5. **Kompilasi CLI Build (Ubah jadi File '.exe')**:

   Proses ini akan otomatis menyertakan (_embed_) aset Web UI yang sudah di-_build_ pada langkah sebelumnya ke dalam file _binary_. Anda bebas menamai hasil output file eksekusinya (misalnya `qa-bot`):
   ```bash
   go build -o qa-bot
   ```

6. **Kompilasi Menjadi Aplikasi Desktop (Opsional - Mode Wails):**

   Jika sistem Anda telah mendukung prasyarat *library WebView/GTK* Wails, Anda bisa mengubah ekosistem CGO di atas menjadi aplikasi `.exe` atau `.app` *desktop window* modern.
   ```bash
   # 1. Pastikan Wails CLI terinstall di Go Anda
   go install github.com/wailsapp/wails/v2/cmd/wails@latest

   # 2. Pindah ke direktori desktop wrapper
   cd desktop

   # 3. Rakit Aplikasinya
   wails build
   ```
   *(Catatan Linux: Butuh dependensi tambahan `sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev` sebelum melakukan kompilasi GUI di atas)*

---

## 🐳 Menjalankan dengan Docker

### Build image

```bash
docker build -t web-qa-automation:latest .
```

### Jalankan command CLI dari container

```bash
docker run --rm -it \
  -v "$(pwd)/proofs:/app/proofs" \
  web-qa-automation:latest scan --target="https://example.com" --depth=1
```

> Gunakan volume mount jika Anda ingin menyimpan output report/screenshot ke host lokal.

---

## 🚀 CI/CD Docker (GitHub Actions)

Workflow Docker ada di `.github/workflows/docker.yml` dengan perilaku:

- **Pull Request**: build image Docker untuk validasi (tanpa push).
- **Push ke `main` atau tag `v*`**: build dan push image ke **GHCR**.
- Image dipublikasikan ke:
  - `ghcr.io/<owner>/<repo>:latest` (default branch)
  - `ghcr.io/<owner>/<repo>:<branch|tag|sha>`

---

## 📖 Buku Panduan Penggunaan (Commands & Flags)

Aplikasi ini menggunakan teknologi Command-Line dari `spf13/cobra`. File binari (_executable_) yang telah di-_build_ menyajikan dua mode operasi: **Web Dashboard** (UI) dan **CLI** (Command Line).

### 0. `ui` - 🖥️ Membuka Web Dashboard (Rekomendasi Utama Operasional)

Setelah Anda (atau DevOps Anda) berhasil melakukan kompilasi (_build_) aplikasi ini menjadi file *executable*, cara paling efisien untuk mengoperasikan seluruh fitur alat ini adalah melalui server UI terintegrasinya. Cukup jalankan perintah berikut (ganti `<nama-binary-anda>` dengan nama hasil _build_ Anda):

```bash
./<nama-binary-anda> ui
```

*Atau, jika Anda baru saja melakukan `git clone` dan ingin langsung menjajal aplikasi web ini tanpa melakukan kompilasi (_build_) aplikasi Go sama sekali, Anda dapat mengeksekusi _source code_-nya secara langsung.* 

*Satu hal yang perlu diperhatikan, karena aplikasi ini menggunakan teknologi `go:embed` untuk menyatukan _frontend_ dan _backend_, Anda tetap harus menyiapkan aset webnya terlebih dahulu sebelum Golang bisa menyala:*

```bash
# 1. Siapkan aset file Svelte dan Tailwind
cd frontend && npm install && npm run build && cd ..

# 2. Nyalakan server utama secara langsung
go run main.go ui
```

**Langkah-langkah Penggunaan Lengkap via Web:**

1. Buka aplikasi Terminal / Command Prompt Anda.
2. Ketikkan perintah di atas, lalu tekan `Enter`.
3. Akan muncul pengumuman bahwa server siap:
   ```text
   ╔══════════════════════════════════════════════╗
   ║     🕵️  Web QA Automation Dashboard          ║
   ╠══════════════════════════════════════════════╣
   ║  🌐 Open: http://localhost:8080              ║
   ║  ⏹  Stop: Press Ctrl+C                      ║
   ╚══════════════════════════════════════════════╝
   ```
4. Buka Browser Anda (Chrome / Firefox / Safari).
5. Ketik secara manual alamat ini di *address bar* Anda: **`http://localhost:8080`**
6. Selamat! Panel Web QA Automation yang mewah dengan *Dark Theme* kini ada di depan Anda.
7. Di sebelah kiri (Sidebar), Anda bisa mengklik ganti-ganti fitur: 
   - **Scan**: Untuk Mode Audit Keamanan Web / Fuzzing.
   - **Baseline & Compare**: Untuk men-*testing* regresi desain Visual (perubahan jarak tombol, beda warna akibat _update_ web).
   - **A11y**: Mengecek Skor Aksesibilitas WCAG standard Disabilitas.
   - **Load Test**: Untuk menguji beban tembakan ribuan klik per detik.
8. Tinggal isi kolom **Target URL** (cth: `https://shopee.co.id`), atur opsi lewat kursor mouse Anda, lalu klik **Mulai**.
9. **Cara Mematikan:** Jika sudah puas menggunakannya, kembali ke Terminal hitam Anda tadi, lalu tekan gabungan tombol `CTRL + C` di *keyboard* untuk mematikan server.

*(Catatan Lanjut: Jika kebetulan Port 8080 di laptop Anda error atau dipakai aplikasi lain, Anda bebas memindahkannya dengan: `go run main.go ui --port 9090`)*

---

### Mode CLI Terdedikasi (Untuk Integrasi Skrip & CI/CD)

> **Catatan:** Apabila Anda sekadar ingin menjalankan *testing* manual, silakan gunakan mode Web Dashboard `ui` di atas. Panduan _flag_ CLI di bawah ini didedikasikan bagi Anda yang berniat mengintegrasikan _binary_ ini ke dalam lingkungan otomatis (seperti Gitlab-CI, GitHub Actions, atau *Cronjob* Server).

Berikut adalah perintah-perintah CLI yang masih beroperasi murni secara *headless* tanpa memicu UI Server:

### 1. `scan` - The Functional Security Audit

Ini adalah mode palu godam. Bot akan merayap ke dalam, merekam anomali jaringan, melakukan _Fuzzing_ paksa, merekam video (jika diminta), dan menangkap letak halaman yang rusak.

**Contoh Cepat (Fast Mode):**

```bash
go run main.go scan --target="https://example.com" --depth=2 --fast-mode
```

**Contoh Audit Dashboard (Menggunakan Login State):**

```bash
go run main.go scan --target="https://example.com/admin" --depth=3 --auth-json="./session.json"
```

### 2. `baseline` - The Pristine Snapshot

Perintah ini akan menyuruh bot berjalan-jalan dengan Sangat Sopan (Tanpa meretas/klik paksa). Ia akan menunggu sampai animasi halaman web selesai lalu diam-diam merekam _screenshot layout_ yang sempurna (_Baseline_).

```bash
go run main.go baseline --target="https://example.com" --depth=2
```

_Barang bukti foto murni akan secara otomatis dikelompokkan ke folder harian: `proofs/<DD-MM-YYYY>/baseline/`_

### 3. `compare` - The Visual Regression Engine

Kapanpun tim _developer_ Anda mencurigai perubahan kode `CSS/Frontend` merusak tampilan, jalankan kode ini! Ia akan mencocokkan kondisi website hari ini dengan histori foto _Baseline_ Anda. Secara otomatis program akan membandingkannya dengan _baseline_ di hari yang sama, namun Anda dapat merujuk ke tanggal _baseline_ masa lalu melalui parameter `--baseline-date`.

```bash
go run main.go compare --target="https://example.com" --depth=2 --baseline-date="22-02-2026"
```

_Hasil perhitungan pergeseran visualnya dicetak ke `visual_regression_report.md` beserta foto kemerah-merahannya di folder hari ini: `proofs/<DD-MM-YYYY>/diff/`_.

### 4. `a11y` - The Accessibility Audit

Perintah ini akan berkeliling web Anda khusus untuk melakukan perhitungan matematis rasio kontras warna dan kelengkapan struktur HTML bagi pembaca layar Tuna Netra (WCAG Standards).

```bash
go run main.go a11y --target="https://example.com"
```

_Daftar elemen navigasi/tombol yang cacat dan melanggar standar disabilitas akan didokumentasikan di dalam `accessibility_audit_report.md`._

### 5. `load` - Stateful High-Concurrency Stress Test

Menguji kekuatan otot _Server API_ / Infrastruktur web Anda. Pisau bedah paling ampuh untuk mengetahui seberapa banyak _Request Per Second (RPS)_ yang mampu ditahan oleh Server sebelum web Anda melambat atau mati total (Error 500). Mendukung injeksi tag `{{RANDOM}}` untuk Payload tipe POST.

```bash
# Contoh 1: Pengetesan biasa (GET)
go run main.go load --target="https://example.com/api" --users=500 --duration=30s

# Contoh 2: Serangan Database murni dengan injeksi Dynamic JSON Cache-Busting (POST)
go run main.go load --target="https://example.com/api/register" --method="POST" --body-json='{"email": "hacked_{{RANDOM}}@mail.com"}' --users=200 --duration=15s
```

_Sistem akan menghujani target dengan ratusan Goroutines (Simulasi pengguna bersamaan). Tag `{{RANDOM}}` akan mencetak string acak per koneksi sehingga Cloudflare Cache tidak bisa menolong server target Anda. Hasil dicetak ke dalam `load_test_report.md`._

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

1. **`qa_audit_report.md`** : Buku Suci Laporan Utama. Terbagi atas ringkasan jumlah laman yang ditelusuri (![Crawl Map]), Log Kesalahan Jaringan Rinci (Daftar API yang ambruk dan mati), dan Bug-Bug Kritis. Khusus bug fungsionalitas, akan ada rincian tabel _Steps-To-Reproduce_ (Aksi tombol mana yang membuat aplikasi Anda rusak, dan peringatan _Broken Access Control / Token Tampering_).
2. **`network_anomalies.json`** : Basis data mentah berformat JSON bagi Anda yang ingin membuat skrip pipa otomatis (_CI/CD_) untuk sekadar membaca rekap kesalahan HTTP/POST.
3. **`visual_regression_report.md`** : Laporan kembaran. Ini HANYA dicetak bila perintah `compare` dipanggil. Memamerkan daftar persentase (%) kerusakan layout per halaman lengkap dengan tabel foto rontgen visual-nya.
4. **`accessibility_audit_report.md`**: Rekap medis kelayakan fungsi aksesibilitas UI/UX aplikasi Anda untuk penyandang disabilitas. Hadir jika menggunakan fungsi `a11y`.
5. **`load_test_report.md`**: Sertifikat kekuatan mesin server Anda (RPS & Latency Hit). Hadir jika memanggil perintah `load`.
6. **Folder `./proofs/<DD-MM-YYYY>/`** : Laci barang-bukti yang **sekarang otomatis dikelompokkan langsung berdasarkan tanggal sesi pelacakan** (Format _DD-MM-YYYY_), sehingga tangkapan layar web harian Anda tidak lagi menumpuk dan tercampur menjadi satu.
   - `proofs/<DD-MM-YYYY>/scan/crash_X.png`: Saksi bisu letak tombol pemicu kepanikan Java Script.
   - `proofs/<DD-MM-YYYY>/scan/view_X.png`: Sketsa muka web di setiap halaman.
   - `proofs/<DD-MM-YYYY>/baseline/` & `.../current/` & `.../diff/`: Repositori log operasi tata-letak (_Visual testing_).

---

## 🛡️ Disclaimer (Peringatan Penting Hukum)

Aplikasi ini dirakit dengan teknik manipulasi DOM tingkat tinggi dan algoritma _Form-Fuzzing_ agresif.

1. **Jangan PERNAH menjalankan instruksi `scan` di aplikasi Produksi/Server Utama tanpa sepengetahuan Atasan/Pemilik IT.** Active Fuzzing akan mengirim ratusan entri aneh (Emoji, karakter pembobol SQL) ke Form Kontak Terbuka, Database, maupun Registrasi Web. Ini bisa memicu alarm keamanan AWS/Cloudflare atau merusak (_corrupting_) Database asli.
2. Selalu arahkan atribut `--target` ke lingkungan _Sandbox_, _Localhost_, atau _Staging Server_ demi keselamatan Anda.
3. Modul Visual (`baseline` & `compare`) bersifat mode _Read-Only_ dan **100% aman** untuk dipergunakan pada situs yang tayang (Live/Produksi).
4. **Web Dashboard (`ui`)** secara _default_ hanya bisa diakses dari _localhost_ (komputer lokal Anda). Jangan mengekspos port dashboard ini ke internet publik tanpa memasang sistem _authentication_ tambahan.

---

**Dibuat untuk merevolusi Dunia Quality Assurance secara Otomatis.**
