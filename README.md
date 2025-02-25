## Yapılandırma dosyasında her iş için şu ayarları desteklemelisin:

- İşin başlatılacağı komut (command).
- Kaç süreç çalıştırılacağı (numprocs).
- Programın başlatıldığında otomatik başlayıp başlamayacağı (autostart).
- Yeniden başlatma politikası (autorestart: always, never, unexpected).
- Beklenen çıkış kodları (exitcodes).
- Başarılı başlangıç için gereken minimum çalışma süresi (startsecs).
- Yeniden başlatma deneme sayısı (startretries).
- Durdurma sinyali (stopsignal, örneğin SIGTERM).
- Zarif durdurma sonrası öldürme bekleme süresi (stoptime).
- Çıktıları dosyaya yönlendirme veya yok sayma (stdout, stderr).
- Ortam değişkenleri (env).
- Çalışma dizini (workingdir).
- Umask ayarı (umask).
