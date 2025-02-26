# Taskmaster

Taskmaster, Unix benzeri sistemlerde iş kontrolü (job control) sağlamak için geliştirilmiş bir araçtır. Supervisor'a benzer bir işlevsellik sunar ve süreçleri başlatma, durdurma, yeniden başlatma ve izleme gibi görevleri yerine getirir. Bu proje, yapılandırma dosyası üzerinden süreçleri yönetir ve kullanıcıya bir kontrol kabuğu (shell) sağlar.

Bu proje, Go dili ile yazılmıştır ve bir sanal makine üzerinde çalışacak şekilde tasarlanmıştır.

## Özellikler
- **Süreç Yönetimi**: Yapılandırma dosyasında tanımlı süreçleri başlatır, izler ve gerektiğinde yeniden başlatır.
- **Kontrol Kabuğu**: Kullanıcıya süreçleri yönetmek için basit bir komut satırı arayüzü sunar.
- **Yapılandırma Yenileme**: Çalışırken yapılandırma dosyasını yeniden yükleme (`reload`) desteği.
- **Loglama**: Olayları (başlatma, durdurma, yeniden başlatma vb.) bir log dosyasına kaydeder.
- **Otomatik Yeniden Başlatma**: `autorestart` politikaları (`always`, `never`, `unexpected`) ile süreçlerin otomatik yönetimi.

### Desteklenen Komutlar
- `status`: Tüm süreçlerin durumunu gösterir.
- `start <program_adı>`: Belirtilen programı başlatır.
- `reload`: Yapılandırma dosyasını yeniden yükler ve süreçleri günceller.
- `exit`: Taskmaster'ı kapatır.

**Not:** `stop` ve `restart` komutları henüz uygulanmadı (TODO).

## Kurulum

### Gereksinimler
- Go 1.21 veya üstü
- YAML kütüphanesi: `gopkg.in/yaml.v2`
- Sanal makine (örneğin VirtualBox veya Vagrant ile Ubuntu)

### Adımlar
1. **Depoyu Klonla**
   ```bash
   git clone <repository_url>
   cd taskmaster
2. **Bağımlılıkları Yükle**
    ```bash
    go mod init taskmaster
    go get gopkg.in/yaml.v2
3. **Yapılandırma Dosyasını Hazırla** config/config.yaml dosyasını oluşturun ve süreçlerinizi tanımlayın. Örnek:

    ```bash
    programs:
   nginx:
    command: "echo 'Nginx running' && sleep 10"
    numprocs: 1
    autostart: true
    autorestart: never
    exitcodes: [0]
    startsecs: 5
    startretries: 3
    stopsignal: TERM
    stoptime: 10
    stdout: "/tmp/nginx.stdout"
    stderr: "/tmp/nginx.stderr"
    env:
      STARTED_BY: "taskmaster"
    workingdir: "/tmp"
    umask: 022

### Örnek Kullanım
    
    ```bash
    taskmaster> status
    Program         Status
    ---------------------
    nginx[0]        running
    taskmaster> start nginx
    Hata: 'nginx' zaten maksimum süreç sayısında çalışıyor
    taskmaster> reload
    Yapılandırma başarıyla yenilendi.
    taskmaster> exit

### Yapılandırma Dosyası

- command: Çalıştırılacak komut.
- numprocs: Kaç süreç çalıştırılacağı.
- autostart: Programın otomatik başlayıp başlamayacağı.
- autorestart: Yeniden başlatma politikası (always, never,  unexpected).
- exitcodes: Beklenen çıkış kodları.
- startsecs: Başarılı başlangıç için minimum çalışma süresi.
- startretries: Yeniden başlatma deneme sayısı.
- stopsignal: Durdurma sinyali (örneğin TERM).
- stoptime: Zarif durdurma sonrası bekleme süresi.
- stdout/stderr: Çıktıların yönlendirileceği dosyalar.
- env: Ortam değişkenleri.
- workingdir: Çalışma dizini.
- umask: Dosya izin maskesi.


### Geliştirme Durumu
Bu proje şu anda temel işlevselliği destekliyor, ancak bazı özellikler tamamlanmadı:

- [x] status komutu
- [x] start komutu
- [x] reload komutu
- [x] stop komutu
- [x] restart komutu
- [ ] startretries ile yeniden başlatma sınırı
- [ ] Gelişmiş loglama entegrasyonu