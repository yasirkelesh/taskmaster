# Taskmaster

**Taskmaster**, Unix ve Unix benzeri işletim sistemleri için geliştirilmiş bir iş kontrol daemon'u ve süreç yöneticisidir. Supervisor programından ilham alınarak oluşturulmuştur ve alt süreçlerin yaşam döngüsünü yönetmek için güçlü bir araçtır.

## 🎯 Amaç

Taskmaster, supervisor'a benzer özelliklere sahip tam teşekküllü bir iş kontrol daemon'u oluşturmayı amaçlar. Ana görevleri:

- Alt süreçleri başlatmak ve canlı tutmak
- Süreç durumlarını izlemek
- Gerektiğinde süreçleri yeniden başlatmak
- Yapılandırma dosyası üzerinden esnek kontrol sağlamak

## 🌟 Özellikler

### İş Kontrolü
- Süreç gruplarının askıya alınması
- Süreç devam ettirme işlemleri
- Süreç sonlandırma yönetimi
- Sinyal gönderme desteği

### Süreç Yönetimi
- Alt süreçleri başlatma
- Süreç durumlarını izleme (canlı/ölü)
- Otomatik yeniden başlatma
- Yapılandırılabilir yeniden başlatma stratejileri

### Yapılandırma Yönetimi
- YAML formatında yapılandırma dosyası
- Başlangıçta otomatik yükleme
- SIGHUP ile dinamik yapılandırma yeniden yükleme
- Hot-reload: Değiştirilmemiş süreçleri yeniden başlatmaz

### Günlükleme
- Yerel dosyaya olay kaydı
- Başlatma, durdurma, yeniden başlatma olayları
- Beklenmedik çıkış kayıtları
- Yapılandırma değişikliği logları

### Kontrol Kabuğu
- Interaktif kontrol arayüzü
- Satır düzenleme, geçmiş ve tamamlama desteği
- Supervisorctl benzeri komutlar
- Gerçek zamanlı süreç yönetimi

## 📋 Yapılandırma

### Yapılandırma Dosyası Formatı

Yapılandırma dosyası YAML formatında olmalıdır ve her program için aşağıdaki parametreleri içermelidir:

```yaml
programs:
  example_program:
    command: "/usr/bin/my_program --option value"
    numprocs: 1
    autostart: true
    autorestart: unexpected
    exitcodes: [0, 2]
    startretries: 3
    startsecs: 1
    stopsignal: TERM
    stopwaitsecs: 10
    stdout_logfile: "/var/log/example_program.log"
    stderr_logfile: "/var/log/example_program_error.log"
    environment:
      PATH: "/usr/local/bin:/usr/bin:/bin"
      PYTHONPATH: "/opt/myapp"
    directory: "/opt/myapp"
    umask: 022
```

### Yapılandırma Parametreleri

| Parametre | Açıklama |
|-----------|----------|
| `command` | Programı başlatmak için kullanılacak komut |
| `numprocs` | Başlatılacak ve çalışır durumda tutulacak işlem sayısı |
| `autostart` | Programın başlangıçta otomatik başlatılıp başlatılmayacağı |
| `autorestart` | Yeniden başlatma stratejisi (always/never/unexpected) |
| `exitcodes` | Beklenen çıkış kodları |
| `startsecs` | Başarılı başlatma için minimum çalışma süresi |
| `startretries` | Yeniden başlatma deneme sayısı |
| `stopsignal` | Durdurmak için kullanılacak sinyal |
| `stopwaitsecs` | Zararsız durdurmadan sonra bekleme süresi |
| `stdout_logfile` | Standart çıktı log dosyası |
| `stderr_logfile` | Hata çıktı log dosyası |
| `environment` | Ortam değişkenleri |
| `directory` | Çalışma dizini |
| `umask` | Dosya oluşturma izinleri |

## 🚀 Kullanım

### Başlatma

```bash
./taskmaster config.yaml
```

### Kontrol Kabuğu Komutları

Program başlatıldığında interaktif bir kabuk açılır:

```
taskmaster> status
program_name                 RUNNING   pid 1234, uptime 0:05:23

taskmaster> start program_name
program_name: started

taskmaster> stop program_name
program_name: stopped

taskmaster> restart program_name
program_name: stopped
program_name: started

taskmaster> reload
Reloaded configuration

taskmaster> quit
Shutting down taskmaster...
```

