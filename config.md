# --- ИДЕАЛЬНЫЙ СЕРВЕР (MAKE_ZIPER / FOR_MAK_GPT) ---
# GitHub: git@github.com:Vovarama1992/for_mak_gpt.git

# --- ОСНОВА ---
Ubuntu 22.04 LTS
nginx + certbot + docker + docker-compose
всё обновлено: apt update && apt upgrade -y

# --- ФАЙЛЫ ---
/var/www/back/          # код бэкенда (git clone)
/var/www/back/.env      # переменные окружения (S3, DB, PORT и т.п.)
/var/www/back/docker-compose.yml
/etc/nginx/sites-enabled/for_mak_gpt
/etc/systemd/system/makeziper.service (если без docker)

/var/log/nginx/
 /var/log/makeziper/    # логи приложения (если нужны отдельно)

# --- NGINX ---
server {
    listen 80;
    server_name makeziper.ru www.makeziper.ru;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name makeziper.ru www.makeziper.ru;

    ssl_certificate     /etc/letsencrypt/live/makeziper.ru/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/makeziper.ru/privkey.pem;

    client_max_body_size 25m;

    # только API
    location / {
        proxy_pass         http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header   Upgrade $http_upgrade;
        proxy_set_header   Connection "upgrade";
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_read_timeout 300s;
    }
}
