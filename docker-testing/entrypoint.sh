#!/bin/sh
set -e

# ── defaults ──────────────────────────────────────────────────────────────────
APP_ENV_DEFAULT="dev"
LOG_LEVEL_DEFAULT="info"
DB_HOST_DEFAULT="postgres"
DB_PORT_DEFAULT="5432"
DB_USER_DEFAULT="postgres"
DB_PASSWORD_DEFAULT="postgres"
DB_NAME_DEFAULT="verify_auth"
DB_SSLMODE_DEFAULT="disable"
JWT_ACCESS_TTL_DEFAULT="15m"
JWT_REFRESH_TTL_DEFAULT="168h"

APP_ENV="${APP_ENV:-$APP_ENV_DEFAULT}"
LOG_LEVEL="${LOG_LEVEL:-$LOG_LEVEL_DEFAULT}"
TOKEN_WHATSAPP="${TOKEN_WHATSAPP:-}"
BASE_URL_GRAPH_API="${BASE_URL_GRAPH_API:-}"
PHONE_NUMBER_ID="${PHONE_NUMBER_ID:-}"
SMTP_PORT="${SMTP_PORT:-}"
SMTP_USER="${SMTP_USER:-}"
SMTP_PASS="${SMTP_PASS:-}"
SMTP_HOST="${SMTP_HOST:-}"
DB_HOST="${DB_HOST:-$DB_HOST_DEFAULT}"
DB_PORT="${DB_PORT:-$DB_PORT_DEFAULT}"
DB_USER="${DB_USER:-$DB_USER_DEFAULT}"
DB_PASSWORD="${DB_PASSWORD:-$DB_PASSWORD_DEFAULT}"
DB_NAME="${DB_NAME:-$DB_NAME_DEFAULT}"
DB_SSLMODE="${DB_SSLMODE:-$DB_SSLMODE_DEFAULT}"
JWT_ACCESS_SECRET="${JWT_ACCESS_SECRET:-}"
JWT_REFRESH_SECRET="${JWT_REFRESH_SECRET:-}"
JWT_ACCESS_TTL="${JWT_ACCESS_TTL:-$JWT_ACCESS_TTL_DEFAULT}"
JWT_REFRESH_TTL="${JWT_REFRESH_TTL:-$JWT_REFRESH_TTL_DEFAULT}"

cat > /app/.env << EOF
APP_ENV=${APP_ENV}
LOG_LEVEL=${LOG_LEVEL}
TOKEN_WHATSAPP=${TOKEN_WHATSAPP}
BASE_URL_GRAPH_API=${BASE_URL_GRAPH_API}
PHONE_NUMBER_ID=${PHONE_NUMBER_ID}
SMTP_PORT=${SMTP_PORT}
SMTP_USER=${SMTP_USER}
SMTP_PASS=${SMTP_PASS}
SMTP_HOST=${SMTP_HOST}
DB_HOST=${DB_HOST}
DB_PORT=${DB_PORT}
DB_USER=${DB_USER}
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=${DB_NAME}
DB_SSLMODE=${DB_SSLMODE}
JWT_ACCESS_SECRET=${JWT_ACCESS_SECRET}
JWT_REFRESH_SECRET=${JWT_REFRESH_SECRET}
JWT_ACCESS_TTL=${JWT_ACCESS_TTL}
JWT_REFRESH_TTL=${JWT_REFRESH_TTL}
EOF

# ── info banner ───────────────────────────────────────────────────────────────
mask() {
    val="$1"
    if [ -z "$val" ]; then
        echo "(empty)"
    elif [ "${#val}" -le 6 ]; then
        echo "***"
    else
        head="$(echo "$val" | cut -c1-3)"
        tail="$(echo "$val" | rev | cut -c1-3 | rev)"
        echo "${head}***${tail}"
    fi
}

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║            DOCKER TESTING CONFIGURATION                 ║"
echo "╠══════════════════════════════════════════════════════════╣"
echo "║  App Env       : $(printf '%-37s' "${APP_ENV}")║"
echo "║  Log Level     : $(printf '%-37s' "${LOG_LEVEL}")║"
echo "║  API Port      : $(printf '%-37s' "8080")║"
echo "╠══════════════════════════════════════════════════════════╣"
echo "║  DB Host       : $(printf '%-37s' "${DB_HOST}")║"
echo "║  DB Port       : $(printf '%-37s' "${DB_PORT}")║"
echo "║  DB User       : $(printf '%-37s' "${DB_USER}")║"
echo "║  DB Password   : $(printf '%-37s' "$(mask "${DB_PASSWORD}")")║"
echo "║  DB Name       : $(printf '%-37s' "${DB_NAME}")║"
echo "║  DB SSL Mode   : $(printf '%-37s' "${DB_SSLMODE}")║"
echo "╠══════════════════════════════════════════════════════════╣"
echo "║  JWT Access TTL : $(printf '%-35s' "${JWT_ACCESS_TTL}")║"
echo "║  JWT Refresh TTL: $(printf '%-35s' "${JWT_REFRESH_TTL}")║"
echo "║  JWT Access Sec : $(printf '%-35s' "$(mask "${JWT_ACCESS_SECRET}")")║"
echo "║  JWT Refresh Sec: $(printf '%-35s' "$(mask "${JWT_REFRESH_SECRET}")")║"
echo "╠══════════════════════════════════════════════════════════╣"
echo "║  WhatsApp Token: $(printf '%-35s' "$(mask "${TOKEN_WHATSAPP}")")║"
echo "║  Graph API URL : $(printf '%-35s' "${BASE_URL_GRAPH_API:-not set}")║"
echo "║  Phone Number  : $(printf '%-35s' "${PHONE_NUMBER_ID:-not set}")║"
echo "╠══════════════════════════════════════════════════════════╣"
echo "║  SMTP Host     : $(printf '%-37s' "${SMTP_HOST:-not set}")║"
echo "║  SMTP Port     : $(printf '%-37s' "${SMTP_PORT:-not set}")║"
echo "║  SMTP User     : $(printf '%-37s' "${SMTP_USER:-not set}")║"
echo "║  SMTP Pass     : $(printf '%-37s' "$(mask "${SMTP_PASS}")")║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

exec "$@"