### Temel Komutlar

- `status` - Tüm programların durumunu gösterir
- `start <program>` - Programı başlatır
- `stop <program>` - Programı durdurur
- `restart <program>` - Programı yeniden başlatır
- `reload` - Yapılandırma dosyasını yeniden yükler
- `quit` - Taskmaster'ı sonlandırır

## 🔧 Kurulum

### Sistem Gereksinimleri

- Unix/Linux işletim sistemi
- Python 3.8+ (örnek uygulama için)
- YAML yapılandırma dosyası

### Kurulum Adımları

1. Projeyi klonlayın:
   ```bash
   git clone <repository-url>
   cd taskmaster
   ```

2. Bağımlılıkları yükleyin:
   ```bash
   pip install -r requirements.txt
   ```

3. Yapılandırma dosyasını hazırlayın:
   ```bash
   cp config.yaml.example config.yaml
   # config.yaml dosyasını düzenleyin
   ```

4. Taskmaster'ı başlatın:
   ```bash
   ./taskmaster config.yaml
   ```

## 🔒 Güvenlik

- Program root olarak çalışmaz
- Daemon olarak çalışması gerekmez
- Sanal makine ortamında güvenli çalışma
- Ayrıcalık azaltma desteği (bonus özellik)

## 📊 Günlükleme

Taskmaster aşağıdaki olayları günlükler:

- Program başlatma/durdurma olayları
- Yeniden başlatma denemeleri
- Beklenmedik çıkışlar
- Yapılandırma değişiklikleri
- Hata durumları

Günlük dosyası varsayılan olarak `/var/log/taskmaster.log` konumunda oluşturulur.

## 🎁 Bonus Özellikler

### Mevcut Bonus Özellikler

- [ ] Başlatma sırasında ayrıcalık azaltma
- [ ] İstemci/sunucu mimarisi
- [ ] Gelişmiş günlükleme (e-posta/http/syslog)
- [ ] Süreç konsola "ekleme" özelliği (tmux/screen benzeri)

## 🧪 Test Senaryoları

Savunma oturumu için aşağıdaki senaryolar test edilmelidir:

1. **Normal Çalışma**: Programların başlatılması ve çalışması
2. **Manuel Öldürme**: Süreçlerin manuel olarak öldürülmesi
3. **Başarısız Başlatma**: Hatalı komutlarla başlatma denemeleri
4. **Çok Çıktı**: Yoğun çıktı üreten programlar
5. **Yapılandırma Değişiklikleri**: Hot-reload testleri
6. **Sinyal Yönetimi**: SIGHUP ve diğer sinyaller

## 📝 Geliştirme

### Proje Yapısı

```
taskmaster/
├── src/
│   ├── main.py
│   ├── config.py
│   ├── process_manager.py
│   ├── shell.py
│   └── logger.py
├── tests/
├── config.yaml.example
├── README.md
└── requirements.txt
```

### Katkıda Bulunma

1. Fork edin
2. Feature branch oluşturun
3. Testlerinizi yazın
4. Pull request açın

## 📄 Lisans

Bu proje MIT lisansı altında lisanslanmıştır.

## 🔍 Sorun Giderme

### Yaygın Sorunlar

1. **Yapılandırma Hatası**: YAML formatını kontrol edin
2. **İzin Hatası**: Dosya izinlerini kontrol edin
3. **Port Kullanımda**: Başka bir taskmaster instance'ı çalışıyor olabilir

### Hata Ayıklama

```bash
# Detaylı günlükleme ile başlatma
./taskmaster --debug config.yaml

# Yapılandırma dosyasını doğrulama
./taskmaster --check-config config.yaml
```

## 📚 Referanslar

- [Supervisor Documentation](http://supervisord.org/)
- [Unix Process Management](https://en.wikipedia.org/wiki/Process_management_(computing))
- [YAML Specification](https://yaml.org/spec/)

---

**Not**: Bu proje eğitim amaçlı geliştirilmiştir ve production ortamında kullanılmadan önce kapsamlı testlerden geçirilmelidir.