FROM postgres:16

WORKDIR /app

# Install cron
RUN apt-get update && apt-get -y install cron

COPY backup.sh /app/backup.sh
COPY restore.sh /app/restore.sh

COPY cron /app/cron

COPY .env.db /app/.env.db

RUN chmod +x /app/backup.sh
RUN chmod +x /app/restore.sh

# Apply the cron job
RUN chmod +x /app/cron
RUN crontab /app/cron
